# Open-Core Pricing: Tiered vs Mix

Command:

```bash
profitctl compare benchmark_scenarios/open_core_tiered.yml benchmark_scenarios/open_core_mix.yml --markdown
```

Question:

Should the open-core packaging lead with cumulative tiered pricing or an explicit plan-mix model?

Result:

`open_core_tiered` wins on revenue, operating margin, and cost per user in the current benchmark assumptions. The `mix` configuration still passes covenants, but it gives up too much revenue without reducing cost enough to justify the trade.

Output:

```markdown
## profitctl Scenario Comparison

Baseline: **open_core_tiered**

| Scenario | Mode | Revenue | Recurring Revenue | Payment Fees | Booked Margin | Operating Margin | Cost/User | Covenants |
|----------|------|---------|-------------------|--------------|---------------|------------------|-----------|-----------|
| open_core_tiered | tiered | $2425.00 | $2425.00 | $0.00 | 75.26% | 75.26% | $6.00 | PASS |
| open_core_mix | mix | $880.00 | $880.00 | $35.66 | 27.77% | 27.77% | $6.36 | PASS |

### Delta vs Baseline

| Scenario | Revenue Delta | Booked Margin Delta | Operating Margin Delta | Cost/User Delta |
|----------|---------------|---------------------|------------------------|-----------------|
| open_core_mix | $-1545.00 | -47.49 pts | -47.49 pts | $+0.36 |

### Leaders

- Highest revenue: **open_core_tiered**
- Highest operating margin: **open_core_tiered**
- Lowest cost/user: **open_core_tiered**
- Best covenant health: **open_core_tiered**
```
