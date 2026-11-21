// Run in the head so the requested palette is set before the first paint.
(() => {
  const choiceKey = "profitctl.theme";
  const previewKey = "profitctl.theme.preview";
  const valid = value => ["dark", "light", "system"].includes(value);
  const storedMode = (storageName, key) => {
    try {
      const value = window[storageName].getItem(key);
      return valid(value) ? value : null;
    } catch {
      return null;
    }
  };
  const storeMode = (storageName, key, value) => {
    try {
      if (value === null) window[storageName].removeItem(key);
      else window[storageName].setItem(key, value);
    } catch {
      // The current document still honors a choice when storage is unavailable.
    }
  };

  let preference;
  let preview;
  let picker;

  function render() {
    const mode = preview || preference || "system";
    document.documentElement.dataset.theme = mode;
    for (const meta of document.querySelectorAll("meta[data-theme-color]")) {
      const scheme = meta.dataset.themeColor;
      meta.media = mode === "system" ? "(prefers-color-scheme: " + scheme + ")" : mode === scheme ? "all" : "not all";
    }
    if (picker) picker.value = mode;
  }

  function restore() {
    preference = storedMode("localStorage", choiceKey);
    const shared = new URLSearchParams(location.search).get("theme");
    preview = valid(shared) ? shared : storedMode("sessionStorage", previewKey);
    if (valid(shared)) storeMode("sessionStorage", previewKey, shared);
    render();
  }

  restore();
  addEventListener("DOMContentLoaded", () => {
    const control = document.querySelector("[data-theme-control]");
    if (!control) return;
    picker = control.querySelector("select");
    picker.value = preview || preference || "system";
    control.hidden = false;
    picker.addEventListener("change", () => {
      if (!valid(picker.value)) return;
      preference = picker.value;
      preview = null;
      storeMode("sessionStorage", previewKey, null);
      storeMode("localStorage", choiceKey, preference === "system" ? null : preference);
      const url = new URL(location.href);
      if (url.searchParams.has("theme")) {
        url.searchParams.delete("theme");
        history.replaceState(history.state, "", url);
      }
      render();
    });
  }, { once: true });

  addEventListener("pageshow", event => { if (event.persisted) restore(); });
  addEventListener("storage", event => {
    if (event.key === choiceKey || event.key === null) {
      preference = storedMode("localStorage", choiceKey);
      render();
    }
  });
})();
