#!/bin/bash
# bluefin-cli macOS phase 2: software install + wallpapers.
# Tier 1 is hermetic (fake brew); tier 2 does a REAL install of cowsay from
# a local Brewfile and uninstalls it after; tier 3 checks the wallpaper flow
# against real Homebrew metadata. Cleanup restores everything.
set -u
export PATH="/opt/homebrew/bin:$PATH"
BIN="$1"
PASS=0; FAIL=0
ok()  { echo "ok   $1"; PASS=$((PASS+1)); }
bad() { echo "FAIL $1"; FAIL=$((FAIL+1)); }
SANDBOX=$(mktemp -d)
trap 'rm -rf "$SANDBOX"' EXIT
run() { HOME="$SANDBOX" SHELL=/bin/zsh "$BIN" "$@"; }

# --- Tier 1: hermetic bundle chain (fake brew) --------------------------
FAKE=$(mktemp -d)
LOG="$FAKE/brew.log"
cat > "$FAKE/brew" <<STUB
#!/bin/sh
echo "\$@" >> $LOG
for a in "\$@"; do case "\$a" in --file=*) cat "\${a#--file=}" >> $LOG ;; esac; done
exit 0
STUB
chmod +x "$FAKE/brew"
HOME="$SANDBOX" PATH="$FAKE:$PATH" "$BIN" install cli >/dev/null 2>&1
if grep -q "bundle install --file=" "$LOG" 2>/dev/null && grep -q 'brew "' "$LOG"; then
  ok "hermetic: bundle chain reaches brew with package lines (darwin)"
else
  bad "hermetic bundle chain broken"; cat "$LOG" 2>/dev/null | head -3
fi
rm -rf "$FAKE"

# --- Tier 2: REAL install from a local Brewfile -------------------------
if brew list cowsay >/dev/null 2>&1; then
  echo "note: cowsay already installed — skipping real-install tier"
else
  printf 'brew "cowsay"\n' > "$SANDBOX/Brewfile"
  run install "$SANDBOX/Brewfile" >/dev/null 2>&1
  if brew list cowsay >/dev/null 2>&1; then
    ok "real install: local Brewfile actually installed cowsay via brew"
    command -v cowsay >/dev/null && ok "real install: cowsay is executable on PATH" || bad "cowsay missing from PATH"
    brew uninstall -q cowsay >/dev/null 2>&1 && ok "cleanup: cowsay uninstalled" || bad "cleanup failed — run: brew uninstall cowsay"
  else
    bad "real install did not deliver cowsay"
  fi
fi

# --- Tier 3: wallpapers -------------------------------------------------
if command -v tmux >/dev/null; then
  S=bfcwp-$$
  tmux kill-session -t $S 2>/dev/null
  tmux new-session -d -s $S -x 100 -y 28
  tmux send-keys -t $S "HOME=$SANDBOX TERM=xterm-256color $BIN menu" Enter
  sleep 2.5
  cap=$(tmux capture-pane -t $S -p)
  echo "$cap" | grep -q "Sunset" && bad "Sunset offered on macOS (Windows/WSL only)" || ok "Sunset correctly hidden on macOS"
  tmux send-keys -t $S / ; tmux send-keys -t $S w a l l; sleep 0.4; tmux send-keys -t $S Enter; sleep 3
  cap=$(tmux capture-pane -t $S -p)
  if echo "$cap" | grep -qi "wallpapers"; then
    echo "$cap" | grep -qE "bluefin|aurora|bazzite" \
      && ok "wallpaper cask list renders on macOS" \
      || bad "wallpaper screen empty"
  else
    bad "wallpapers screen did not open"
  fi
  tmux send-keys -t $S Escape; sleep 0.3; tmux send-keys -t $S q; sleep 0.5
  tmux kill-session -t $S 2>/dev/null
fi

echo; echo "phase 2 results: $PASS ok, $FAIL failed"
[ $FAIL -eq 0 ]
