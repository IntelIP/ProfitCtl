# IntelIP Tooling Cost Inventory

This inventory maps IntelIP's currently observed stack to the first-pass cost model used in the ProfitCtl scenario pack.

It is intentionally split into:

- product COGS and delivery costs that should influence pricing
- business tooling overhead that matters for operating margin, but not necessarily gross margin
- a first-pass economics-layer split across `delivery`, `productization`, and `adoption`

These entries are based on repository configuration, local secret inventory, and public vendor pricing as of April 13, 2026. They are not invoice exports.

## Verified local inventory on April 13, 2026

What is verified from repo code and local Doppler secret-name inventory:

- `intelip_frontend` (`stg` and `prd`) carries Clerk, Cloudflare, PostHog, and Stripe billing inputs. There is no active R2, OpenPanel, or Faro secret inventory left in the frontend config.
- `web-app-2.0` no longer contains active R2 upload code, OpenPanel fan-out, or Grafana Faro browser instrumentation.
- `intelip_api` `dev` still carries `REDIS_URL`, `CHROMA_*`, `OPENROUTER_API_KEY`, `COMPOSIO_DEV_API_KEY`, `EXA_API_KEY`, and a stale `OPENPANEL_CLIENT_*` pair.
- `intelip_api` `prd` is now richer than the initial minimal inventory: the missing Chroma, Composio, and runtime-auth secrets were backfilled during the Apr 13 backend bootstrap. The original visible minimal set (`DB_URL`, `STRIPE_KEY`, `FEATURE_FLAGS`, `LOGGING`, `PRIVATE_KEY`) should no longer be treated as the full production backend picture.
- GCP-side backend reality is now active rather than partial: both `intelip-agentos-staging` and `intelip-agentos-prod` exist as Cloud Run services in `intelip-prod-2025`.
- The landing-page repo still documents Hostinger/VPS deployment, not Cloudflare Pages, so the current modeled landing-hosting line should remain Hostinger/VPS unless infra changed outside the repo after April 13, 2026.

What is still blocked on live auth or network access:

- Neon, PostHog, Composio, and Chroma current spend still need dashboard or API access; the repo only proves they are wired.

## Secret ownership rule used for this inventory

The modeled stack now assumes a split secret-management model rather than one
shared system:

- frontend runtime:
  - `Doppler` is the operator-facing source of truth
  - `Cloudflare` is the runtime mirror
- backend Cloud Run runtime:
  - `Doppler` or the vendor dashboard may be the editing surface
  - `Google Secret Manager` is the runtime mirror
  - `Cloud Run` is the final consumer

This matters for cost modeling because local Doppler visibility alone is not
proof that a backend production cost surface is fully active. For Condere-style
runtime assumptions, production readiness should be judged from Cloud Run plus
Google Secret Manager, not from Doppler alone.

## Live Stripe test-mode findings on April 13, 2026

After completing Stripe CLI login locally, the current test-mode IntelIP billing catalog is now verified:

- `web-app-2.0` expects four env-backed Stripe price IDs:
  - `STRIPE_PRICE_STARTER_MONTHLY`
  - `STRIPE_PRICE_STARTER_ANNUAL`
  - `STRIPE_PRICE_PRO_MONTHLY`
  - `STRIPE_PRICE_PRO_ANNUAL`
- The active Stripe test-mode catalog matches that shape exactly:
  - Starter Monthly: `$49` via lookup key `intelip_starter_monthly`
  - Starter Annual: `$490` via lookup key `intelip_starter_annual`
  - Pro Monthly: `$99` via lookup key `intelip_pro_monthly`
  - Pro Annual: `$990` via lookup key `intelip_pro_annual`
- Recent test-mode billing activity is almost entirely Starter Monthly:
  - 2 active Starter Monthly subscriptions
  - 2 recent paid Starter Monthly invoices at `$49`
  - no recent IntelIP Pro subscription activity in the sampled test-mode data
- Recent observed Stripe processing fees on Starter Monthly charges are:
  - `$49.00` charge -> `$1.72` fee -> `$47.28` net
  - `$49.00` charge -> `$1.72` fee -> `$47.28` net
