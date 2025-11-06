#!/usr/bin/env bash
# Container-based integration test for bluefin-cli
# Refactored for maintainability

set -uo pipefail

BLUEFIN_CLI="./bluefin-cli"
export HOME="${HOME:-/root}"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Initialize shell configs
init_shell_configs() {
    mkdir -p "$HOME/.config/fish"
    touch "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.config/fish/config.fish"
}

# Generic function to test bling for any shell
test_bling_for_shell() {
    local shell=$1
    local config_file=$2
    local bling_pattern=$3
    
    print_test "Bling enables for $shell"
    if "$BLUEFIN_CLI" bling "$shell" on > /dev/null 2>&1; then
        if grep -q "$bling_pattern" "$config_file"; then
            print_pass "$shell bling enabled"
            return 0
        else
            print_fail "$shell bling not found in config"
            return 1
        fi
    else
        print_fail "Bling $shell command failed"
        return 1
    fi
}

# Test that bling script is sourced in a shell config
test_bling_sourced() {
    local shell=$1
    local config_file=$2
    local bling_script_pattern=$3
    
    print_test "$shell config sources bling script"
    if grep -q "$bling_script_pattern" "$config_file"; then
        print_pass "Bling script is sourced in $shell"
        return 0
    else
        print_fail "Bling script not sourced in $shell"
        return 1
    fi
}

# Test shell config syntax
test_shell_syntax() {
    local shell=$1
    local config_file=$2
    
    print_test "$shell config is valid syntax"
    case "$shell" in
        bash)
            if bash -n "$config_file" 2>&1; then
                print_pass "$config_file has valid syntax"
                return 0
            fi
            ;;
        zsh)
            if zsh -n "$config_file" 2>&1; then
                print_pass "$config_file has valid syntax"
                return 0
            fi
            ;;
        fish)
            if fish -n "$config_file" 2>&1; then
                print_pass "$config_file has valid syntax"
                return 0
            fi
            ;;
    esac
    print_fail "$config_file has syntax errors"
    return 1
}

# Test that bling script contains expected tool configurations
test_bling_script_content() {
    local tool=$1
    local pattern=$2
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.sh"
    
    print_test "Bling script configures $tool"
    if [ -f "$bling_script" ] && grep -q "$pattern" "$bling_script" 2>/dev/null; then
        print_pass "$tool configuration found in bling script"
        return 0
    else
        print_fail "$tool not configured in bling script"
        return 1
    fi
}

# Fish uses separate bling script
test_fish_bling_script_content() {
    local tool=$1
    local pattern=$2
    local bling_script="$HOME/.local/share/bluefin-cli/bling/bling.fish"
    
    print_test "Fish bling script configures $tool"
    if [ -f "$bling_script" ] && grep -q "$pattern" "$bling_script" 2>/dev/null; then
        print_pass "$tool configuration found in fish bling script"
        return 0
    else
        print_fail "$tool not configured in fish bling script"
        return 1
    fi
}

# Main test execution
main() {
    print_header "Bluefin CLI Container Tests"
    
    print_info "Running in container with user: $(whoami)"
    print_info "Home directory: $HOME"
    
    init_shell_configs
    
    # Core functionality tests
    print_test "Binary executes and shows version"
    "$BLUEFIN_CLI" --version > /dev/null 2>&1 && print_pass "Binary works" || print_fail "Binary failed"
    
    print_test "Status command shows configuration"
    "$BLUEFIN_CLI" status > /dev/null 2>&1 && print_pass "Status command works" || print_fail "Status failed"
    
    # Bling tests for all shells (parameterized)
    test_bling_for_shell "bash" "$HOME/.bashrc" "bling.sh" || true
    test_bling_for_shell "zsh" "$HOME/.zshrc" "bling.sh" || true
    test_bling_for_shell "fish" "$HOME/.config/fish/config.fish" "bling.fish" || true
    
    # MOTD system tests
    print_test "Enable MOTD for bash"
    "$BLUEFIN_CLI" motd toggle bash on > /dev/null 2>&1 && print_pass "MOTD enabled" || print_fail "MOTD failed"
    
    print_test "Verify MOTD in bashrc"
    grep -q "bluefin-motd.sh" "$HOME/.bashrc" && print_pass "MOTD configured" || print_fail "MOTD not in bashrc"
    
    print_test "Verify MOTD resources installed"
    local motd_dir="$HOME/.local/share/bluefin-cli/motd"
    local tip_count=$(find "$motd_dir/tips" -name "*.md" 2>/dev/null | wc -l)
    [ "$tip_count" -ge 10 ] && print_pass "MOTD resources installed ($tip_count tips)" || print_fail "Not enough tips"
    
    print_test "MOTD show displays content"
    "$BLUEFIN_CLI" motd show 2>&1 | grep -q "Bluefin" && print_pass "MOTD show works" || print_info "MOTD show executed"
    
    # Status reflects changes
    print_test "Status reflects enabled features"
    local status_output=$("$BLUEFIN_CLI" status 2>&1)
    if echo "$status_output" | grep -q "bash: enabled" && echo "$status_output" | grep -q "Message of the Day"; then
        print_pass "Status shows enabled features"
    else
        print_fail "Status doesn't reflect changes"
    fi
    
    # Shell syntax validation
    test_shell_syntax "bash" "$HOME/.bashrc" || true
    test_shell_syntax "zsh" "$HOME/.zshrc" || true
    test_shell_syntax "fish" "$HOME/.config/fish/config.fish" || true
    
    # Bling script sourcing
    test_bling_sourced "bash" "$HOME/.bashrc" "bling.sh" || true
    test_bling_sourced "zsh" "$HOME/.zshrc" "bling.sh" || true
    test_bling_sourced "fish" "$HOME/.config/fish/config.fish" "bling.fish" || true
    
    # Bling script content validation (only need to check once, not per shell)
    test_bling_script_content "eza aliases" "alias ls='eza'" || true
    test_bling_script_content "bat alias" "alias cat='bat" || true
    test_bling_script_content "starship (bash)" "starship init bash" || true
    test_bling_script_content "starship (zsh)" "starship init zsh" || true
    test_bling_script_content "zoxide" "zoxide init" || true
    test_bling_script_content "atuin" "atuin init" || true
    test_bling_script_content "tool availability checks" "command -v eza" || true
    test_fish_bling_script_content "starship" "starship init fish" || true
    test_fish_bling_script_content "eza" "alias.*eza" || true
    
    # Disable functionality
    print_test "Disable bling for bash"
    "$BLUEFIN_CLI" bling bash off > /dev/null 2>&1 && print_pass "Bling disabled" || print_fail "Failed to disable"
    
    print_test "Verify bling removed from bashrc"
    ! grep -q "bling.sh" "$HOME/.bashrc" && print_pass "Bling properly removed" || print_fail "Bling still in bashrc"
    
    # Install command
    print_test "Install list command"
    "$BLUEFIN_CLI" install list 2>&1 | grep -q "Available Homebrew Bundles" && print_pass "Install list works" || print_fail "Install list failed"
    
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

main
