package main

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestNewVersionAndString(t *testing.T) {
	cases := []struct {
		name                string
		major, minor, patch uint64
		pre                 string
		want                string
	}{
		{name: "zero", want: "0.0.0"},
		{name: "all numeric", major: 1, minor: 2, patch: 3, want: "1.2.3"},
		{name: "with pre", major: 1, minor: 2, patch: 3, pre: "rc1", want: "1.2.3-rc1"},
		{name: "dotted pre", major: 1, minor: 2, patch: 3, pre: "rc.1", want: "1.2.3-rc.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := newVersion(tc.major, tc.minor, tc.patch, tc.pre).String()
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	// Each row asserts left vs right yields want (-1/0/1) and the reverse
	// yields the opposite.
	cases := []struct {
		name        string
		left, right version
		want        int
	}{
		{name: "equal", left: newVersion(1, 2, 3, ""), right: newVersion(1, 2, 3, ""), want: 0},
		{name: "patch", left: newVersion(1, 2, 3, ""), right: newVersion(1, 2, 4, ""), want: -1},
		{name: "minor", left: newVersion(1, 2, 3, ""), right: newVersion(1, 3, 0, ""), want: -1},
		{name: "major", left: newVersion(1, 2, 3, ""), right: newVersion(2, 0, 0, ""), want: -1},
		{name: "release greater than pre", left: newVersion(1, 2, 3, ""), right: newVersion(1, 2, 3, "rc1"), want: 1},
		{name: "pre lex within identifier", left: newVersion(1, 2, 3, "rc1"), right: newVersion(1, 2, 3, "rc2"), want: -1},
		{name: "pre numeric within ident path", left: newVersion(1, 2, 3, "rc.1"), right: newVersion(1, 2, 3, "rc.2"), want: -1},
		{name: "pre alpha vs beta", left: newVersion(1, 2, 3, "alpha"), right: newVersion(1, 2, 3, "beta"), want: -1},
		{name: "pre numeric vs numeric", left: newVersion(1, 2, 3, "1"), right: newVersion(1, 2, 3, "2"), want: -1},
		{name: "pre numeric ranks below alpha", left: newVersion(1, 2, 3, "1"), right: newVersion(1, 2, 3, "alpha"), want: -1},
		{name: "longer pre wins on tie", left: newVersion(1, 2, 3, "alpha"), right: newVersion(1, 2, 3, "alpha.1"), want: -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.left.Compare(tc.right); got != tc.want {
				t.Fatalf("Compare: want %d, got %d", tc.want, got)
			}
			if rev := tc.right.Compare(tc.left); rev != -tc.want {
				t.Fatalf("reverse Compare: want %d, got %d", -tc.want, rev)
			}
		})
	}
}

func TestParseConstraintErrors(t *testing.T) {
	cases := []string{
		"",
		"   ",
		"||",
		"1.0.0 ||",
		"1.invalid",
		"^",
		"!!1.2.3",
		">> 1.2.3",
		"1.2.3-",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := parseConstraint(in); err == nil {
				t.Fatalf("want error for %q, got nil", in)
			}
		})
	}
}

func TestParseConstraintOK(t *testing.T) {
	cases := []string{
		"1.2.3",
		"=1.2.3",
		"==1.2.3",
		"!=1.2.3",
		">1.2.3",
		">=1.2.3",
		"<1.2.3",
		"<=1.2.3",
		"^1.2.3",
		"~1.2.3",
		"*",
		"x",
		"X",
		"*-0",
		"1",
		"1.2",
		"1.x",
		"1.2.x",
		">=1.2.3 <2.0.0",
		">= 1.2.3 < 2.0.0",
		"1.2.3, 1.2.4",
		"1.2.3 || 1.2.4",
		">=1.0.0-rc.1 <2.0.0-0",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			if _, err := parseConstraint(in); err != nil {
				t.Fatalf("unexpected error for %q: %v", in, err)
			}
		})
	}
}

