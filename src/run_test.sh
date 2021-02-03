#!/bin/bash

set -e

CDPATH="" cd -- "$(dirname -- "$(dirname -- "$0")")"

test_run() {
  tmpspace="$SHUNIT_TMPDIR/test_run"

  install_go_args() {
    jq \
      --arg install_go_version "$1" \
      --arg target_dir "$2" \
      --arg tip_target_dir "$3" \
      --arg install_go_tip "$4" \
      -n '{
       "install_go_version": $install_go_version,
       "target_dir": $target_dir,
       "tip_target_dir": $tip_target_dir,
       "install_go_tip": $install_go_tip
      }'
  }
  export -f install_go_args
  dummy_run_install_go() {
    install_go_args "$@" "$INSTALL_GO_TIP"
  }

  export run_install_go_fn="dummy_run_install_go"
  export -f dummy_run_install_go
  export RUNNER_TOOL_CACHE="$tmpspace/toolcache"
  mkdir -p "$RUNNER_TOOL_CACHE/go"
  export RUNNER_WORKSPACE="$tmpspace/runner_workspace"
  workspace_go="$RUNNER_WORKSPACE/setup-go-faster/go"
  mkdir -p "$RUNNER_WORKSPACE/go"
  mkdir -p "$RUNNER_TOOL_CACHE/go/1.13.3/x64"
  mkdir -p "$RUNNER_TOOL_CACHE/go/1.14.5/x64"
  mkdir -p "$workspace_go/1.15.2/x64"

  want_tip="$workspace_go/tip/x64"

  got="$(GO_VERSION='1.13.4' ./src/run)"
  want="$(install_go_args "1.13.4" "$workspace_go/1.13.4/x64" "$want_tip")"
  assertEquals "1.13.4" "$want" "$got"

  got="$(GO_VERSION='1.15.x' ./src/run)"
  want="$(install_go_args "1.15.2" "$workspace_go/1.15.2/x64" "$want_tip")"
  assertEquals "1.15.x" "$want" "$got"

  got="$(GO_VERSION='1.13.3' ./src/run)"
  want="$(install_go_args "1.13.3" "$RUNNER_TOOL_CACHE/go/1.13.3/x64" "$want_tip")"
  assertEquals "1.13.3" "$want" "$got"

  got="$(GO_VERSION='1.13.x' ./src/run)"
  want="$(install_go_args "1.13.3" "$RUNNER_TOOL_CACHE/go/1.13.3/x64" "$want_tip")"
  assertEquals "1.13.x" "$want" "$got"

  got="$(GO_VERSION='tip' ./src/run)"
  want="$(install_go_args "1.15.2" "$workspace_go/1.15.2/x64" "$want_tip" "1")"
  assertEquals "tip" "$want" "$got"

  # 999.999 will probably never exist, so we know it is resolved without calling the endpoint
  got="$(GO_VERSION='999.999' ./src/run)"
  want="$(install_go_args "999.999" "$workspace_go/999.999/x64" "$want_tip")"
  assertEquals "999.999" "$want" "$got"
}

. ./external/shunit2