- There was also a stray Stripe CLI-created `myproduct` subscription at `$15/month` with a `$0.74` fee. That object was canceled during the Apr 13 cleanup, and its active recurring price / product were archived so it no longer pollutes the active IntelIP test catalog.
- Live-mode Stripe inspection through the regular Stripe CLI auth shows a different production-side shape:
  - sampled active live subscriptions are on `Professional Plan` products with `Autumn Price (Fixed)` recurring prices at `$99/month`
  - sampled live subscriptions currently carry a `launch` `100%` forever coupon, so sampled live invoices are `$0`
  - sampled live prices do not use the `intelip_starter_*` / `intelip_pro_*` lookup-key pattern the app expects in test mode

Important modeling implication:

- We can now ground IntelIP monthly Stripe payment-fee assumptions against an observed Starter Monthly test fee of roughly `3.5%` effective processing cost.
- The current scenario pack still treats Stripe fees as a per-paid-user reserve, not as a per-invoice or per-workspace charge. That is good enough for current seat-mix scenarios, but it is still an approximation for workspace-minimum and paid-pilot motions.
- `intelip_frontend/prd` now carries the production Stripe billing env set and can sync Cloudflare production billing secrets directly.
- The live Stripe account now has a canonical IntelIP billing catalog matching the app contract, while active legacy pro-monthly subscriptions are carried through `STRIPE_PRICE_PRO_MONTHLY_ALIASES` during cutover.

## Live infrastructure findings on April 13, 2026

After refreshing `gcloud` auth locally, the current live Cloud Run footprint we can verify is:

- project: `intelip-prod-2025`
- live service observed: `intelip-agentos-staging`
- region: `us-east1`
- no Cloud Run services observed in `us-central1`
- no Cloud Run jobs observed in `us-east1`
- service runtime shape:
  - image: `us-east1-docker.pkg.dev/intelip-prod-2025/agentos/condere-api:staging-d84d547ac6c0`
  - CPU: `1`
  - memory: `2Gi`
  - concurrency: `10`
  - request timeout: `900s`
  - autoscaling:
    - service max scale annotation: `20`
    - active revision max scale annotation: `3`

This is enough to tighten the current backend runtime assumption:

- IntelIP is not currently carrying a broad multi-service Cloud Run fleet in the verified project state.
- The active backend runtime we can prove is a single staging Condere service with modest limits, not a large always-on fleet.
- Until we can extract billing or usage metrics, the Cloud Run line should remain conservative but should not assume multiple large production services are already running.

Additional live blockers:

- The production backend `DB_URL` resolves to a private `10.x` host and times out from this machine even after normalizing the stored `psql://` scheme to `postgresql://`. That means Neon footprint still needs either VPN/private-network access or a different public connection path.
- The production `STRIPE_KEY` secret in `intelip_api/prd` now resolves successfully against the live IntelIP Stripe account.
- Google Secret Manager now contains the production backend secret set required by the current Cloud Run manifest, including the previously missing `chroma-*`, `composio-*`, `os-security-key`, and `jwt-verification-key` entries.
- The first intentional production backend Cloud Run deploy completed successfully, creating `intelip-agentos-prod` in `us-east1`.
- Production backend smoke checks now pass for `/health`, `/config`, inbox, and integrations status.

## Current product stack observed in repo

### Web app

- Cloudflare Workers for the app runtime and custom domains
- Neon Postgres for application data
- Clerk for auth
- Stripe for billing
- PostHog for analytics / telemetry
- Condere backend via Cloud Run-style authenticated service calls
- No R2 upload/storage baseline
- No OpenPanel or Grafana Faro frontend vendor path

### Backend / agent runtime

- Cloud Run deployment path for Condere
- OpenRouter for model access
- Composio for integrations
- Exa for optional web research
- Chroma Cloud and Redis are present in backend configuration and deployment docs
- OpenPanel secrets are still present in `intelip_api` dev Doppler, but there are no active OpenPanel codepaths left in the app repos; treat those secrets as stale until removed.

### Landing and growth surface

- Hostinger / VPS-hosted landing page
- PostHog on the landing page

## Vendor-grounded cost posture

### Product COGS / delivery