// TestAllows is the central correctness gate. Each row asserts that
// `ver` satisfies `c` iff want=true.
func TestAllows(t *testing.T) {
	cases := []struct {
		c    string
		ver  version
		want bool
	}{
		// Exact equality.
		{"1.2.3", newVersion(1, 2, 3, ""), true},
		{"1.2.3", newVersion(1, 2, 4, ""), false},
		{"1.2.3", newVersion(1, 2, 3, "rc1"), false}, // pre-release excluded
		{"=1.2.3", newVersion(1, 2, 3, ""), true},
		{"==1.2.3", newVersion(1, 2, 3, ""), true},

		// Wildcards.
		{"*", newVersion(1, 2, 3, ""), true},
		{"*", newVersion(0, 0, 0, ""), true},
		{"*", newVersion(1, 2, 3, "rc1"), false}, // pre-release excluded by *
		{"x", newVersion(9, 9, 9, ""), true},
		{"X", newVersion(0, 1, 0, ""), true},
		{"*-0", newVersion(1, 2, 3, ""), true},
		{"*-0", newVersion(1, 2, 3, "rc1"), true}, // pre-release allowed by -0
		{"*-0", newVersion(0, 0, 0, "alpha"), true},

		// Bare minor / major (tilde-like).
		{"1", newVersion(1, 0, 0, ""), true},
		{"1", newVersion(1, 99, 99, ""), true},
		{"1", newVersion(2, 0, 0, ""), false},
		{"1", newVersion(0, 99, 99, ""), false},
		{"1.2", newVersion(1, 2, 0, ""), true},
		{"1.2", newVersion(1, 2, 99, ""), true},
		{"1.2", newVersion(1, 3, 0, ""), false},
		{"1.2", newVersion(1, 1, 99, ""), false},
		{"1.x", newVersion(1, 5, 0, ""), true},
		{"1.x", newVersion(2, 0, 0, ""), false},
		{"1.2.x", newVersion(1, 2, 5, ""), true},
		{"1.2.x", newVersion(1, 3, 0, ""), false},

		// Inequality.
		{"!=1.2.3", newVersion(1, 2, 4, ""), true},
		{"!=1.2.3", newVersion(1, 2, 3, ""), false},

		// Greater / less.
		{">1.2.3", newVersion(1, 2, 4, ""), true},
		{">1.2.3", newVersion(1, 2, 3, ""), false},
		{">1.2", newVersion(1, 3, 0, ""), true},
		{">1.2", newVersion(1, 2, 99, ""), false}, // strict above the whole 1.2.x range
		{">1", newVersion(2, 0, 0, ""), true},
		{">1", newVersion(1, 99, 99, ""), false},
		{">=1.2.3", newVersion(1, 2, 3, ""), true},
		{">=1.2.3", newVersion(1, 2, 2, ""), false},
		{"<1.2.3", newVersion(1, 2, 2, ""), true},
		{"<1.2.3", newVersion(1, 2, 3, ""), false},
		{"<1.2", newVersion(1, 1, 99, ""), true},
		{"<1.2", newVersion(1, 2, 0, ""), false},
		{"<=1.2.3", newVersion(1, 2, 3, ""), true},
		{"<=1.2.3", newVersion(1, 2, 4, ""), false},
		{"<=1.2", newVersion(1, 2, 99, ""), true}, // entire 1.2.x range matches
		{"<=1.2", newVersion(1, 3, 0, ""), false},

		// Tilde.
		{"~1.2.3", newVersion(1, 2, 3, ""), true},
		{"~1.2.3", newVersion(1, 2, 99, ""), true},
		{"~1.2.3", newVersion(1, 3, 0, ""), false},
		{"~1.2.3", newVersion(1, 2, 2, ""), false},
		{"~1.2", newVersion(1, 2, 99, ""), true},
		{"~1.2", newVersion(1, 3, 0, ""), false},
		{"~1", newVersion(1, 99, 99, ""), true},
		{"~1", newVersion(2, 0, 0, ""), false},

		// Caret with major >= 1.
		{"^1.2.3", newVersion(1, 2, 3, ""), true},
		{"^1.2.3", newVersion(1, 99, 99, ""), true},
		{"^1.2.3", newVersion(2, 0, 0, ""), false},
		{"^1.2.3", newVersion(1, 2, 2, ""), false},
		{"^1", newVersion(1, 99, 99, ""), true},
		{"^1", newVersion(2, 0, 0, ""), false},
		{"^1.15.999", newVersion(1, 15, 7, ""), false},
		{"^1.15.999", newVersion(1, 16, 0, ""), true},

		// Caret with major == 0.
		{"^0.2.3", newVersion(0, 2, 3, ""), true},
		{"^0.2.3", newVersion(0, 2, 99, ""), true},
		{"^0.2.3", newVersion(0, 3, 0, ""), false},
		{"^0.0.3", newVersion(0, 0, 3, ""), true},
		{"^0.0.3", newVersion(0, 0, 4, ""), false},
		{"^0", newVersion(0, 99, 99, ""), true},
		{"^0", newVersion(1, 0, 0, ""), false},
		{"^0.0", newVersion(0, 0, 99, ""), true},
		{"^0.0", newVersion(0, 1, 0, ""), false},

		// AND.
		{">=1.2.3 <2.0.0", newVersion(1, 5, 0, ""), true},
		{">=1.2.3 <2.0.0", newVersion(1, 2, 2, ""), false},
		{">=1.2.3 <2.0.0", newVersion(2, 0, 0, ""), false},
		{">=1.2.3, <2.0.0", newVersion(1, 5, 0, ""), true},

		// OR.
		{"1.2.3 || 1.2.4", newVersion(1, 2, 3, ""), true},
		{"1.2.3 || 1.2.4", newVersion(1, 2, 4, ""), true},
		{"1.2.3 || 1.2.4", newVersion(1, 2, 5, ""), false},
		{"^1 || ^2", newVersion(1, 5, 0, ""), true},
		{"^1 || ^2", newVersion(2, 5, 0, ""), true},
		{"^1 || ^2", newVersion(3, 0, 0, ""), false},

		// Pre-release atom unlocks pre-release candidates for that atom.
		{">=1.2.3-0", newVersion(1, 2, 3, "rc1"), true},
		{">=1.2.3", newVersion(1, 2, 3, "rc1"), false},
		// Within an AND, every atom must individually allow pre-releases.
		{">=1.0.0-0 <2.0.0", newVersion(1, 5, 0, "rc1"), false},
		{">=1.0.0-0 <2.0.0-0", newVersion(1, 5, 0, "rc1"), true},
	}
	for _, tc := range cases {
		name := tc.c + "_" + tc.ver.String()
		t.Run(name, func(t *testing.T) {
			c, err := parseConstraint(tc.c)
			if err != nil {
				t.Fatalf("parseConstraint(%q): %v", tc.c, err)
			}
			if got := c.allows(tc.ver); got != tc.want {
				t.Fatalf("%q.allows(%q): want %v, got %v", tc.c, tc.ver, tc.want, got)
			}
		})
	}
}

