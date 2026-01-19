# Installation Guide for profitctl v0.0.1

## Installation Methods

### Method 1: Install from GitHub (Currently requires module path fix)

**Note:** The module path (`github.com/profitctl/profitctl`) doesn't match the repository (`github.com/IntelIP/ProfitCtl`), so `go install` won't work directly. Use Method 2 (build from source) until the module path is updated.

If the module path is fixed to match the repository, you could use:

```bash
go install github.com/IntelIP/ProfitCtl@v0.0.1
```

After installation, `profitctl` will be in your `$GOPATH/bin` or `$HOME/go/bin` directory. Make sure this directory is in your `PATH`:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Method 2: Build from Source (Recommended for v0.0.1)

1. Clone the repository:

```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
```

2. Checkout the v0.0.1 tag:

```bash
git checkout v0.0.1
```

3. Build the binary:

```bash
go build -o profitctl .
```

The binary will be created in the current directory as `profitctl`. You can move it to a location in your PATH:

```bash
# Move to a local bin directory
mkdir -p ~/.local/bin
mv profitctl ~/.local/bin/
export PATH=$PATH:~/.local/bin

# Or move to system-wide location (requires sudo)
sudo mv profitctl /usr/local/bin/
```

### Method 3: Download Pre-built Binary (Future)

When GitHub Releases are set up, you can download pre-built binaries for your platform from the releases page.

## Requirements

- Go 1.21 or later (for building from source)
- No external dependencies required at runtime (single binary)

## Verification

After installation, verify it works:

```bash
profitctl --help
```

You should see:

```
profitctl is a CLI tool that helps developers simulate, stress-test, 
and enforce profitability for software products.

Usage:
  profitctl [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize profitctl configuration from template
  simulate    Run cost and profit simulations
  validate    Validate profit configuration

Flags:
  -f, --file string   Config file (default "./profit.yml")
  -h, --help          help for profitctl
```

## Quick Start Testing

### 1. Test with Example Configuration

```bash
# Run simulation with example config
profitctl simulate -f examples/valid_profit.yml

# Output as JSON
profitctl simulate -f examples/valid_profit.yml --json

# Output as Markdown (for PR comments)
profitctl simulate -f examples/valid_profit.yml --markdown

# Verbose mode (shows scale simulation details)
profitctl simulate -f examples/valid_profit.yml --verbose

# Quiet mode (minimal output)
profitctl simulate -f examples/valid_profit.yml --quiet
```

### 2. Test Covenant Validation

The example config includes covenants that may fail. To see covenant breach behavior:

```bash
profitctl simulate -f examples/valid_profit.yml
# Exit code: 1 (covenant breach)

echo $?  # Should show 1
```

### 3. Test with Minimal Config

```bash
# Test with minimal configuration
profitctl simulate -f test/fixtures/minimal_config.yml
```

### 4. Test Config Parsing Errors

```bash
# Create an invalid config
cat > invalid.yml << EOF
simulation:
  base_users: 100
  invalid: [unclosed bracket
EOF

profitctl simulate -f invalid.yml
# Should show error and exit code 2
```

### 5. Test All Output Formats

```bash
# CLI format (default)
profitctl simulate -f examples/valid_profit.yml > output.txt

# JSON format (parse with jq)
profitctl simulate -f examples/valid_profit.yml --json | jq .

# Markdown format (for GitHub PRs)
profitctl simulate -f examples/valid_profit.yml --markdown > pr-comment.md
```

## Testing Checklist

- [ ] Installation successful (`profitctl --help` works)
- [ ] Simulate command runs with example config
- [ ] CLI output format displays correctly
- [ ] JSON output is valid (can be parsed with `jq`)
- [ ] Markdown output renders correctly (test in GitHub preview)
- [ ] Covenant breaches return exit code 1
- [ ] Config errors return exit code 2
- [ ] Verbose mode shows additional details
- [ ] Quiet mode produces minimal output

## Example Output

### CLI Format
```
=== profitctl Simulation Results ===

Scenario: 100 users
───────────────────────────────
Mean margin: 25.00%
Fixed COGS: $1100.00/month
Variable COGS: $1.0000/user
Cost per user: $12.00

Covenant Status: ✅ PASSED
```

### JSON Format
```json
{
  "scenario": {
    "users": 100,
    "months": 12
  },
  "costs": {
    "fixed": {...},
    "variable": {...}
  },
  "margin": {
    "gross": 25.0,
    "cost_per_user": 12.0
  },
  "covenants": {
    "passed": true,
    "violations": []
  }
}
```

### Markdown Format
```markdown
## profitctl Results

| Metric | Value |
|--------|-------|
| Users | 100 |
| Margin | 25.0% |
| Covenant Status | ✅ PASSED |

### Cost Breakdown
| Layer | Fixed | Variable |
|-------|-------|----------|
| Infrastructure | $700 | $50 |
...
```

## Troubleshooting

### Binary not found after installation

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH:

```bash
echo $PATH | grep -q "$(go env GOPATH)/bin" || export PATH=$PATH:$(go env GOPATH)/bin
```

Add to your shell profile (`.bashrc`, `.zshrc`, etc.) for persistence.

### Permission denied

Make the binary executable:

```bash
chmod +x profitctl
```

### Config file not found

Use the `-f` or `--file` flag to specify the config file:

```bash
profitctl simulate -f /path/to/your/profit.yml
```

## Next Steps

1. Create your own `profit.yml` configuration
2. Run simulations with your cost and pricing data
3. Set up covenants to enforce profitability constraints
4. Integrate into CI/CD pipelines using exit codes
5. Use JSON output for automated analysis
6. Use Markdown output for PR comments