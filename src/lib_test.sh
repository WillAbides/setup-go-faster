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

do_test_select_go_version() {
  local want="$1"
  local pattern="$2"
  local versions="${*:3}"
  local got
  local test_name="\n pattern: $pattern \n versions: $versions\n"
  versions="$(echo "$versions" | tr " " "\n")"
  got="$(select_go_version "$pattern" "$versions")"
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
  do_test_select_go_version "1.15.6" "^1.13.0" "1.14.2" "1.15.6"
  do_test_select_go_version "1.13.3" "1.x" "1.13.3"
  do_test_select_go_version "1.13.3" "1.13.x" "1.13.3"
  do_test_select_go_version "1.13.3" "x" "1.13.3"
  do_test_select_go_version "1.13.3" "x" "1.13.3" "1.14beta1"
  do_test_select_go_version "0" "x" "0"
  do_test_select_go_version "" "1.2.3"
}

. ./external/shunit2
