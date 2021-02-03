#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$(dirname -- "$0")")"

setUp() {
  . src/lib
}

test_install_go() {
  (
    tmpspace="${RUNNER_TEMP:-"$SHUNIT_TMPDIR/test_install_go"}"
    target="$tmpspace/go_target"
    version="1.15.4"
    install_go "$version" "$target"
    got_version="$("$target/bin/go" version)"
    assertEquals "go version go1.15.4 $(goos)/amd64" "$got_version"
  )
}

. ./external/shunit2
