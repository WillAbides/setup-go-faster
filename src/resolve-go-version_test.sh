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
  ex_dl_json='[]'
  for ver in $ex_go_versions; do
    th="$(printf '. + [{"version": "go%s"}]' "$ver")"
    ex_dl_json="$(echo "$ex_dl_json" | jq "$th")"
  done
}

available_head_versions=''

mock_version_available_to_dl() {
  v="$1"
  for ver in $available_head_versions ; do
      if [ "$ver" = "$v" ]; then
        return 0
      fi
  done
  return 1
}

mock_head_versions() {
  export available_head_versions="$1"
  export mock_head_go_version="1"
}

test_version() {
  export dl_json="$ex_dl_json"
  tmpdir="$SHUNIT_TMPDIR/${FUNCNAME[0]}"
  toolcache="$tmpdir/go"
  mkdir -p "$toolcache"
  touch "$toolcache/1.14.2"
  touch "$toolcache/1.15.6"
  mock_head_versions '1.15
1.15beta1'

  tests='*;1.15.6
1.15;1.15
1.15.x;1.15.6
^1;1.15.6
1.13.x;1.13.3
tip;tip
1.15beta1;1.15beta1
1.16beta1;1.16beta1'
  for td in $tests; do
    input="$(echo "$td" | cut -d ';' -f1)"
    want="$(echo "$td" | cut -d ';' -f2)"
    got="$(./src/resolve-go-version "$input" "$toolcache")"
    assertEquals "failed on input '$input'" "$want" "$got"
  done
}


test_version_ignore_local_go() {
  export dl_json="$ex_dl_json"
  tmpdir="$SHUNIT_TMPDIR/${FUNCNAME[0]}"
  toolcache="$tmpdir/go"
  mkdir -p "$toolcache"
  touch "$toolcache/1.14.2"
  touch "$toolcache/1.15.6"
    mock_head_versions '1.15
1.15beta1'

  export version_available_to_dl_func="mock_version_available_to_dl"

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

. ./third_party/shunit2/shunit2
