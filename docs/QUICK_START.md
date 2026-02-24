# Quick Start

## 1. Simulate unit economics

```bash
profitctl simulate -f examples/valid_profit.yml
```

## 2. Validate config

```bash
profitctl validate -f examples/valid_profit.yml
```

## 3. Detect dependencies/services (LLM)

```bash
OPENROUTER_API_KEY=... profitctl detect --path . --out detect-report.json
```

## 4. Output modes

```bash
profitctl simulate -f examples/valid_profit.yml --json
profitctl simulate -f examples/valid_profit.yml --markdown
profitctl simulate -f examples/valid_profit.yml --quiet
```
