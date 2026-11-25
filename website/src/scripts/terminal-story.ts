export const terminalStory = [
  { title: "Check the scenario", detail: "A fictional AI product starts with 1,000 users, a $20 plan, and a 60% margin target.", args: ["validate", "-f", "demo/baseline.yml"], exitCode: 0, scenario: "baseline" },
  { title: "Run the baseline", detail: "At 60 research requests per user, the product clears its margin target.", args: ["simulate", "-f", "demo/baseline.yml"], exitCode: 0, scenario: "baseline" },
  { title: "Let usage grow", detail: "Usage rises to 300 requests. Revenue stays the same, while the margin target fails.", args: ["simulate", "-f", "demo/growth.yml"], exitCode: 1, scenario: "growth" },
  { title: "Lower the cost per request", detail: "At the same usage, cutting the cost per request from $0.08 to $0.02 brings the margin back above 60%.", args: ["compare", "demo/baseline.yml", "demo/growth.yml", "demo/optimized.yml"], exitCode: 1, scenario: "optimized" },
];
