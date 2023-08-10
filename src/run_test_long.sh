#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$0")/.."

. src/lib

# Nothing special about this one. It just happens to be HEAD of main when writing this.
# Most recent version is go1.21rc4
STABLE_VERSIONS_URL="https://raw.githubusercontent.com/WillAbides/goreleases/077db58ac86a8a2fb63c90817090e132eded0f3d/versions.txt"

do_test_run() {
  CONSTRAINT="$1"
  WANT_VERSION="$2"
  tmp_dir="$SHUNIT_TMPDIR"/test_run
  rm -rf -- "$tmp_dir"
  mkdir -p -- "$tmp_dir"
  export RUNNER_TEMP="$tmp_dir/runner_temp"
  export RUNNER_TOOL_CACHE="$tmp_dir/runner_tool_cache"
  export RUNNER_WORKSPACE="$tmp_dir/runner_workspace"
  export GITHUB_OUTPUT="$tmp_dir/github_output"
  export GITHUB_ACTION_PATH="$PWD"
  export VERSIONS_URL="${VERSIONS_URL:-$STABLE_VERSIONS_URL}"
  export SKIP_MATCHER=1
  export IGNORE_LOCAL_GO=1
  export GOROOT=""
  GO_VERSION="$CONSTRAINT" ./src/run
  WANT_GOROOT="$RUNNER_WORKSPACE/setup-go-faster/go/$WANT_VERSION/x64"
  echo "start GITHUB_OUTPUT"
  cat "$GITHUB_OUTPUT"
  echo "end GITHUB_OUTPUT"
  assertContains "$(grep '^GOROOT=' "$GITHUB_OUTPUT")" "$WANT_GOROOT"
}

test_run_1_15_x() {
  do_test_run 1.15.x 1.15.15
}

test_run_star() {
  do_test_run '*' 1.20.7
}

test_run_1_16rc1() {
  do_test_run 1.16rc1 1.16rc1
}

. ./external/shunit2
