// Command goversion-select picks the highest Go release version from a
// candidate list that satisfies a constraint expression.
//
// Usage:
//
//	goversion-select -c <constraint> [--orig] [<candidate>... | -]
//
// Behavior is fixed (no flags to toggle it):
//   - candidates are always parsed in Go-native format (go1.20, 1.21rc1, ...)
//   - unparseable candidates are silently ignored
//   - only the single highest match is printed
//
// With --orig the original input string(s) for the top match are printed
// instead of the canonical "MAJOR.MINOR.PATCH[-PRE]" form. Versions are
// keyed by value, so two inputs that parse to the same version (e.g.
// "go1.21" and "1.21") share an entry and both are printed.
//
// A candidate of "-" means: read remaining candidates from stdin, one per
// line.
//
// The underlying model is `version` — a Go release as a numeric (major,
// minor, patch) triple plus an optional pre-release tag like "rc1". The
// struct shape is semver-2.0.0-compatible by design, which is what lets the
// constraint grammar — which uses the familiar `>=`, `^`, `~`, `1.x`, `*-0`
// etc. syntax with hyphenated pre-release tags — Just Work on top of it.
// The candidate parser ingests Go-native strings like "go1.21rc1" directly
// into this struct; nothing is ever "translated" to semver. Build metadata
// (the semver "+meta" suffix) is not supported anywhere — Go releases never
// carry it and no caller of goversion-select needs it.
package main

import (
	"bufio"
	"cmp"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("goversion-select", flag.ContinueOnError)
	fs.SetOutput(stderr)
	constraintFlag := fs.String("c", "", "constraint to match (required)")
	origFlag := fs.Bool("orig", false, "output the original input string for the top match instead of its canonical form")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *constraintFlag == "" {
		fmt.Fprintln(stderr, "goversion-select: -c <constraint> is required")
		return 2
	}
	c, err := parseConstraint(*constraintFlag)
	if err != nil {
		fmt.Fprintf(stderr, "goversion-select: invalid constraint: %q\n", *constraintFlag)
		return 1
	}
	versions, orig := collectCandidates(fs.Args(), stdin)
	top, ok := topMatch(c, versions)
	if !ok {
		return 0
	}
	if *origFlag {
		for _, s := range orig[top] {
			fmt.Fprintln(stdout, s)
		}
		return 0
	}
	fmt.Fprintln(stdout, top.String())
	return 0
}

// collectCandidates parses every candidate in Go-native format, silently
// dropping anything unparseable, and returns both the parsed versions and a
// map from each parsed version to the original input strings that produced
// it. A candidate of "-" switches to reading remaining candidates from
// stdin, one per line.
func collectCandidates(args []string, stdin io.Reader) ([]version, map[version][]string) {
	versions := make([]version, 0, len(args))
	orig := map[version][]string{}
	useStdin := false
	for _, a := range args {
		if a == "-" {
			useStdin = true
			break
		}
		versions, orig = addCandidate(a, versions, orig)
	}
	if useStdin && stdin != nil {
		sc := bufio.NewScanner(stdin)
		for sc.Scan() {
			versions, orig = addCandidate(sc.Text(), versions, orig)
		}
	}
	return versions, orig
}

func addCandidate(
	in string,
	versions []version,
	orig map[version][]string,
) ([]version, map[version][]string) {
	v, err := parseGoVersion(in)
	if err != nil {
		return versions, orig
	}
	orig[v] = append(orig[v], in)
	return append(versions, v), orig
}

// topMatch returns the highest version that satisfies the constraint, or
// (zero, false) if no version matches.
func topMatch(c constraint, versions []version) (version, bool) {
	matches := make([]version, 0, len(versions))
	for _, v := range versions {
		if c.allows(v) {
			matches = append(matches, v)
		}
	}
	if len(matches) == 0 {
		return version{}, false
	}
	return slices.MaxFunc(matches, version.Compare), true
}

