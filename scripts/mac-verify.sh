#!/bin/bash
# bluefin-cli macOS verification — run on the MacBook.
# Everything state-touching uses a sandbox HOME; your real setup is untouched.
set -u
BIN="$1"
PASS=0; FAIL=0
ok()  { echo "ok   $1"; PASS=$((PASS+1)); }
bad() { echo "FAIL $1"; FAIL=$((FAIL+1)); }

SANDBOX=$(mktemp -d)
trap 'rm -rf "$SANDBOX"' EXIT
run() { HOME="$SANDBOX" SHELL=/bin/zsh "$BIN" "$@"; }

# --- CLI basics ---------------------------------------------------------
v=$("$BIN" --version 2>&1) && ok "version: $v" || bad "binary does not execute: $v"
run status >/dev/null 2>&1 && ok "status runs" || bad "status failed"
run doctor >/dev/null 2>&1; [ $? -le 1 ] && ok "doctor runs" || bad "doctor crashed"
run doctor --bench 2>&1 | grep -q "rc overhead" && ok "doctor --bench measures zsh" || bad "bench failed"

# --- init script syntax under real macOS zsh/bash -----------------------
run init zsh > "$SANDBOX/z.zsh" && zsh -n "$SANDBOX/z.zsh" && ok "init zsh is valid zsh" || bad "init zsh syntax"
run init bash > "$SANDBOX/b.bash" && bash -n "$SANDBOX/b.bash" && ok "init bash is valid bash" || bad "init bash syntax"

# --- desired-state round trip (sandbox HOME) ----------------------------
touch "$SANDBOX/.zshrc"
run shell zsh on >/dev/null 2>&1
grep -q "bluefin-cli init" "$SANDBOX/.zshrc" && ok "shell zsh on reached .zshrc" || bad "enable did not write rc"
run status 2>/dev/null | grep -q "zsh:.*enabled" && ok "status agrees" || bad "status disagrees"
run shell zsh off >/dev/null 2>&1
grep -q "bluefin-cli init" "$SANDBOX/.zshrc" && bad "disable left rc line" || ok "shell zsh off cleaned .zshrc"

run theme mocha >/dev/null 2>&1
grep -q "flavor: mocha" "$SANDBOX/.config/bluefin-cli/config.yaml" && ok "theme persisted" || bad "theme not persisted"

run profile export "$SANDBOX/p.json" >/dev/null 2>&1
grep -q '"version": 1' "$SANDBOX/p.json" && ok "profile export" || bad "profile export"
run theme auto >/dev/null 2>&1
run profile import "$SANDBOX/p.json" >/dev/null 2>&1
grep -q "flavor: mocha" "$SANDBOX/.config/bluefin-cli/config.yaml" && ok "profile import restored flavor" || bad "profile import"

# --- TUI (needs tmux; brew install tmux if missing) ---------------------
if command -v tmux >/dev/null; then
  S=bfcmac-$$
  tmux kill-session -t $S 2>/dev/null
  tmux new-session -d -s $S -x 100 -y 28
  tmux send-keys -t $S "HOME=$SANDBOX SHELL=/bin/zsh TERM=xterm-256color $BIN menu" Enter
  sleep 2.5
  cap=$(tmux capture-pane -t $S -p)
  echo "$cap" | grep -q "Bluefin CLI › Home" && ok "TUI shell renders" || bad "TUI did not render"
  echo "$cap" | grep -q "⣽" && ok "dino on screen" || bad "dino missing"
  tmux send-keys -t $S / ; tmux send-keys -t $S s t a; sleep 0.5
  tmux capture-pane -t $S -p | grep -q "Starship" && ok "fuzzy filter works" || bad "filter broken"
  tmux send-keys -t $S Escape; sleep 0.3
  tmux send-keys -t $S C-p; sleep 0.5; tmux send-keys -t $S d i n o; sleep 0.3; tmux send-keys -t $S Enter; sleep 1.5
  tmux capture-pane -t $S -p | grep -q "score " && ok "dino game runs" || bad "game broken"
  tmux send-keys -t $S Escape; sleep 0.3; tmux send-keys -t $S q; sleep 0.5
  tmux kill-session -t $S 2>/dev/null
else
  echo "note: tmux not installed — TUI checks skipped (brew install tmux)"
fi

echo; echo "macOS results: $PASS ok, $FAIL failed"
[ $FAIL -eq 0 ]
