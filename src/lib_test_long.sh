#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$0")/.."

setUp() {
  . src/lib
}

test_install_go() {
  (
    tmpspace="${RUNNER_TEMP:-"$SHUNIT_TMPDIR/test_install_go"}"
    RUNNER_TEMP="${RUNNER_TEMP:-"$tmpspace/runner_temp"}"
    export RUNNER_TEMP
    target="$tmpspace/go_target"
    version="1.16.4"
    inst_tmp="$tmpspace/inst_tmp"
    mkdir -p "$inst_tmp"
    install_go "$version" "$target" "$inst_tmp"
    got_version="$("$target/bin/go" version)"
    assertEquals "go version go1.16.4 $(go_system)" "$got_version"
  )
}

. ./external/shunit2
