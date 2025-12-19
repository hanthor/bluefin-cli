#!/usr/bin/env fish

# Source the configuration environment file if it exists
set BLING_ENV_FILE "$HOME/.local/share/bluefin-cli/bling/bling-env.fish"
if test -f "$BLING_ENV_FILE"
    source "$BLING_ENV_FILE"
end

# Default to enabled if variable is not set (backwards compatibility)
if not set -q BLING_ENABLE_EZA
    set BLING_ENABLE_EZA 1
end
if not set -q BLING_ENABLE_UGREP
    set BLING_ENABLE_UGREP 1
end
if not set -q BLING_ENABLE_BAT
    set BLING_ENABLE_BAT 1
end
if not set -q BLING_ENABLE_ATUIN
    set BLING_ENABLE_ATUIN 1
end
if not set -q BLING_ENABLE_STARSHIP
    set BLING_ENABLE_STARSHIP 1
end
if not set -q BLING_ENABLE_ZOXIDE
    set BLING_ENABLE_ZOXIDE 1
end

# ls aliases
if test "$BLING_ENABLE_EZA" -eq 1; and type -q eza
    alias ll='eza -l --icons=auto --group-directories-first'
    alias l.='eza -d .*'
    alias ls='eza'
    alias l1='eza -1'
end

# ugrep for grep
if test "$BLING_ENABLE_UGREP" -eq 1; and type -q ug
    alias grep='ug'
    alias egrep='ug -E'
    alias fgrep='ug -F'
    alias xzgrep='ug -z'
    alias xzegrep='ug -zE'
    alias xzfgrep='ug -zF'
end

# bat for cat
if test "$BLING_ENABLE_BAT" -eq 1
    alias cat='bat --style=plain --pager=never' 2>/dev/null
end

if status is-interactive
    # Initialize atuin before starship to ensure proper command history capture
    # Atuin allows these flags: "--disable-up-arrow" and/or "--disable-ctrl-r"
    # Use by setting a universal variable, e.g. set -U ATUIN_INIT_FLAGS "--disable-up-arrow"
    # Or set in config.fish before this file is sourced
    if test "$BLING_ENABLE_ATUIN" -eq 1; and type -q atuin
        atuin init fish $ATUIN_INIT_FLAGS | source
    end

    if test "$BLING_ENABLE_STARSHIP" -eq 1; and type -q starship
        starship init fish | source
    end

    if test "$BLING_ENABLE_ZOXIDE" -eq 1; and type -q zoxide
        zoxide init fish | source
    end
end