var goPattern = regexp.MustCompile(`^(?:go)?([1-9]\d*)(?:\.(0|[1-9]\d*))?(?:\.(0|[1-9]\d*))?([a-zA-Z][a-zA-Z0-9.-]*)?$`)

// parseGoVersion parses a Go-native version string (with optional "go"
// prefix and a pre-release suffix attached directly to the numbers, e.g.
// "go1.21rc1") into a version. The pre-release segment is stored as-is, so
// 1.21rc1 round-trips through String() as "1.21.0-rc1".
func parseGoVersion(ver string) (version, error) {
	m := goPattern.FindStringSubmatch(ver)
	if len(m) == 0 {
		return version{}, fmt.Errorf("could not parse version %q", ver)
	}
	major, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return version{}, fmt.Errorf("could not parse major version %q: %v", ver, err)
	}
	var minor, patch uint64
	if m[2] != "" {
		minor, err = strconv.ParseUint(m[2], 10, 64)
		if err != nil {
			return version{}, fmt.Errorf("could not parse minor version %q: %v", ver, err)
		}
	}
	if m[3] != "" {
		patch, err = strconv.ParseUint(m[3], 10, 64)
		if err != nil {
			return version{}, fmt.Errorf("could not parse patch version %q: %v", ver, err)
		}
	}
	return newVersion(major, minor, patch, m[4]), nil
}

// version is a Go release version.
type version struct {
	major, minor, patch uint64
	pre                 string
}

// newVersion constructs a version from explicit numeric parts and an optional
// pre-release tag (without the leading hyphen).
func newVersion(major, minor, patch uint64, pre string) version {
	return version{major: major, minor: minor, patch: patch, pre: pre}
}

// String returns the canonical "MAJOR.MINOR.PATCH[-PRE]" form. The hyphen-
// separated pre-release form is what downstream callers in src/lib expect —
// notably `normalize_go_version` relies on it to feed `is_precise_version`.
func (v version) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d.%d.%d", v.major, v.minor, v.patch)
	if v.pre != "" {
		sb.WriteByte('-')
		sb.WriteString(v.pre)
	}
	return sb.String()
}

// Compare returns -1, 0, or 1 if v is less than, equal to, or greater than o.
// Pre-release ordering follows semver 2.0.0.
func (v version) Compare(o version) int {
	return cmp.Or(
		cmp.Compare(v.major, o.major),
		cmp.Compare(v.minor, o.minor),
		cmp.Compare(v.patch, o.patch),
		comparePre(v.pre, o.pre),
	)
}

// comparePre orders two pre-release strings per the semver 2.0.0 spec:
//
//   - An empty pre-release is greater than any non-empty pre-release (because
//     1.0.0 > 1.0.0-rc1).
//   - Otherwise each dot-separated identifier is compared in turn.
//   - Numeric identifiers compare numerically and rank below non-numeric ones.
//   - If all shared identifiers are equal, the longer set wins.
func comparePre(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	apLen := len(ap)
	bpLen := len(bp)
	for i := range min(apLen, bpLen) {
		d := cmpIdent(ap[i], bp[i])
		if d != 0 {
			return d
		}
	}
	return cmp.Compare(apLen, bpLen)
}

func cmpIdent(a, b string) int {
	an, aerr := strconv.ParseUint(a, 10, 64)
	bn, berr := strconv.ParseUint(b, 10, 64)
	switch {
	case aerr == nil && berr == nil:
		return cmp.Compare(an, bn)
	case aerr == nil:
		// numeric identifiers always rank below alphanumeric ones
		return -1
	case berr == nil:
		return 1
	}
	return cmp.Compare(a, b)
}

// constraint is a parsed constraint expression. A version satisfies the
// expression when it satisfies every atom in at least one group; groups are
// joined by "||", atoms within a group are joined by whitespace or commas.
type constraint struct {
	groups [][]atom
}

// wildLevel describes how much of an atom's version was wildcarded.
//
//	wildNone   "1.2.3"
//	wildPatch  "1.2"  or "1.2.x"
//	wildMinor  "1"    or "1.x"
//	wildMajor  "*"    or "x"
type wildLevel int8

