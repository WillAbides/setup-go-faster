#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$0")/.."

setUp() {
  . src/lib
}

# Nothing special about this one. It just happens to be HEAD of main when writing this.
# Most recent version is go1.21rc4
STABLE_VERSIONS_URL="https://raw.githubusercontent.com/WillAbides/goreleases/077db58ac86a8a2fb63c90817090e132eded0f3d/versions.txt"

do_test_run() {
  tmp_dir="$SHUNIT_TMPDIR"/test_run
  rm -rf -- "$tmp_dir"
  mkdir -p -- "$tmp_dir"
  export RUNNER_TEMP="$tmp_dir/runner_temp"
  export RUNNER_TOOL_CACHE="$tmp_dir/runner_tool_cache"
  export RUNNER_WORKSPACE="$tmp_dir/runner_workspace"
  export GITHUB_OUTPUT="$tmp_dir/github_output"
  export GITHUB_ACTION_PATH="$PWD"
  export VERSIONS_URL="$STABLE_VERSIONS_URL"
  export SKIP_MATCHER=1
  export IGNORE_LOCAL_GO=1
  export GOROOT=""
  ./src/run
  WANT_GOROOT="$RUNNER_WORKSPACE/setup-go-faster/go/$WANT_VERSION/x64"
  if [ "$(goos)" = "windows" ]; then
    # Windows is trickier because of how it translates paths.
    # Just check that is has the right suffix.
    WANT_GOROOT="setup-go-faster\\go\\$WANT_VERSION\\x64"
  fi
  assertContains "$(grep '^GOROOT=' "$GITHUB_OUTPUT")" "$WANT_GOROOT"
}

test_run_1_16_x() {
  GO_VERSION="1.16.x" \
    WANT_VERSION="1.16.15" \
    do_test_run
}

test_run_star() {
  GO_VERSION="*" \
    WANT_VERSION="1.20.7" \
    do_test_run
}

test_run_1_16rc1() {
  GO_VERSION="1.16rc1" \
    WANT_VERSION="1.16rc1" \
    do_test_run
}

test_run_1_21rc4() {
  GO_VERSION="1.21rc4" \
    WANT_VERSION="1.21rc4" \
    do_test_run
}

test_go_mod() {
  GO_VERSION_FILE="$SHUNIT_TMPDIR"/test_go_mod/go.mod
  mkdir -p -- "$(dirname -- "$GO_VERSION_FILE")"
  echo "
module foo
go 1.20.7 // the go version
" > "$GO_VERSION_FILE"
  GO_VERSION_FILE="$GO_VERSION_FILE" \
    WANT_VERSION="1.20.7" \
    do_test_run
}

. ./external/shunit2
