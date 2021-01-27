#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$(dirname -- "$0")")"

setUp() {
  . src/lib
}

test_install_go() {
  (
    export RUNNER_TEMP="$SHUNIT_TMPDIR/test_install_go/runner_temp"
    target="$SHUNIT_TMPDIR/test_install_go/go_target"
    version="1.15.4"
    install_go "$version" "$target"
    got_version="$("$target/bin/go" version)"
    assertEquals "go version go1.15.4 $(goos)/amd64" "$got_version"
  )
}

. ./third_party/shunit2/shunit2
