#!/usr/bin/env sh

# Check if bling has already been sourced so that we dont break atuin. https://github.com/atuinsh/atuin/issues/380#issuecomment-1594014644
[ "${BLING_SOURCED:-0}" -eq 1 ] && return 
BLING_SOURCED=1

# Source the configuration environment file if it exists
BLING_ENV_FILE="${HOME}/.local/share/bluefin-cli/bling/bling-env.sh"
[ -f "$BLING_ENV_FILE" ] && . "$BLING_ENV_FILE"

# Default to enabled if variable is not set (backwards compatibility)
: "${BLING_ENABLE_EZA:=1}"
: "${BLING_ENABLE_UGREP:=1}"
: "${BLING_ENABLE_BAT:=1}"
: "${BLING_ENABLE_ATUIN:=1}"
: "${BLING_ENABLE_STARSHIP:=1}"
: "${BLING_ENABLE_ZOXIDE:=1}"

# eza
# ls aliases
if [ "$BLING_ENABLE_EZA" -eq 1 ] && [ "$(command -v eza)" ]; then
    alias ll='eza -l --icons=auto --group-directories-first'
    alias l.='eza -d .*'
    alias ls='eza'
    alias l1='eza -1'
fi

# ugrep 
# for grep
if [ "$BLING_ENABLE_UGREP" -eq 1 ] && [ "$(command -v ug)" ]; then
    alias grep='ug'
    alias egrep='ug -E'
    alias fgrep='ug -F'
    alias xzgrep='ug -z'
    alias xzegrep='ug -zE'
    alias xzfgrep='ug -zF'
fi

# bat 
# for cat
if [ "$BLING_ENABLE_BAT" -eq 1 ]; then
    alias cat='bat --style=plain --pager=never' 2>/dev/null
fi

HOMEBREW_PREFIX="${HOMEBREW_PREFIX:-/home/linuxbrew/.linuxbrew}"

# set ATUIN_INIT_FLAGS in your ~/.bashrc before ublue-bling is sourced.
# Atuin allows these flags: "--disable-up-arrow" and/or "--disable-ctrl-r"
ATUIN_INIT_FLAGS=${ATUIN_INIT_FLAGS:-""}

if [ "$(basename "$SHELL")" = "bash" ]; then
    [ -f "/etc/profile.d/bash-preexec.sh" ] && . "/etc/profile.d/bash-preexec.sh"
    [ -f "/usr/share/bash-prexec" ] && . "/usr/share/bash-prexec"
    [ -f "/usr/share/bash-prexec.sh" ] && . "/usr/share/bash-prexec.sh"
    [ -f "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh" ] && . "${HOMEBREW_PREFIX}/etc/profile.d/bash-preexec.sh"
    
    # Initialize atuin before starship to ensure proper command history capture
    # See: https://github.com/atuinsh/atuin/issues/2804 
    if [ "$BLING_ENABLE_ATUIN" -eq 1 ] && [ "$(command -v atuin)" ]; then
        eval "$(atuin init bash ${ATUIN_INIT_FLAGS})"
    fi

    if [ "$BLING_ENABLE_STARSHIP" -eq 1 ] && [ "$(command -v starship)" ]; then
        eval "$(starship init bash)"
    fi
    
    if [ "$BLING_ENABLE_ZOXIDE" -eq 1 ] && [ "$(command -v zoxide)" ]; then
        eval "$(zoxide init bash)"
    fi
elif [ "$(basename "$SHELL")" = "zsh" ]; then
    # Initialize atuin before starship to ensure proper command history capture
    if [ "$BLING_ENABLE_ATUIN" -eq 1 ] && [ "$(command -v atuin)" ]; then
        eval "$(atuin init zsh ${ATUIN_INIT_FLAGS})"
    fi

    if [ "$BLING_ENABLE_STARSHIP" -eq 1 ] && [ "$(command -v starship)" ]; then
        eval "$(starship init zsh)"
    fi

    if [ "$BLING_ENABLE_ZOXIDE" -eq 1 ] && [ "$(command -v zoxide)" ]; then
        eval "$(zoxide init zsh)"
    fi
fi
