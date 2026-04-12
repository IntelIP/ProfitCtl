# Hybrid Contracts: Steady-State vs Pilot

Command:

```bash
profitctl compare benchmark_scenarios/hybrid_steady_contract.yml benchmark_scenarios/hybrid_pilot_contract.yml --markdown
```

Question:

Does pilot revenue improve the contract, or does it only flatter booked economics while steady-state health gets worse?

Result:

The pilot contract increases total revenue, but it weakens operating margin and raises cost per user. This is the right benchmark to show when a team needs to separate one-time onboarding revenue from recurring business health.

Output:

```markdown
## profitctl Scenario Comparison

Baseline: **hybrid_steady_contract**

| Scenario | Mode | Revenue | Recurring Revenue | Payment Fees | Booked Margin | Operating Margin | Cost/User | Covenants |
|----------|------|---------|-------------------|--------------|---------------|------------------|-----------|-----------|
| hybrid_steady_contract | hybrid | $2000.00 | $2000.00 | $0.00 | 95.00% | 95.00% | $2.00 | PASS |
| hybrid_pilot_contract | hybrid | $6500.00 | $2000.00 | $200.50 | 95.38% | 91.50% | $3.40 | PASS |

### Delta vs Baseline

| Scenario | Revenue Delta | Booked Margin Delta | Operating Margin Delta | Cost/User Delta |
|----------|---------------|---------------------|------------------------|-----------------|
| hybrid_pilot_contract | $+4500.00 | +0.38 pts | -3.50 pts | $+1.40 |

### Leaders

- Highest revenue: **hybrid_pilot_contract**
- Highest operating margin: **hybrid_steady_contract**
- Lowest cost/user: **hybrid_steady_contract**
- Best covenant health: **hybrid_pilot_contract**
```
