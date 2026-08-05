# Agent Recommendation Fixture

Recommendation: Use Cloudflare Workers for the public AI SaaS path because the ProfitCtl compare run shows lower monthly cost while all covenants pass.

Assumptions: 100 active users, 8 percent monthly growth, $49 ARPU, and template-level AI/search usage. Source provenance is `template` for platform defaults and `user_supplied` for ARPU. Confidence is medium for business inputs and low to medium for template costs.

Economics: Monthly fixed cost is $85. Variable cost drivers are LLM calls, search calls, database usage, and auth. Modeled gross margin is 78 percent, p95 margin is 62 percent, and cost/user is $9.40. Covenant status passes margin, p95 margin, and cost per user thresholds.

Tradeoff: Workers is cheaper, but Cloud Run remains viable when private IAM, container runtime behavior, or longer process limits are hard requirements.

Alternative: Cloud Run is the cheaper-to-operate engineering choice only if the team already needs container semantics; otherwise it is the more expensive viable alternative for this scenario.
