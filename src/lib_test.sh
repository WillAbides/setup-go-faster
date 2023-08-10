#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$(dirname -- "$0")")"

setUp() {
  . src/lib
}

test_homedir() {
  (
    export USERPROFILE="windows home"
    export HOME="my home"
    assertEquals "$HOME" "$(RUNNER_OS=Linux homedir)"
    assertEquals "$USERPROFILE" "$(RUNNER_OS=Windows homedir)"
  )
}

test_download_go_url() {
  assertEquals \
    "https://storage.googleapis.com/golang/go1.15.5.linux-amd64.tar.gz" \
    "$(RUNNER_OS=Linux download_go_url "1.15.5")"

  assertEquals \
    "https://storage.googleapis.com/golang/go1.15.5.darwin-amd64.tar.gz" \
    "$(RUNNER_OS=macOS download_go_url "1.15.5")"

  assertEquals \
    "https://storage.googleapis.com/golang/go1.15.5.windows-amd64.zip" \
    "$(RUNNER_OS=Windows download_go_url "1.15.5")"
}

test_is_precise_version() {
  versions='
1
1.15
1.15.1
1.1.1
9999.9999.9999
1.15beta1
1.16rc1
1.21.0
1.21.1
  '

  for v in $versions; do
    is_precise_version "$v"
    assertTrue "$v" $?
  done

  not_versions='
1.21
*
1.x
1.15.x
^ 1.15.1
  '

  echo "$not_versions" | while IFS= read -r v; do
    is_precise_version "$v"
    assertFalse "$v" $?
  done
}

do_test_select_go_version() {
  local want="$1"
  local pattern="$2"
  local versions="${*:3}"
  local got
  local test_name="\n pattern: $pattern \n versions: $versions\n"

  got="$(select_go_version "$pattern" "$(echo "$versions" | tr " " "\n")")"
  r_val=$?
  if [ "$want" = "" ]; then
    assertFalse " unexpected exit code\n$test_name" $r_val
  else
    assertTrue " unexpected exit code\n$test_name" $r_val
  fi
  assertEquals " unexpected value\n$test_name" "$want" "$got" || true
}

test_select_go_version() {
  do_test_select_go_version "1.2" "1.x" "1.1" "1.2"
  do_test_select_go_version "" "1.13.x" "1.14.2" "1.15.6"
  do_test_select_go_version "" "^1.13.0" "1.14.2" "1.15.6"
  do_test_select_go_version "1.13.3" "1.x" "1.13.3"
  do_test_select_go_version "1.13.3" "1.13.x" "1.13.3"
  do_test_select_go_version "1.13.3" "x" "1.13.3"
  do_test_select_go_version "1.13.3" "x" "1.13.3" "1.14beta1"
  do_test_select_go_version "" "x" "1.14beta1"
  do_test_select_go_version "0" "x" "0"
  do_test_select_go_version "" "1.2.3"
  do_test_select_go_version "1.21.0" "1.21.x" "1.21.0"
  do_test_select_go_version "1.21.0" "1.21" "1.21.0"
  do_test_select_go_version "go1.20" "1.20.x" "go1.20"
}

test_select_remote_version() {
  versions='1.15.7
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
1.16beta1
1.16rc1'

  tests='*;1.15.7
1.17.x;
1.16.x;
1.15;1.15
1.15.x;1.15.7
^1;1.15.7
^1.15.999;
1.13.x;1.13.3
1.16beta1;1.16beta1
x;1.15.7'

  for td in $tests; do
    input="$(echo "$td" | cut -d ';' -f1)"
    want="$(echo "$td" | cut -d ';' -f2)"
    got="$(select_remote_version "$input" "$versions")"
    assertEquals "failed on input '$input'" "$want" "$got"
  done
}

. ./external/shunit2