func TestInvalidConstraintMessage(t *testing.T) {
	_, err := parseConstraint("1.invalid")
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "invalid constraint") {
		t.Fatalf("want error to mention %q, got %q", "invalid constraint", err.Error())
	}
}

func Test_parseGoVersion(t *testing.T) {
	cases := []struct {
		input string
		want  string // empty = expect a "could not parse" error
	}{
		{input: "go1.15.2", want: "1.15.2"},
		{input: "1.15.2", want: "1.15.2"},
		{input: "go1.15", want: "1.15.0"},
		{input: "1.15rc1", want: "1.15.0-rc1"},
		{input: "go1.21rc1", want: "1.21.0-rc1"},
		{input: "1.16", want: "1.16.0"},
		{input: "g1.15"},
		{input: "go1", want: "1.0.0"},
		{input: "1", want: "1.0.0"},
		{input: " "},
		{input: " 1"},
	}
	for _, td := range cases {
		td := td
		t.Run(td.input, func(t *testing.T) {
			got, err := parseGoVersion(td.input)
			if td.want == "" {
				if err == nil {
					t.Fatalf("want error, got nil (parsed %q)", got)
				}
				wantErr := `could not parse version "` + td.input + `"`
				if err.Error() != wantErr {
					t.Fatalf("want error %q, got %q", wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != td.want {
				t.Fatalf("want %q, got %q", td.want, got.String())
			}
		})
	}

	overflowCases := []struct {
		name    string
		input   string
		wantSub string
	}{
		{name: "overflow major", input: "go18446744073709551616.2.3", wantSub: "could not parse major version"},
		{name: "overflow minor", input: "go1.18446744073709551616.3", wantSub: "could not parse minor version"},
		{name: "overflow patch", input: "go1.2.18446744073709551616", wantSub: "could not parse patch version"},
	}
	for _, td := range overflowCases {
		td := td
		t.Run(td.name, func(t *testing.T) {
			_, err := parseGoVersion(td.input)
			if err == nil {
				t.Fatalf("want error, got nil")
			}
			if !strings.Contains(err.Error(), td.wantSub) {
				t.Fatalf("want error containing %q, got %q", td.wantSub, err.Error())
			}
		})
	}
}

func Test_run(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		stdin    io.Reader
		wantOut  []string // exact stdout lines (post trim/split)
		wantExit int
		wantErr  string // expected stderr (post trim)
	}{
		{
			name:    "matches exact version",
			args:    []string{"-c", "1.2.3", "1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4"},
			wantOut: []string{"1.2.3"},
		},
		{
			name:    "outputs canonical top version",
			args:    []string{"-c", "1", "1.2", "1", "1.2.3", "1.2.4"},
			wantOut: []string{"1.2.4"},
		},
		{
			name:    "accepts go-style versions by default",
			args:    []string{"-c", "1.2.3", "go1.2.0", "go1.2.3rc1", "go1.2.3", "go1.2.4"},
			wantOut: []string{"1.2.3"},
		},
		{
			name:    "ignores invalid candidates",
			args:    []string{"-c", "1.2.3", "1.2.0", "1.2.3-rc1", "1.2.3", "1.2.4", "not-a-version"},
			wantOut: []string{"1.2.3"},
		},
		{
			name:    "no match prints nothing",
			args:    []string{"-c", "9.9.9", "1.2.0", "1.2.3"},
			wantOut: nil,
		},
		{
			name:    "accepts stdin via -",
			args:    []string{"-c", "1.2.3", "-"},
			stdin:   strings.NewReader("1.2.0\n1.2.3-rc1\n1.2.3\n1.2.4\n"),
			wantOut: []string{"1.2.3"},
		},
		{
			name:    "stdin mixed with positional",
			args:    []string{"-c", "1", "1.2.0", "-"},
			stdin:   strings.NewReader("1.2.4\n"),
			wantOut: []string{"1.2.4"},
		},
		{
			name:    "--orig prints original for top match",
			args:    []string{"--orig", "-c", "1.2.3", "go1.2.0", "go1.2.3rc1", "go1.2.3", "go1.2.4"},
			wantOut: []string{"go1.2.3"},
		},
		{
			// the *-0 trick used by normalize_go_version: match anything, including
			// prereleases, so a single input is normalized through to canonical form.
			name:    "normalize via *-0 constraint",
			args:    []string{"-c", "*-0", "1.21rc1"},
			wantOut: []string{"1.21.0-rc1"},
		},
		{
			name:    "normalize via *-0 constraint with bare minor",
			args:    []string{"-c", "*-0", "1.16"},
			wantOut: []string{"1.16.0"},
		},
		{
			name:     "missing constraint",
			args:     []string{"1.2.3"},
			wantExit: 2,
			wantErr:  "goversion-select: -c <constraint> is required",
		},
		{
			name:     "invalid constraint",
			args:     []string{"-c", "1.invalid", "1.2.3"},
			wantExit: 1,
			wantErr:  `goversion-select: invalid constraint: "1.invalid"`,
		},
	}
	for _, td := range cases {
		td := td
		t.Run(td.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run(td.args, td.stdin, &stdout, &stderr)
			if exit != td.wantExit {
				t.Fatalf("exit: want %d, got %d (stderr=%q)", td.wantExit, exit, stderr.String())
			}
			gotErr := strings.TrimSpace(stderr.String())
			if gotErr != td.wantErr {
				t.Fatalf("stderr: want %q, got %q", td.wantErr, gotErr)
			}
			gotOut := strings.TrimSpace(stdout.String())
			if len(td.wantOut) == 0 {
				if gotOut != "" {
					t.Fatalf("stdout: want empty, got %q", gotOut)
				}
				return
			}
			gotLines := strings.Split(gotOut, "\n")
			if !reflect.DeepEqual(gotLines, td.wantOut) {
				t.Fatalf("stdout: want %v, got %v", td.wantOut, gotLines)
			}
		})
	}
}

// Test_run_invalidFlag covers the flag.Parse failure path (exit 2). The flag
// package writes its own usage message to stderr; we just confirm the exit
// code so we don't lock the test to flag's exact wording.
func Test_run_invalidFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := run([]string{"--nope"}, nil, &stdout, &stderr)
	if exit != 2 {
		t.Fatalf("exit: want 2, got %d", exit)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty, got %q", stdout.String())
	}
}

