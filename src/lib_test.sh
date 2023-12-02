#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$0")/.."

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
  run_test() {
    local system="$1"
    local go_version="$2"
    local want_url="$3"
    got_url="$(download_go_url "$go_version" "$system")"
    assertEquals "$want_url" "$got_url"
  }

  run_test linux/amd64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.linux-amd64.tar.gz"
  run_test linux/386 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.linux-386.tar.gz"
  run_test linux/arm64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.linux-arm64.tar.gz"
  run_test darwin/amd64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.darwin-amd64.tar.gz"
  run_test darwin/arm64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.darwin-arm64.tar.gz"
  run_test windows/amd64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.windows-amd64.zip"
  run_test windows/386 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.windows-386.zip"
  run_test windows/arm64 1.15.5 "https://storage.googleapis.com/golang/go1.15.5.windows-arm64.zip"
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
  versions='go1.15.7
go1.15.6
go1.14.3
go1.15.6
go1.15.5
go1.15.4
go1.15.3
go1.15.2
go1.15.1
go1.15
go1.14.3
go1.14.2
go1.14.1
go1.14
go1.13.3
go1.13.2
go1.13.1
go1.13
go1.3.3
go1.3.2
go1.3.1
go1.3
go1.2.2
go1
go1.16beta1
go1.16rc1'

  tests='*;go1.15.7
1.17.x;
1.16.x;
1.15;go1.15
1.15.x;go1.15.7
^1;go1.15.7
^1.15.999;
1.13.x;go1.13.3
1.16beta1;go1.16beta1
x;go1.15.7'

  for td in $tests; do
    input="$(echo "$td" | cut -d ';' -f1)"
    want="$(echo "$td" | cut -d ';' -f2)"
    got="$(select_remote_version "$input" "$versions")"
    assertEquals "failed on input '$input'" "$want" "$got"
  done
}

test_supported_system() {
  assertTrue "linux/amd64" 'supported_system "linux/amd64"'
  assertTrue "linux/386" 'supported_system "linux/386"'
  assertTrue "linux/arm64" 'supported_system "linux/arm64"'
  assertTrue "darwin/amd64" 'supported_system "darwin/amd64"'
  assertTrue "darwin/arm64" 'supported_system "darwin/arm64"'
  assertTrue "windows/amd64" 'supported_system "windows/amd64"'
  assertTrue "windows/386" 'supported_system "windows/386"'
  assertTrue "windows/arm64" 'supported_system "windows/arm64"'

  assertFalse "linux/arm" 'supported_system "linux/arm"'
  assertFalse "darwin/386" 'supported_system "darwin/386"'
  assertFalse "asdf" 'supported_system "asdf"'
  assertFalse "" 'supported_system ""'
  assertFalse " " 'supported_system " "'

}

. ./external/shunit2
