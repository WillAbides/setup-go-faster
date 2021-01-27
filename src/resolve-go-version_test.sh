#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$(dirname -- "$0")")"

ex_go_versions='1.15.7
1.15.6
1.14.3
1.15.6
1.15.5
1.15.4
1.15.3
1.15.2
1.15.1
1.15
1.14.3
1.14.2
1.14.1
1.14
1.13.3
1.13.2
1.13.1
1.13
1.3.3
1.3.2
1.3.1
1.3
1.2.2
1
1.16beta1'

oneTimeSetUp() {
  tmpspace="${RUNNER_TEMP:-"$SHUNIT_TMPDIR/resolve-go-version_test"}"
  ex_dl_json='[]'
  for ver in $ex_go_versions; do
    th="$(printf '. + [{"version": "go%s"}]' "$ver")"
    ex_dl_json="$(echo "$ex_dl_json" | jq "$th")"
  done
}

test_version() {
  export dl_json="$ex_dl_json"
  tmpdir="$tmpspace/${FUNCNAME[0]}"
  toolcache="$tmpdir/go"
  mkdir -p "$toolcache"
  touch "$toolcache/1.14.2"
  touch "$toolcache/1.15.6"

  tests='*;1.15.7
1.15;1.15
1.15.x;1.15.7
^1;1.15.7
1.13.x;1.13.3
tip;tip
1.15beta1;1.15beta1
1.16beta1;1.16beta1
x;1.15.7'
  for td in $tests; do
    input="$(echo "$td" | cut -d ';' -f1)"
    want="$(echo "$td" | cut -d ';' -f2)"
    got="$(./src/resolve-go-version "$input" "$toolcache")"
    assertEquals "failed on input '$input'" "$want" "$got"
  done
}

test_version_ignore_local_go() {
  export dl_json="$ex_dl_json"
  tmpdir="$tmpspace/${FUNCNAME[0]}"
  toolcache="$tmpdir/go"
  mkdir -p "$toolcache"
  touch "$toolcache/1.14.2"
  touch "$toolcache/1.15.6"

  tests='*;1.15.7
1.15;1.15
1.15.x;1.15.7
^1;1.15.7
1.13.x;1.13.3
tip;tip
1.15beta1;1.15beta1
1.16beta1;1.16beta1'

  for td in $tests; do
    input="$(echo "$td" | cut -d ';' -f1)"
    want="$(echo "$td" | cut -d ';' -f2)"
    got="$(IGNORE_LOCAL_GO=1 ./src/resolve-go-version "$input" "$toolcache")"
    assertEquals "failed on input '$input'" "$want" "$got"
  done
}

. ./external/shunit2
