#!/bin/sh

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
  '

  for v in $versions; do
    is_precise_version "$v"
    assertTrue "$v" $?
  done

  not_versions='
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

. ./third_party/shunit2/shunit2