const (
	wildNone wildLevel = iota
	wildPatch
	wildMinor
	wildMajor
)

type atom struct {
	op   string
	ver  version
	wild wildLevel
}

// parseConstraint parses a constraint string. Returns an error if any atom
// is malformed or any group is empty.
func parseConstraint(s string) (constraint, error) {
	parts := strings.Split(s, "||")
	groups := make([][]atom, 0, len(parts))
	for _, g := range parts {
		atoms, err := parseGroup(g)
		if err != nil {
			return constraint{}, err
		}
		groups = append(groups, atoms)
	}
	return constraint{groups: groups}, nil
}

// allows reports whether v satisfies the constraint.
func (c constraint) allows(v version) bool {
	for _, group := range c.groups {
		if allMatch(group, v) {
			return true
		}
	}
	return false
}

func allMatch(group []atom, v version) bool {
	for i := range group {
		if !group[i].check(v) {
			return false
		}
	}
	return true
}

// stickyOpRe glues a comparison operator to the following segment so that
// "> 1.2.3" becomes ">1.2.3" and a simple field split sees one token per atom.
var stickyOpRe = regexp.MustCompile(`(<=|>=|!=|==|=|<|>|\^|~)\s+`)

func parseGroup(s string) ([]atom, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty constraint group")
	}
	s = stickyOpRe.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty constraint group")
	}
	atoms := make([]atom, 0, len(parts))
	for _, p := range parts {
		a, err := parseAtom(p)
		if err != nil {
			return nil, err
		}
		atoms = append(atoms, a)
	}
	return atoms, nil
}

func parseAtom(s string) (atom, error) {
	orig := s
	invalid := func() (atom, error) {
		return atom{}, fmt.Errorf("invalid constraint: %q", orig)
	}

	// Peel optional operator prefix.
	var op string
	for _, candidate := range []string{"<=", ">=", "!=", "==", "=", "<", ">", "^", "~"} {
		rest, ok := strings.CutPrefix(s, candidate)
		if !ok {
			continue
		}
		op = candidate
		s = rest
		break
	}
	if op == "==" {
		op = "="
	}

	// Split off the pre-release tag, if any.
	var pre string
	var ok bool
	s, pre, ok = strings.Cut(s, "-")
	if ok && !validPre(pre) {
		return invalid()
	}

	// Parse 1–3 dot-separated numeric or wildcard segments.
	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return invalid()
	}
	a := atom{op: op, ver: version{pre: pre}}
	fields := [3]*uint64{&a.ver.major, &a.ver.minor, &a.ver.patch}
	levels := [3]wildLevel{wildMajor, wildMinor, wildPatch}
	for i := range 3 {
		if i >= len(parts) {
			a.wild = levels[i]
			break
		}
		if isWild(parts[i]) {
			a.wild = levels[i]
			break
		}
		n, err := strconv.ParseUint(parts[i], 10, 64)
		if err != nil {
			return invalid()
		}
		*fields[i] = n
	}
	return a, nil
}

func isWild(s string) bool {
	return s == "x" || s == "X" || s == "*"
}

// validPre reports whether s is a non-empty dot-separated list of
// pre-release identifiers per the semver grammar — each identifier is one
// or more of [0-9A-Za-z-].
func validPre(s string) bool {
	if s == "" {
		return false
	}
	for _, ident := range strings.Split(s, ".") {
		if ident == "" {
			return false
		}
		for _, r := range ident {
			switch {
			case r >= '0' && r <= '9':
			case r >= 'a' && r <= 'z':
			case r >= 'A' && r <= 'Z':
			case r == '-':
			default:
				return false
			}
		}
	}
	return true
}

