# Benchmark Scenarios: Astro VPS vs Next.js on Vercel

This folder contains ProfitCtl configs to compare a simple Astro deployment on a VPS against a Next.js deployment on Vercel Pro.

All inputs are based on public pricing as of Jan 28, 2026, plus Fermi modeling for traffic assumptions.

## Modeling assumptions

Traffic model:
- Monthly visits (modeled as "users"): 50,000
- Pages per visit: 1.2
- Page weight: 1.5 MB
- Bandwidth per visit: 1.2 * 1.5 MB = 1.8 MB
- GB per visit (approx): 1.8 / 1024 = 0.0018 GB

Simulation:
- 12 months
- Growth factor: 1.15 (15% MoM)
- Monte Carlo iterations: 10,000
- Variability: 25% stddev around mean

Important limitation:
ProfitCtl does not model "included bandwidth" tiers directly. These configs apply the provider overage rate to all outbound traffic (conservative/worst-case).

If your usage stays within included transfer:
- Set bandwidth `cost_per_unit` to 0
  OR
- Reduce the effective GB/visit based on your allowance.

## Scenarios

Astro on VPS:
- `benchmark_scenarios/astro_vps_digitalocean.yml`
- `benchmark_scenarios/astro_vps_linode.yml`
- `benchmark_scenarios/astro_vps_vultr.yml`
- `benchmark_scenarios/astro_vps_hetzner_eu.yml` (EU pricing)

Next.js on Vercel:
- `benchmark_scenarios/nextjs_vercel_pro.yml`

## Run commands

```bash
./profitctl simulate -f benchmark_scenarios/astro_vps_digitalocean.yml
./profitctl simulate -f benchmark_scenarios/astro_vps_linode.yml
./profitctl simulate -f benchmark_scenarios/astro_vps_vultr.yml
./profitctl simulate -f benchmark_scenarios/astro_vps_hetzner_eu.yml
./profitctl simulate -f benchmark_scenarios/nextjs_vercel_pro.yml
```

## Scenario performance (local run)

Measured on Jan 28, 2026 with 10,000 Monte Carlo iterations.

Each scenario completed in ~6.4–6.6s wall time with ~8 MB peak RSS.

Per-scenario highlights:
- Astro + DigitalOcean: fixed $6.00/mo, runtime ~6.5s
- Astro + Hetzner (EU): fixed $3.79/mo, runtime ~6.4s
- Astro + Linode: fixed $5.00/mo, runtime ~6.6s
- Astro + Vultr: fixed $5.00/mo, runtime ~6.5s
- Next.js + Vercel Pro: fixed $20.00/mo, runtime ~6.5s

Note: CLI output rounds costs to 2 decimals. With the current traffic assumptions, variable costs are very small and may show as $0.00 in CLI output.

## Pricing sources (checked Jan 28, 2026)

DigitalOcean:
- Droplet pricing table (Basic 1GB $6/mo, 1TB transfer)
- Bandwidth overage: $0.01 per GiB

Linode:
- Shared CPU pricing table (Nanode 1GB $5/mo, 1TB transfer)
- Egress overage: $0.005 per GB

Vultr:
- Cloud Compute pricing table (Regular Performance 1GB $5/mo, 1TB transfer)
- Bandwidth overage: $0.01 per GB

Hetzner (EU):
- CX shared vCPU plan pricing (CX22 €3.79/mo, 20TB traffic)
- Overage: €1 ($1.20) per TB

Vercel:
- Pro platform fee: $20/mo (includes $20 usage credit, 1 TB Fast Data Transfer)
- Fast Data Transfer overage (US region): $0.15 per GB
