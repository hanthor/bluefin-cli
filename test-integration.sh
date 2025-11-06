#!/usr/bin/env bash
# Integration test for bluefin-cli
# This script runs comprehensive tests in a container environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Print functions
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

# Setup test environment
setup() {
    print_info "Setting up test environment..."
    export TEST_HOME="/tmp/bluefin-cli-test-$$"
    export HOME="$TEST_HOME"
    mkdir -p "$TEST_HOME"
    
    # Build the binary
    print_info "Building bluefin-cli..."
    if go build -o bluefin-cli; then
        print_pass "Build successful"
    else
        print_fail "Build failed"
        exit 1
    fi
}

# Cleanup test environment
cleanup() {
    print_info "Cleaning up test environment..."
    rm -rf "$TEST_HOME"
    rm -f bluefin-cli
}

# Test: Binary executes and shows help
test_help() {
    print_test "Binary shows help output"
    if ./bluefin-cli --help > /dev/null 2>&1; then
        print_pass "Help command works"
    else
        print_fail "Help command failed"
    fi
}

# Test: Version command
test_version() {
    print_test "Version command works"
    if ./bluefin-cli --version | grep -q "bluefin-cli version"; then
        print_pass "Version command works"
    else
        print_fail "Version command failed"
    fi
}

# Test: Status command
test_status() {
    print_test "Status command executes"
    if ./bluefin-cli status > /dev/null 2>&1; then
        print_pass "Status command works"
    else
        print_fail "Status command failed"
    fi
}

# Test: Enable bling for bash
test_bling_enable() {
    print_test "Enable bling for bash"
    
    if ./bluefin-cli bling bash on > /dev/null 2>&1; then
        # Check if .bashrc was modified
        if [ -f "$HOME/.bashrc" ] && grep -q "bluefin-cli bling" "$HOME/.bashrc"; then
            print_pass "Bling enabled for bash"
        else
            print_fail "Bling marker not found in .bashrc"
        fi
    else
        print_fail "Failed to enable bling"
    fi
}

# Test: Bling script exists
test_bling_script() {
    print_test "Bling script is installed"
    
    if [ -f "$HOME/.local/share/bluefin-cli/bling/bling.sh" ]; then
        if [ -x "$HOME/.local/share/bluefin-cli/bling/bling.sh" ]; then
            print_pass "Bling script exists and is executable"
        else
            print_fail "Bling script is not executable"
        fi
    else
        print_fail "Bling script not found"
    fi
}

# Test: Disable bling
test_bling_disable() {
    print_test "Disable bling for bash"
    
    if ./bluefin-cli bling bash off > /dev/null 2>&1; then
        # Check if marker was removed
        if [ -f "$HOME/.bashrc" ] && ! grep -q "bluefin-cli bling" "$HOME/.bashrc"; then
            print_pass "Bling disabled for bash"
        else
            print_fail "Bling marker still in .bashrc"
        fi
    else
        print_fail "Failed to disable bling"
    fi
}

# Test: MOTD setup
test_motd_enable() {
    print_test "Enable MOTD for bash"
    
    if ./bluefin-cli motd toggle bash on > /dev/null 2>&1; then
        # Check if .bashrc was modified
        if [ -f "$HOME/.bashrc" ] && grep -q "bluefin-cli motd" "$HOME/.bashrc"; then
            print_pass "MOTD enabled for bash"
        else
            print_fail "MOTD marker not found in .bashrc"
        fi
    else
        print_fail "Failed to enable MOTD"
    fi
}

# Test: MOTD show
test_motd_show() {
    print_test "MOTD show command"
    
    if ./bluefin-cli motd show > /dev/null 2>&1; then
        print_pass "MOTD show command works"
    else
        print_fail "MOTD show command failed"
    fi
}

