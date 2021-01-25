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

. ./third_party/shunit2/shunit2