// jsonStreamSample is a hand-rolled fixture in go.dev /dl/?mode=json
// shape — semver-descending, mixed pre-releases and patches, deliberately
// minimal per object. Tests cover both early matches (latest stable case)
// and deep matches (older minor cases). Two non-version fields are
// included on the first object to confirm v2's default ignore-unknown
// behavior actually skips them.
const jsonStreamSample = `[
  {"version":"go1.26.3","stable":true,"files":[]},
  {"version":"go1.26.2"},
  {"version":"go1.26.1"},
  {"version":"go1.26.0"},
  {"version":"go1.26rc3"},
  {"version":"go1.25.10"},
  {"version":"go1.25.9"},
  {"version":"go1.21.13"},
  {"version":"go1.21.0"},
  {"version":"go1.16.15"},
  {"version":"go1.16beta1"},
  {"version":"go1.10.8"}
]`

func Test_run_jsonInput(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		stdin    string
		wantOut  []string
		wantExit int
		wantErr  string // substring expected in stderr
	}{
		{
			name:    "match at index 0",
			args:    []string{"--json-input", "-c", "1.26.x"},
			stdin:   jsonStreamSample,
			wantOut: []string{"1.26.3"},
		},
		{
			name:    "match deep in stream",
			args:    []string{"--json-input", "-c", "1.10.x"},
			stdin:   jsonStreamSample,
			wantOut: []string{"1.10.8"},
		},
		{
			name:    "match mid stream",
			args:    []string{"--json-input", "-c", "1.21.x"},
			stdin:   jsonStreamSample,
			wantOut: []string{"1.21.13"},
		},
		{
			name:    "orig echoes upstream version",
			args:    []string{"--json-input", "--orig", "-c", "1.26.x"},
			stdin:   jsonStreamSample,
			wantOut: []string{"go1.26.3"},
		},
		{
			name:    "star matches highest",
			args:    []string{"--json-input", "--orig", "-c", "*"},
			stdin:   jsonStreamSample,
			wantOut: []string{"go1.26.3"},
		},
		{
			name:    "prerelease exact match",
			args:    []string{"--json-input", "--orig", "-c", "=1.26.0-rc3"},
			stdin:   jsonStreamSample,
			wantOut: []string{"go1.26rc3"},
		},
		{
			name:    "no match returns empty",
			args:    []string{"--json-input", "-c", "1.99.x"},
			stdin:   jsonStreamSample,
			wantOut: nil,
		},
		{
			name:    "empty array no match",
			args:    []string{"--json-input", "-c", "1.26.x"},
			stdin:   `[]`,
			wantOut: nil,
		},
		{
			name:    "skip unparseable version field",
			args:    []string{"--json-input", "-c", "1.26.x"},
			stdin:   `[{"version":"weekly.2020-01-01"},{"version":"go1.26.3"}]`,
			wantOut: []string{"1.26.3"},
		},
		{
			name:     "positional candidate rejected with --json-input",
			args:     []string{"--json-input", "-c", "1.26.x", "go1.26.3"},
			stdin:    jsonStreamSample,
			wantExit: 2,
			wantErr:  "does not accept positional candidates",
		},
		{
			name:     "missing version field is schema break",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    `[{"stable":true},{"version":"go1.26.3"}]`,
			wantExit: 1,
			wantErr:  `missing required "version" field`,
		},
		{
			name:     "null version is schema break",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    `[{"version":null},{"version":"go1.26.3"}]`,
			wantExit: 1,
			wantErr:  `missing required "version" field`,
		},
		{
			name:     "non-string version is decode error",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    `[{"version":123}]`,
			wantExit: 1,
			wantErr:  "--json-input",
		},
		{
			name:     "non-array root rejected",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    `{"version":"go1.26.3"}`,
			wantExit: 1,
			wantErr:  "expected JSON array",
		},
		{
			name:     "malformed JSON before any match",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    `[{"version":"go1.26.3"`,
			wantExit: 1,
			wantErr:  "--json-input",
		},
		{
			name:     "truncated array without close bracket",
			args:     []string{"--json-input", "-c", "1.99.x"},
			stdin:    `[{"version":"go1.26.3"},{"version":"go1.26.2"}`,
			wantExit: 1,
			wantErr:  "--json-input",
		},
		{
			name:     "trailing junk after closing bracket",
			args:     []string{"--json-input", "-c", "1.99.x"},
			stdin:    `[{"version":"go1.26.3"}] garbage`,
			wantExit: 1,
			wantErr:  "trailing data",
		},
		{
			name:     "empty stdin",
			args:     []string{"--json-input", "-c", "1.26.x"},
			stdin:    ``,
			wantExit: 1,
			wantErr:  "--json-input",
		},
	}
	for _, td := range cases {
		t.Run(td.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			exit := run(td.args, strings.NewReader(td.stdin), &stdout, &stderr)
			if exit != td.wantExit {
				t.Fatalf("exit: want %d, got %d (stderr=%q stdout=%q)", td.wantExit, exit, stderr.String(), stdout.String())
			}
			gotErr := strings.TrimSpace(stderr.String())
			if td.wantErr == "" {
				if gotErr != "" {
					t.Fatalf("stderr: want empty, got %q", gotErr)
				}
			} else if !strings.Contains(gotErr, td.wantErr) {
				t.Fatalf("stderr: want substring %q, got %q", td.wantErr, gotErr)
			}
			gotOut := strings.TrimSpace(stdout.String())
			if len(td.wantOut) == 0 {
				if gotOut != "" {
					t.Fatalf("stdout: want empty, got %q", gotOut)
				}
				return
			}
			gotLines := strings.Split(gotOut, "\n")
			if !reflect.DeepEqual(gotLines, td.wantOut) {
				t.Fatalf("stdout: want %v, got %v", td.wantOut, gotLines)
			}
		})
	}
}

// Test_run_jsonInput_nilStdin covers the explicit nil-stdin guard in
// runJSONInput. Going through run() requires a non-nil reader (the binary
// always supplies os.Stdin), so this is a defensive belt-and-suspenders
// case for callers that bypass run().
func Test_run_jsonInput_nilStdin(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := run([]string{"--json-input", "-c", "1.26.x"}, nil, &stdout, &stderr)
	if exit != 1 {
		t.Fatalf("exit: want 1, got %d", exit)
	}
	if !strings.Contains(stderr.String(), "requires stdin") {
		t.Fatalf("stderr: want %q, got %q", "requires stdin", stderr.String())
	}
}
