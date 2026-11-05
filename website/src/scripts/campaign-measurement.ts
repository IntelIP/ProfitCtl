type CampaignEvent = "campaign_landing" | "demo_started" | "comparison_completed" | "example_comparison_completed" | "quickstart_opened" | "package_download_clicked";
type Attribution = { key: string; medium: string; creative: string; revision: string; test: string; expires: number; sent: string[] };
declare global { interface Window { op?: ((...args: unknown[]) => void) & { q?: unknown[] }; } }
const storageKey = "profitctl.campaign.v1";
const safe = /^[a-z0-9][a-z0-9_-]{0,79}$/;

function attribution(): Attribution | null {
  if (!window.op || navigator.doNotTrack === "1" || (navigator as Navigator & { globalPrivacyControl?: boolean }).globalPrivacyControl) return null;
  try {
    const params = new URL(location.href).searchParams;
    const previous = JSON.parse(sessionStorage.getItem(storageKey) || "null") as Attribution | null;
    const key = params.get("utm_campaign"), medium = params.get("utm_medium"), creative = params.get("utm_content"), revision = params.get("measurement_revision");
    const tagged = params.get("utm_source") === "linkedin" && key && safe.test(key) && creative && safe.test(creative) &&
      ["organic_social", "paid_social"].includes(medium || "") && /^[1-9][0-9]{0,2}$/.test(revision || "");
    let value = previous && previous.expires > Date.now() ? previous : null;
    if (tagged && (!value || value.key !== key || value.medium !== medium || value.creative !== creative || value.revision !== revision)) {
      value = { key, medium: medium!, creative: creative!, revision: revision!, test: params.get("campaign_test") === "1" ? "true" : "false", expires: Date.now() + 30*60*1000, sent: [] };
    }
    if (value) { value.expires = Date.now() + 30*60*1000; sessionStorage.setItem(storageKey, JSON.stringify(value)); }
    return value;
  } catch { return null; }
}

export function trackCampaignAction(event: CampaignEvent) {
  const a = attribution();
  if (!a || a.sent.includes(event)) return;
  if (event !== "campaign_landing" && !a.sent.includes("campaign_landing")) trackCampaignAction("campaign_landing");
  window.op!("track", event, { campaign_key: a.key, campaign_source: "linkedin", campaign_medium: a.medium,
    campaign_creative: a.creative, measurement_revision: a.revision, campaign_test: a.test });
  try {
    const latest = JSON.parse(sessionStorage.getItem(storageKey)!) as Attribution;
    latest.sent = [...new Set([...latest.sent, event])];
    sessionStorage.setItem(storageKey, JSON.stringify(latest));
  } catch { /* No storage means no tracked attribution on the next action. */ }
}

export function initializeCampaignMeasurement() {
  trackCampaignAction("campaign_landing");
  document.addEventListener("click", (event) => {
    if (!event.isTrusted) return;
    const link = (event.target as Element | null)?.closest<HTMLAnchorElement>("a[href]");
    if (!link) return;
    const url = new URL(link.href);
    if (url.origin === location.origin && (url.pathname.startsWith("/docs") || url.hash === "#get-started")) trackCampaignAction("quickstart_opened");
    if (url.origin === "https://github.com" && url.pathname.startsWith("/IntelIP/ProfitCtl/releases/download/")) trackCampaignAction("package_download_clicked");
  });
}
