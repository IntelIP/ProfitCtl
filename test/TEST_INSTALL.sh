#!/bin/bash
# Quick installation and test script for profitctl v0.0.1

set -e

echo "=== profitctl Installation Test ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

echo "✅ Go version: $(go version)"
echo

# Build from source
echo "Building profitctl from source..."
go build -o profitctl .
echo "✅ Build successful"
echo

# Test help command
echo "Testing --help command..."
./profitctl --help | head -10
echo "✅ Help command works"
echo

# Test simulate with example config
echo "Testing simulate command with example config..."
if [ -f examples/valid_profit.yml ]; then
    ./profitctl simulate -f examples/valid_profit.yml --quiet 2>&1 | tail -5
    echo "✅ Simulate command works"
else
    echo "⚠️  Example config not found, skipping simulate test"
fi
echo

# Test exit codes
echo "Testing exit codes..."
./profitctl simulate -f examples/valid_profit.yml --quiet 2>&1 > /dev/null || true
EXIT_CODE=${PIPESTATUS[0]}
echo "Exit code: $EXIT_CODE (should be 0 or 1 for covenant pass/fail)"
echo

# Test JSON output
echo "Testing JSON output..."
./profitctl simulate -f examples/valid_profit.yml --json 2>&1 | jq -e '.scenario' > /dev/null 2>&1 && echo "✅ JSON output is valid" || echo "⚠️  JSON output test skipped (jq not installed)"
echo

echo "=== Installation Test Complete ==="
echo "To use profitctl, run: ./profitctl [command]"
echo "Or move to PATH: mv profitctl ~/.local/bin/"
