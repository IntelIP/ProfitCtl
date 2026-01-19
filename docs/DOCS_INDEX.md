# profitctl Documentation Index

## Getting Started

1. **[QUICK_START.md](QUICK_START.md)** - Quick installation and testing guide
   - Fastest way to get up and running
   - Essential commands for testing
   - Common issues and solutions

2. **[INSTALL.md](INSTALL.md)** - Comprehensive installation guide
   - Detailed installation methods
   - Requirements and verification
   - Troubleshooting tips

## Understanding How It Works

3. **[HOW_IT_WORKS.md](HOW_IT_WORKS.md)** - Detailed explanation of how profitctl works
   - Step-by-step execution flow
   - Component explanations
   - Configuration structure
   - Statistical distributions
   - Complete example workflows

4. **[ARCHITECTURE.md](ARCHITECTURE.md)** - System architecture overview
   - Component architecture diagrams
   - Data flow visualization
   - Design patterns
   - Extension points
   - Performance considerations

## Examples

5. **[examples/valid_profit.yml](examples/valid_profit.yml)** - Example configuration
   - Demonstrates all features
   - Well-commented YAML
   - Best practices

6. **[test/fixtures/](test/fixtures/)** - Test configuration files
   - `valid_config.yml` - Comprehensive valid config
   - `minimal_config.yml` - Minimal required fields
   - `covenant_breach_config.yml` - Covenant violation example
   - `no_pricing_config.yml` - No pricing config

## Quick Reference

### Installation
```bash
git clone https://github.com/IntelIP/ProfitCtl.git
cd ProfitCtl
git checkout v0.0.1
go build -o profitctl .
```

### Basic Usage
```bash
# Run simulation
./profitctl simulate -f examples/valid_profit.yml

# JSON output
./profitctl simulate -f examples/valid_profit.yml --json

# Markdown output
./profitctl simulate -f examples/valid_profit.yml --markdown
```

### Exit Codes
- `0` - Success (all covenants passed)
- `1` - Covenant breach (fails CI build)
- `2` - Error (config parsing, file not found, etc.)

## Documentation Structure

```
profitctl/
├── QUICK_START.md      # Quick installation and testing
├── INSTALL.md          # Detailed installation guide
├── HOW_IT_WORKS.md     # Detailed how-it-works explanation
├── ARCHITECTURE.md     # System architecture overview
├── DOCS_INDEX.md       # This file
├── examples/           # Example configurations
└── test/fixtures/      # Test configurations
```

## Recommended Reading Order

1. **New users**: Start with [QUICK_START.md](QUICK_START.md)
2. **Installation issues**: See [INSTALL.md](INSTALL.md)
3. **Understanding the system**: Read [HOW_IT_WORKS.md](HOW_IT_WORKS.md)
4. **System design**: Check [ARCHITECTURE.md](ARCHITECTURE.md)
5. **Examples**: Look at [examples/valid_profit.yml](examples/valid_profit.yml)

## Need Help?

- Check [INSTALL.md](INSTALL.md) for installation issues
- Review [HOW_IT_WORKS.md](HOW_IT_WORKS.md) for execution flow questions
- See [ARCHITECTURE.md](ARCHITECTURE.md) for design questions
- Look at [examples/valid_profit.yml](examples/valid_profit.yml) for configuration examples