// check reports whether v satisfies this atom.
//
// A candidate with a non-empty pre-release segment only satisfies an atom
// whose constraint version also carries a pre-release segment. This is the
// rule that makes "*-0" the canonical "match everything including
// pre-releases" idiom.
func (a atom) check(v version) bool {
	if v.pre != "" && a.ver.pre == "" {
		return false
	}
	switch a.op {
	case "", "=":
		return a.matchEqual(v)
	case "!=":
		return !a.matchEqual(v)
	case ">":
		return a.matchGreater(v)
	case ">=":
		return v.Compare(a.ver) >= 0
	case "<":
		return v.Compare(a.ver) < 0
	case "<=":
		return a.matchLessEq(v)
	case "~":
		return a.matchTilde(v)
	case "^":
		return a.matchCaret(v)
	}
	return false
}

// matchEqual implements both "" and "=" semantics. For wildcarded atoms the
// candidate must fall anywhere within the wild range; for fully-specified
// atoms it is strict equality.
func (a atom) matchEqual(v version) bool {
	switch a.wild {
	case wildMajor:
		return true
	case wildMinor:
		return v.major == a.ver.major
	case wildPatch:
		return v.major == a.ver.major && v.minor == a.ver.minor
	default:
		return v.Compare(a.ver) == 0
	}
}

// matchGreater implements ">". For wildcarded atoms the candidate must lie
// strictly above the entire wild range — so ">1.2" matches 1.3.0 but not
// 1.2.99.
func (a atom) matchGreater(v version) bool {
	switch a.wild {
	case wildMajor:
		return false
	case wildMinor:
		return v.major > a.ver.major
	case wildPatch:
		return v.major > a.ver.major ||
			(v.major == a.ver.major && v.minor > a.ver.minor)
	default:
		return v.Compare(a.ver) > 0
	}
}

// matchLessEq implements "<=". For wildcarded atoms the candidate must lie
// at or below the top of the wild range — so "<=1.2" matches 1.2.99 (the
// entire 1.2.x range) but not 1.3.0.
func (a atom) matchLessEq(v version) bool {
	switch a.wild {
	case wildMajor:
		return true
	case wildMinor:
		return v.major <= a.ver.major
	case wildPatch:
		return v.major < a.ver.major ||
			(v.major == a.ver.major && v.minor <= a.ver.minor)
	default:
		return v.Compare(a.ver) <= 0
	}
}

// matchTilde implements "~". Tilde allows changes to the rightmost specified
// segment:
//
//	~1.2.3  -> >=1.2.3, <1.3.0  (same major + minor)
//	~1.2    -> >=1.2.0, <1.3.0  (same major + minor)
//	~1      -> >=1.0.0, <2.0.0  (same major)
func (a atom) matchTilde(v version) bool {
	switch a.wild {
	case wildMajor:
		return true
	case wildMinor:
		return v.major == a.ver.major
	default:
		return v.major == a.ver.major &&
			v.minor == a.ver.minor &&
			v.Compare(a.ver) >= 0
	}
}

// matchCaret implements "^". Caret allows changes that do not modify the
// leftmost non-zero segment:
//
//	^1.2.3  -> >=1.2.3, <2.0.0
//	^0.2.3  -> >=0.2.3, <0.3.0
//	^0.0.3  -> >=0.0.3, <0.0.4
//	^1      -> >=1.0.0, <2.0.0  (minor wild, major locks)
//	^0      -> >=0.0.0, <1.0.0  (minor wild, major locks)
func (a atom) matchCaret(v version) bool {
	if a.wild == wildMajor {
		return true
	}
	if a.ver.major != 0 {
		return v.major == a.ver.major && v.Compare(a.ver) >= 0
	}
	// Major is zero: leftmost non-zero segment (if any) locks.
	if a.wild == wildMinor {
		return v.major == 0
	}
	if a.ver.minor != 0 {
		return v.major == 0 &&
			v.minor == a.ver.minor &&
			v.Compare(a.ver) >= 0
	}
	if a.wild == wildPatch {
		return v.major == 0 && v.minor == 0
	}
	// ^0.0.Z (no wild) — only that exact version (and pre-release siblings).
	return v.major == 0 && v.minor == 0 && v.patch == a.ver.patch &&
		v.Compare(a.ver) >= 0
}
