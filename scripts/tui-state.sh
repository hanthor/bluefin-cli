#!/bin/sh
# TUI end-to-end DESIRED-STATE test: drive the real menu with tmux keys in
# an isolated HOME and assert the files the UI claims to manage actually
# changed. Complements tui-smoke.sh (which asserts what's on screen).
#
# Usage: scripts/tui-state.sh <binary>
set -eu

BIN=$(realpath "${1:?usage: tui-state.sh <binary>}")
SESSION="bfc-state-$$"
H=$(mktemp -d)
trap 'tmux kill-session -t "$SESSION" 2>/dev/null || true; rm -rf "$H"' EXIT

mkdir -p "$H/.config"
printf '# base\n' > "$H/.bashrc"

fail=0
ok()   { echo "ok   $1"; }
bad()  { echo "FAIL $1"; fail=1; }

# wait_for <pattern> <seconds>: poll the pane until the pattern appears.
wait_for() {
  n=0
  while [ $n -lt $(( $2 * 2 )) ]; do
    tmux capture-pane -t "$SESSION" -p | grep -q "$1" && return 0
    sleep 0.5; n=$((n+1))
  done
  return 1
}

tmux new-session -d -s "$SESSION" -x 100 -y 28
tmux send-keys -t "$SESSION" "env HOME=$H SHELL=/bin/bash TERM=xterm-256color $BIN menu" Enter
sleep 2

# Home -> Bluefin Shell (Status, Doctor, Shell) -> toggle current shell on.
tmux send-keys -t "$SESSION" j; sleep 0.2
tmux send-keys -t "$SESSION" j; sleep 0.2
tmux send-keys -t "$SESSION" Enter; sleep 1
tmux send-keys -t "$SESSION" Enter            # runner: "Enabling bash"
wait_for "✓ done" 30 || bad "enable runner did not finish"
tmux send-keys -t "$SESSION" Escape; sleep 0.6

grep -q "bluefin-cli init" "$H/.bashrc" \
  && ok "TUI toggle wrote the init line to .bashrc" \
  || bad "TUI toggle did not reach .bashrc"

# The menu label must now offer to Disable (state agreement in the UI).
tmux capture-pane -t "$SESSION" -p | grep -q "Disable for current shell" \
  && ok "menu label reflects the new state" \
  || bad "menu label did not refresh after toggle"

# Toggle back off and verify the line is gone but the base content survives.
tmux send-keys -t "$SESSION" Enter
wait_for "✓ done" 30 || bad "disable runner did not finish"
tmux send-keys -t "$SESSION" Escape; sleep 0.6
grep -q "bluefin-cli init" "$H/.bashrc" \
  && bad "TUI disable left the init line behind" \
  || ok "TUI disable removed the init line"
grep -q "# base" "$H/.bashrc" \
  && ok "pre-existing rc content survived the round trip" \
  || bad "round trip destroyed pre-existing rc content"

# MOTD toggle through the TUI must land in the shell config JSON.
tmux send-keys -t "$SESSION" j; sleep 0.2   # Components
tmux send-keys -t "$SESSION" j; sleep 0.2   # MOTD
tmux send-keys -t "$SESSION" Enter; sleep 1
tmux send-keys -t "$SESSION" Enter; sleep 1.5  # toggle MOTD
if grep -qi '"Motd"' "$H"/.config/bluefin-cli/*.json 2>/dev/null; then
  ok "TUI MOTD toggle persisted to shell config"
else
  bad "TUI MOTD toggle did not persist"
fi

tmux send-keys -t "$SESSION" Escape; sleep 0.3
tmux send-keys -t "$SESSION" Escape; sleep 0.3
tmux send-keys -t "$SESSION" q; sleep 1

if [ "$fail" -eq 0 ]; then
  echo "TUI state: all checks passed"
else
  echo "TUI state: FAILURES"
  exit 1
fi
