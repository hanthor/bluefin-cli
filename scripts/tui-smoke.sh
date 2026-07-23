#!/usr/bin/env bash
# End-to-end TUI smoke test: drives the real binary in a tmux pane and
# asserts on captured screen content. Catches integration breakage the
# unit tests can't see (terminal init, rendering, key handling, legacy
# flow handoff). Requires tmux.
#
# Usage: scripts/tui-smoke.sh [path-to-binary]
set -euo pipefail

BIN=${1:-./bluefin-cli}
SESSION=bfc-smoke-$$
DIR=$(mktemp -d)
trap 'tmux kill-session -t "$SESSION" 2>/dev/null || true; rm -rf "$DIR"' EXIT

[ -x "$BIN" ] || { echo "FAIL: binary not found: $BIN (build it first)"; exit 1; }
command -v tmux >/dev/null || { echo "SKIP: tmux not available"; exit 0; }

fail=0
assert() { # assert <name> <pattern>
  if grep -q "$2" "$DIR/cap.txt"; then
    echo "ok   $1"
  else
    echo "FAIL $1: expected /$2/ in capture:"
    sed 's/^/  | /' "$DIR/cap.txt" | head -20
    fail=1
  fi
}
cap() { tmux capture-pane -t "$SESSION" -p > "$DIR/cap.txt"; }
keys() { tmux send-keys -t "$SESSION" "$@"; sleep "${SLEEP:-0.5}"; }

tmux new-session -d -s "$SESSION" -x 100 -y 28 "$BIN menu"
sleep 2

cap
assert "main menu renders"        "Bluefin CLI › Home"
assert "menu items present"       "Install Apps"
assert "footer hints present"     "filter"
assert "dino is on screen"        "⣯"

keys j; keys j; keys Enter; sleep 0.5; cap
assert "drill-down breadcrumb"    "› Install Apps"
assert "categories listed"        "CLI Essentials"
assert "back hint appears"        "back"

keys Escape; cap
assert "escape pops to home"      "Bluefin CLI › Home *$"

keys /; keys s t a; cap
assert "filter echoes query"      "🔎 sta"
assert "filter matches fuzzily"   "Starship"

keys Escape; sleep 0.3; keys C-p; cap
assert "palette opens"            "› Palette"

keys Escape; sleep 0.3; keys Escape; sleep 0.3
keys "?"; cap
assert "help overlay opens"       "Keys"

keys Escape; sleep 0.3; keys q; sleep 1
if tmux has-session -t "$SESSION" 2>/dev/null; then
  echo "FAIL: q did not quit the app"
  fail=1
else
  echo "ok   q quits"
fi

[ "$fail" = 0 ] && echo "TUI smoke: all checks passed"
exit "$fail"
