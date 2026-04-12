# Recurring Covenant Safety: Safe vs Breach

Command:

```bash
profitctl compare benchmark_scenarios/hybrid_operating_safe.yml benchmark_scenarios/hybrid_operating_breach.yml --markdown
```

Question:

What does a clearly covenant-safe hybrid contract look like against one that fails recurring operating health?

Result:

`hybrid_operating_breach` fails because operating margin drops well below the configured covenant threshold. This is the benchmark to show when the goal is risk detection, not just revenue comparison.

Note:

This command exits non-zero because the second scenario breaches its covenant. That is expected and useful in CI.

Output:

```markdown
## profitctl Scenario Comparison

Baseline: **hybrid_operating_safe**

| Scenario | Mode | Revenue | Recurring Revenue | Payment Fees | Booked Margin | Operating Margin | Cost/User | Covenants |
|----------|------|---------|-------------------|--------------|---------------|------------------|-----------|-----------|
| hybrid_operating_safe | hybrid | $2000.00 | $2000.00 | $0.00 | 95.00% | 95.00% | $2.00 | PASS |
| hybrid_operating_breach | hybrid | $250.00 | $250.00 | $0.00 | -100.00% | -100.00% | $10.00 | FAIL (1) |

### Delta vs Baseline

| Scenario | Revenue Delta | Booked Margin Delta | Operating Margin Delta | Cost/User Delta |
|----------|---------------|---------------------|------------------------|-----------------|
| hybrid_operating_breach | $-1750.00 | -195.00 pts | -195.00 pts | $+8.00 |

### Failing Scenarios

- **hybrid_operating_breach**: Operating margin must be >= 30%

### Leaders

- Highest revenue: **hybrid_operating_safe**
- Highest operating margin: **hybrid_operating_safe**
- Lowest cost/user: **hybrid_operating_safe**
- Best covenant health: **hybrid_operating_safe**
```
