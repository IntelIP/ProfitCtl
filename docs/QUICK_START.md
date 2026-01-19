# Quick Start Guide - Testing profitctl

## Prerequisites

- Go 1.21+ installed
- `profitctl` binary built or installed

## Installation

### From GitHub (Build from Source - Recommended)

**Note:** Due to module path mismatch, `go install` doesn't work. Build from source instead:

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
git checkout v0.0.1
go build -o profitctl .
```

Add to PATH (if needed):
```bash
export PATH=$PATH:$(pwd)
# Or move to a standard location
mkdir -p ~/.local/bin && mv profitctl ~/.local/bin && export PATH=$PATH:~/.local/bin
```

### Build from Source

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
git checkout v0.0.1
go build -o profitctl .
```

## Quick Test

### 1. Verify Installation

```bash
profitctl --help
```

### 2. Run Simulation

```bash
# Using the example config
./profitctl simulate -f examples/valid_profit.yml

# Or if installed globally:
profitctl simulate -f examples/valid_profit.yml
```

### 3. Test Different Output Formats

```bash
# CLI format (default)
./profitctl simulate -f examples/valid_profit.yml

# JSON format
./profitctl simulate -f examples/valid_profit.yml --json | jq .

# Markdown format
./profitctl simulate -f examples/valid_profit.yml --markdown
```

### 4. Test Covenant Validation

```bash
# Run simulation
./profitctl simulate -f examples/valid_profit.yml

# Check exit code (should be 1 if covenants fail, 0 if pass)
echo "Exit code: $?"
```

### 5. Test Verbose Mode

```bash
./profitctl simulate -f examples/valid_profit.yml --verbose
```

## What to Look For

✅ **Successful Output Should Show:**
- Scenario summary (users, margin)
- Fixed COGS breakdown by layer
- Variable COGS per user
- Covenant status (PASSED or FAILED)
- Stress test results (if verbose)

✅ **JSON Output Should:**
- Be valid JSON (parseable with `jq`)
- Contain all expected fields (scenario, costs, revenue, margin, covenants)
- Be properly indented

✅ **Markdown Output Should:**
- Render as tables in GitHub
- Include all summary metrics
- Show covenant violations (if any)

## Testing Exit Codes

```bash
# Success (covenants pass)
./profitctl simulate -f examples/valid_profit.yml
echo $?  # Should be 0

# Covenant breach
./profitctl simulate -f test/fixtures/covenant_breach_config.yml
echo $?  # Should be 1

# Config error
./profitctl simulate -f /nonexistent/file.yml
echo $?  # Should be 2
```

## Common Issues

**Binary not found:** Add `$GOPATH/bin` to PATH
**Config errors:** Check YAML syntax and required fields
**Permission denied:** Run `chmod +x profitctl`
