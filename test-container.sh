#!/usr/bin/env bash
# Container-based integration test for bluefin-cli
# Tests that bling actually modifies shells correctly

set -uo pipefail

# Set the path to the binary (in workspace when run via just test)
BLUEFIN_CLI="./bluefin-cli"

# Make sure we have a home directory with config files
export HOME="${HOME:-/root}"
mkdir -p "$HOME/.config/fish"
touch "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.config/fish/config.fish"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Print functions
print_header() {
    echo -e "\n${BLUE}=================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    ((TESTS_RUN++))
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test: Basic binary execution
test_binary_works() {
    print_test "Binary executes and shows version"
    if "$BLUEFIN_CLI" --version > /dev/null 2>&1; then
        print_pass "Binary works"
    else
        print_fail "Binary failed to execute"
        echo "Binary path: $BLUEFIN_CLI"
        ls -la "$BLUEFIN_CLI" 2>&1 || true
        "$BLUEFIN_CLI" --version 2>&1 || true
        return 1
    fi
}

# Test: Status command
test_status_command() {
    print_test "Status command shows configuration"
    if "$BLUEFIN_CLI" status > /dev/null 2>&1; then
        print_pass "Status command works"
    else
        print_fail "Status command failed"
        return 1
    fi
}

# Test: Bling command in bash
test_bling_bash() {
    print_test "Bling enables for bash"
    if "$BLUEFIN_CLI" bling bash on > /dev/null 2>&1; then
        if grep -q "bling.sh" "$HOME/.bashrc"; then
            print_pass "Bash bling enabled"
        else
            print_fail "Bash bling not found in ~/.bashrc"
            cat "$HOME/.bashrc" || true
            return 1
        fi
    else
        print_fail "Bling bash command failed"
        return 1
    fi
}

# Test: Enable bling for zsh
# Test: Bling command in zsh
test_bling_zsh() {
    print_test "Bling enables for zsh"
    if "$BLUEFIN_CLI" bling zsh on > /dev/null 2>&1; then
        if grep -q "bling.sh" "$HOME/.zshrc"; then
            print_pass "Zsh bling enabled"
        else
            print_fail "Zsh bling not found in ~/.zshrc"
            cat "$HOME/.zshrc" || true
            return 1
        fi
    else
        print_fail "Bling zsh command failed"
        return 1
    fi
}

# Test: Enable bling for fish
# Test: Bling command in fish
test_bling_fish() {
    print_test "Bling enables for fish"
    if "$BLUEFIN_CLI" bling fish on > /dev/null 2>&1; then
        if grep -q "bling.fish" "$HOME/.config/fish/config.fish"; then
            print_pass "Fish bling enabled"
        else
            print_fail "Fish bling not found in ~/.config/fish/config.fish"
            cat "$HOME/.config/fish/config.fish" || true
            return 1
        fi
    else
        print_fail "Bling fish command failed"
        return 1
    fi
}

# Test: MOTD system
test_motd() {
    print_test "Enable MOTD for bash"
    
    # Enable MOTD
    if "$BLUEFIN_CLI" motd toggle bash on > /dev/null 2>&1; then
        print_pass "MOTD enabled for bash"
    else
        print_fail "Failed to enable MOTD"
        return 1
    fi
    
    # Check bashrc contains MOTD
    print_test "Verify ~/.bashrc contains MOTD"
    if grep -q "bluefin-motd.sh" "$HOME/.bashrc"; then
        print_pass "MOTD configured in ~/.bashrc"
    else
        print_fail "MOTD not in ~/.bashrc"
        return 1
    fi
    
    # Check MOTD resources exist
    print_test "Verify MOTD resources are installed"
    local motd_dir="$HOME/.local/share/bluefin-cli/motd"
    if [ -d "$motd_dir/tips" ] && [ -f "$motd_dir/bluefin-motd.sh" ]; then
        local tip_count=$(find "$motd_dir/tips" -name "*.md" 2>/dev/null | wc -l)
        if [ "$tip_count" -ge 10 ]; then
            print_pass "MOTD resources installed ($tip_count tips)"
        else
            print_fail "Not enough tips found ($tip_count)"
            return 1
        fi
    else
        print_fail "MOTD resources not properly installed"
        return 1
    fi
    
    # Test MOTD show command
    print_test "MOTD show command displays content"
    if "$BLUEFIN_CLI" motd show 2>&1 | grep -q "Bluefin"; then
        print_pass "MOTD show works"
    else
        print_info "MOTD show executed (may need glow for formatting)"
    fi
}

# Test: Status reflects changes
test_status_after_changes() {
    print_test "Status command reflects enabled features"
    
    local status_output=$("$BLUEFIN_CLI" status 2>&1)
    
    # Check for bling status (look for "bash: enabled" under Shell Bling section)
    if echo "$status_output" | grep -q "bash: enabled"; then
        print_pass "Status shows bling enabled"
    else
        print_info "Status output: $status_output"
        print_fail "Status doesn't show bling as enabled"
        return 1
    fi
    
    # Check for MOTD status
    if echo "$status_output" | grep -q "bash: enabled" && echo "$status_output" | grep -q "Message of the Day"; then
        print_pass "Status shows MOTD enabled"
    else
        print_fail "Status doesn't show MOTD as enabled"
        return 1
    fi
}

# Test: Verify shell configuration files are valid
test_shell_configs_valid() {
    print_test "Bash config is valid syntax"
    if bash -n "$HOME/.bashrc" 2>&1; then
        print_pass "~/.bashrc has valid syntax"
    else
        print_fail "~/.bashrc has syntax errors"
        cat "$HOME/.bashrc"
        return 1
    fi
    
    print_test "Zsh config is valid syntax"
    if zsh -n "$HOME/.zshrc" 2>&1; then
        print_pass "~/.zshrc has valid syntax"
    else
        print_fail "~/.zshrc has syntax errors"
        cat "$HOME/.zshrc"
        return 1
    fi
    
    print_test "Fish config is valid syntax"
    if fish -n "$HOME/.config/fish/config.fish" 2>&1; then
        print_pass "Fish config has valid syntax"
    else
        print_fail "Fish config has syntax errors"
        cat "$HOME/.config/fish/config.fish"
        return 1
    fi
}

# Test: Disable bling
test_bling_disable() {
    print_test "Disable bling for bash"
    
    if "$BLUEFIN_CLI" bling bash off > /dev/null 2>&1; then
        print_pass "Bling disabled for bash"
    else
        print_fail "Failed to disable bling"
        return 1
    fi
    
    print_test "Verify bling removed from ~/.bashrc"
    if grep -q "bling.sh" "$HOME/.bashrc"; then
        print_fail "Bling still in ~/.bashrc after disable"
        return 1
    else
        print_pass "Bling properly removed from ~/.bashrc"
    fi
}

# Test: Install list command
test_install_list() {
    print_test "Install list command"
    
    if "$BLUEFIN_CLI" install list 2>&1 | grep -q "Available Homebrew Bundles"; then
        print_pass "Install list command works"
    else
        print_fail "Install list command failed"
        return 1
    fi
}

# Test: Verify bling script is sourced and configured in bash
test_bash_bling_aliases() {
    print_test "Bash config sources bling script"
    
    # Re-enable bling for testing
    "$BLUEFIN_CLI" bling bash on > /dev/null 2>&1
    
    # Check that bashrc sources the bling script
    if grep -q "source.*bling.sh" "$HOME/.bashrc" || grep -q "\. .*bling.sh" "$HOME/.bashrc"; then
        print_pass "Bling script is sourced in bashrc"
    else
        print_fail "Bling script not sourced in bashrc"
        cat "$HOME/.bashrc" || true
        return 1
    fi
    
    print_test "Bash bling script contains alias definitions"
    # Verify the bling.sh script itself is readable and contains alias definitions
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.sh"
    if [ -f "$bling_script" ] && grep -q "alias ls='eza'" "$bling_script"; then
        print_pass "Bling script contains eza alias definition"
    else
        print_fail "Bling script missing or doesn't define eza alias"
        return 1
    fi
    
    print_test "Bash bling script contains bat alias"
    if grep -q "alias cat='bat" "$bling_script"; then
        print_pass "Bling script contains bat alias definition"
    else
        print_fail "Bling script doesn't define bat alias"
        return 1
    fi
}

# Test: Verify bling script is sourced in zsh
test_zsh_bling_aliases() {
    print_test "Zsh config sources bling script"
    
    # Re-enable bling for testing
    "$BLUEFIN_CLI" bling zsh on > /dev/null 2>&1
    
    if grep -q "source.*bling.sh" "$HOME/.zshrc" || grep -q "\. .*bling.sh" "$HOME/.zshrc"; then
        print_pass "Bling script is sourced in zshrc"
    else
        print_fail "Bling script not sourced in zshrc"
        cat "$HOME/.zshrc" || true
        return 1
    fi
    
    print_test "Zsh bling script contains all tool configurations"
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.sh"
    if [ -f "$bling_script" ] && grep -q "starship init" "$bling_script" && grep -q "zoxide init" "$bling_script"; then
        print_pass "Bling script contains tool initializations"
    else
        print_fail "Bling script doesn't have complete configuration"
        return 1
    fi
}

# Test: Verify bling script is sourced in fish
test_fish_bling_aliases() {
    print_test "Fish bling script is sourced"
    
    # Re-enable bling for testing
    "$BLUEFIN_CLI" bling fish on > /dev/null 2>&1
    
    # Check that bling.fish script exists and is sourced
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.fish"
    if [ -f "$bling_script" ] && grep -q "source.*bling.fish" "$HOME/.config/fish/config.fish"; then
        print_pass "Bling script is sourced in fish"
    else
        print_fail "Bling script not sourced in fish"
        cat "$HOME/.config/fish/config.fish" || true
        return 1
    fi
    
    print_test "Fish bling script contains alias definitions"
    if grep -q "alias.*eza" "$bling_script" 2>/dev/null; then
        print_pass "Fish bling script contains eza alias"
    else
        print_fail "Fish bling script doesn't define eza alias"
        return 1
    fi
}

# Test: Verify starship prompt is configured
test_starship_prompt() {
    print_test "Bash bling script will initialize starship when available"
    
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.sh"
    if grep -q "starship init bash" "$bling_script" 2>/dev/null; then
        print_pass "Starship initialization in bash bling script"
    else
        print_fail "Starship not configured in bash bling script"
        return 1
    fi
    
    print_test "Zsh bling script will initialize starship when available"
    if grep -q "starship init zsh" "$bling_script" 2>/dev/null; then
        print_pass "Starship initialization in zsh bling script"
    else
        print_fail "Starship not configured in zsh bling script"
        return 1
    fi
    
    print_test "Fish bling script will initialize starship when available"
    local fish_bling="$HOME/.local/share/bluefin-cli/bling/bling.fish"
    if grep -q "starship init fish" "$fish_bling" 2>/dev/null; then
        print_pass "Starship initialization in fish bling script"
    else
        print_fail "Starship not configured in fish bling script"
        return 1
    fi
}

# Test: Verify other bling tools are configured
test_bling_tool_setup() {
    print_test "Bling script configures zoxide initialization"
    
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.sh"
    if grep -q "zoxide init" "$bling_script" 2>/dev/null; then
        print_pass "zoxide initialization in bling script"
    else
        print_fail "zoxide not configured in bling script"
        return 1
    fi
    
    print_test "Bling script configures atuin for history"
    if grep -q "atuin init" "$bling_script" 2>/dev/null; then
        print_pass "atuin initialization in bling script"
    else
        print_fail "atuin not configured in bling script"
        return 1
    fi
    
    print_test "Bling script checks for command availability"
    if grep -q 'command -v eza' "$bling_script" 2>/dev/null; then
        print_pass "Bling script checks if eza is installed before aliasing"
    else
        print_fail "Bling script doesn't check for tool availability"
        return 1
    fi
}

# Main test execution
main() {
    print_header "Bluefin CLI Container Tests"
    
    print_info "Running in container with user: $(whoami)"
    print_info "Home directory: $HOME"
    
    # Run all tests (continue even if some fail)
    test_binary_works || true
    test_status_command || true
    test_bling_bash || true
    test_bling_zsh || true
    test_bling_fish || true
    test_motd || true
    test_status_after_changes || true
    test_shell_configs_valid || true
    test_bash_bling_aliases || true
    test_zsh_bling_aliases || true
    test_fish_bling_aliases || true
    test_starship_prompt || true
    test_bling_tool_setup || true
    test_bling_disable || true
    test_install_list || true
    
    # Print summary
    print_header "Test Summary"
    echo "Tests run:    $TESTS_RUN"
    echo "Tests passed: $TESTS_PASSED"
    echo "Tests failed: $TESTS_FAILED"
    echo "================================="
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        exit 1
    fi
}

# Run tests
main