| Service | Current role in IntelIP | Pricing posture used for modeling | How modeled |
| --- | --- | --- | --- |
| Cloudflare Workers | App runtime for `app.intelip.co` and `staging.intelip.co` | Workers Paid starts at $5/mo and includes generous request / CPU quotas for an early-stage app | Small fixed monthly baseline |
| Neon | Primary app database | Launch plan is usage-based, with typical low-load spend around the mid-teens and published compute/storage rates | Fixed database baseline with room for staging + prod |
| Clerk | Auth and session management | Hobby is free; Pro starts at $20/mo annually | Fixed auth baseline |
| Stripe Payments | Paid checkout | No monthly base fee for standard payments, but card processing takes a percentage plus a fixed fee | Variable billing / payment processing line |
| Stripe Billing | Subscription logic | Pay-as-you-go pricing is 0.7% of billing volume if used | Folded into billing variable-cost reserve |
| Cloud Run / backend runtime | Condere agent API | Scale-to-zero reduces fixed cost at low traffic, but active agent work turns backend spend into a real variable line | Small fixed baseline plus workflow compute; production service now exists and should be treated as active infrastructure |
| OpenRouter / model spend | Agent reasoning | Pay-as-you-go by routed model | Variable token line |
| Composio | Business integrations and actions | Free tier is generous, then paid usage is inexpensive at low volumes | Variable integration-action line |
| Chroma Cloud | Retrieval / memory | Pure usage-based storage and read/write pricing | Fold into workflow and memory operations |
| Exa | Optional research / search enrichment | Usage-based and optional | Excluded from baseline, should be added per workflow if research becomes core |

### Operating overhead outside product COGS

| Tool | Role | Cost posture | Modeling choice |
| --- | --- | --- | --- |
| Linear | Planning and issue tracking | Team overhead, not user-driven delivery cost | Excluded from COGS, should sit in operating expense |
| Notion | Planning docs / PRDs | Team overhead | Excluded from COGS |
| GitHub | Source control / CI | Team overhead, though CI can affect delivery cost at scale | Excluded from current scenario pack |
| Doppler | Secret management for dev / ops | Team overhead | Excluded from current scenario pack |
| Hostinger / VPS | Landing page hosting | Real cash expense but not tightly coupled to app usage | Better treated as growth / operating overhead |

## Immediate cleanup implications

1. Remove the stale `OPENPANEL_CLIENT_ID` and `OPENPANEL_CLIENT_SECRET` entries from `intelip_api` once backend secret access is refreshed.
2. Keep R2 out of the IntelIP baseline entirely unless file uploads come back as a shipped feature.
3. Restore a valid production Stripe secret before attempting production billing validation or fee inspection.
4. Decide when to retire `STRIPE_PRICE_PRO_MONTHLY_ALIASES` after the remaining legacy live subscriptions are migrated off the old `Autumn Price (Fixed)` prices.
5. Decide whether local billing workflows should be powered primarily by Doppler or `.env.local`, now that `intelip_frontend/prd` has the production Stripe billing set.
6. Keep Chroma Cloud in the modeled stack for now because Condere still requires `CHROMA_API_KEY`, `CHROMA_TENANT`, and `CHROMA_DATABASE` in runtime config and GraphRAG adapters.
7. Keep Redis in the modeled stack for now because the Condere background worker and webhook queue still require `REDIS_URL`.

## Modeling decisions behind the revised scenario pack

1. Cloudflare and Neon were reduced from generic SaaS placeholders to closer-to-reality early-stage baselines.
2. Clerk was normalized to a realistic starter baseline instead of an inflated placeholder.
3. Manual onboarding and support reserve was increased sharply, because IntelIP's early go-to-market motion is likely high-touch.
4. Stripe-style payment processing is now explicit as a variable cost.
5. Integration cost was lowered to better match Composio's published low-volume economics, while keeping room for premium tool usage.
6. Cloud Run should now be modeled as a modest but real Condere staging + production footprint, not as staging-only infrastructure.

## Launch implication

For IntelIP, the real early risk is not Cloudflare or Neon. The early risk is:

- token spend if workflows become agent-heavy
- manual onboarding / support burden
- payment-fee drag once paid conversions begin
- optional premium integration or research workflows

That means pricing should protect against low-scale, high-touch pilots. A free tier plus seat-only pricing can look healthy at 30+ active users while still being painful for the first 10-20.