# Test: MOTD resources
test_motd_resources() {
    print_test "MOTD resources are installed"
    
    local motd_dir="$HOME/.local/share/bluefin-cli/motd"
    
    if [ -d "$motd_dir/tips" ] && [ -f "$motd_dir/bluefin-motd.sh" ]; then
        # Check if tips exist
        local tip_count=$(find "$motd_dir/tips" -name "*.md" | wc -l)
        if [ "$tip_count" -gt 0 ]; then
            print_pass "MOTD resources installed (${tip_count} tips)"
        else
            print_fail "No tips found"
        fi
    else
        print_fail "MOTD resources not properly installed"
    fi
}

# Test: Install list command
test_install_list() {
    print_test "Install list command"
    
    if ./bluefin-cli install list 2>&1 | grep -q "Available Homebrew Bundles"; then
        print_pass "Install list command works"
    else
        print_fail "Install list command failed"
    fi
}

# Test: Brewfile init
test_brewfile_init() {
    print_test "Brewfile init command"
    
    cd "$TEST_HOME"
    if ./bluefin-cli brewfile init > /dev/null 2>&1; then
        if [ -f "Brewfile" ]; then
            print_pass "Brewfile created"
        else
            print_fail "Brewfile not created"
        fi
    else
        print_fail "Brewfile init failed"
    fi
    cd - > /dev/null
}

# Test: Brewfile add
test_brewfile_add() {
    print_test "Brewfile add command"
    
    cd "$TEST_HOME"
    if ./bluefin-cli brewfile add wget > /dev/null 2>&1; then
        if grep -q "wget" "Brewfile"; then
            print_pass "Package added to Brewfile"
        else
            print_fail "Package not found in Brewfile"
        fi
    else
        print_fail "Brewfile add failed"
    fi
    cd - > /dev/null
}

# Test: Multiple shells
test_multiple_shells() {
    print_test "Enable bling for multiple shells"
    
    local success=true
    
    for shell in bash zsh fish; do
        if ! ./bluefin-cli bling "$shell" on > /dev/null 2>&1; then
            success=false
            print_fail "Failed to enable bling for $shell"
        fi
    done
    
    if [ "$success" = true ]; then
        print_pass "Bling enabled for all shells"
    fi
}

# Test: Status shows enabled features
test_status_shows_enabled() {
    print_test "Status shows enabled features"
    
    local output=$(./bluefin-cli status 2>&1)
    
    if echo "$output" | grep -q "bash: enabled" && \
       echo "$output" | grep -q "zsh: enabled" && \
       echo "$output" | grep -q "fish: enabled"; then
        print_pass "Status correctly shows enabled features"
    else
        print_fail "Status doesn't show all enabled features"
    fi
}

# Test: Idempotency - enable twice
test_idempotency() {
    print_test "Idempotency - enable bling twice"
    
    # Fresh start
    rm -f "$HOME/.bashrc"
    
    # Enable first time
    ./bluefin-cli bling bash on > /dev/null 2>&1
    
    # Count markers
    local count1=$(grep -c "bluefin-cli bling" "$HOME/.bashrc" || echo 0)
    
    # Enable second time
    ./bluefin-cli bling bash on > /dev/null 2>&1
    
    # Count markers again
    local count2=$(grep -c "bluefin-cli bling" "$HOME/.bashrc" || echo 0)
    
    if [ "$count1" -eq "$count2" ] && [ "$count1" -eq 1 ]; then
        print_pass "Bling enable is idempotent"
    else
        print_fail "Bling enable created duplicate entries ($count1 vs $count2)"
    fi
}

# Print summary
print_summary() {
    echo ""
    echo "================================="
    echo "Test Summary"
    echo "================================="
    echo "Tests run:    $TESTS_RUN"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
    echo "================================="
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        return 1
    fi
}

# Main test execution
main() {
    echo "================================="
    echo "Bluefin CLI Integration Tests"
    echo "================================="
    echo ""
    
    setup
    
    # Run all tests
    test_help
    test_version
    test_status
    test_bling_enable
    test_bling_script
    test_bling_disable
    test_motd_enable
    test_motd_show
    test_motd_resources
    test_install_list
    test_brewfile_init
    test_brewfile_add
    test_multiple_shells
    test_status_shows_enabled
    test_idempotency
    
    cleanup
    
    print_summary
}

# Run main with error handling
if main; then
    exit 0
else
    exit 1
fi
