"""Read-only Claude account checks for the default native macOS subscription."""

import http.client
import json
import math
import os
import ssl
import subprocess
import sys
import tempfile
from pathlib import Path


class EntitlementError(ValueError):
    """The operator's included subscription could not be verified for inference."""


def _native_account(binary, env):
    if sys.platform != "darwin":
        raise EntitlementError("Claude subscription preflight requires the native macOS Keychain profile.")
    import pwd

    owner = pwd.getpwuid(os.getuid())
    home = Path(owner.pw_dir)
    config = home / ".claude"
    if Path(env.get("HOME", "")) != home or Path(env.get("CLAUDE_CONFIG_DIR", config)) != config:
        raise EntitlementError("Use the default Claude profile for the current macOS user.")
    native_env = {"HOME": str(home), "PATH": env.get("PATH", os.defpath),
                  "USER": owner.pw_name, "LOGNAME": owner.pw_name, "LANG": "C"}
    with tempfile.TemporaryDirectory() as cwd:
        result = subprocess.run(
            [binary, "--restricted", "--setting-sources", "", "--settings",
             '{"disableAllHooks":true,"forceLoginMethod":"claudeai"}', "auth", "status", "--json"],
            cwd=cwd, env=native_env, capture_output=True, timeout=12,
        )
    if result.returncode:
        raise EntitlementError("Claude could not read its native subscription login.")
    account = json.loads(result.stdout)
    plan = account.get("subscriptionType")
    if (account.get("loggedIn") is not True or account.get("authMethod") != "claude.ai"
            or account.get("apiProvider") != "firstParty" or plan not in ("pro", "max")
            or account.get("configDirectory") != str(config)):
        raise EntitlementError("A signed-in first-party Claude Pro or Max subscription is required.")
    return owner.pw_name, plan


def _usage(account, plan):
    # Only the native account metadata endpoint receives this credential. A direct
    # HTTPS connection has no environment-proxy or redirect handling.
    item = subprocess.run(
        ["/usr/bin/security", "find-generic-password", "-a", account, "-s", "Claude Code-credentials", "-w"],
        capture_output=True, timeout=10, env={"PATH": os.defpath},
    )
    if item.returncode:
        raise EntitlementError("The current user's native Claude Keychain login is unavailable.")
    credential = json.loads(item.stdout)["claudeAiOauth"]
    token = credential.get("accessToken")
    if (credential.get("subscriptionType") != plan or "user:profile" not in credential.get("scopes", [])
            or not isinstance(token, str) or not token or not token.isascii()
            or any(ord(ch) < 33 or ord(ch) > 126 for ch in token)):
        raise EntitlementError("Claude's native credential could not verify this subscription account.")
    context = ssl.create_default_context(cafile="/etc/ssl/cert.pem")
    connection = http.client.HTTPSConnection("api.anthropic.com", timeout=12, context=context)
    try:
        connection.request("GET", "/api/oauth/usage", headers={
            "Authorization": "Bearer " + token, "anthropic-beta": "oauth-2025-04-20",
            "Accept": "application/json", "User-Agent": "profitctl-agent/0.1 (account metadata)",
        })
        response = connection.getresponse()
        if response.status != 200:
            raise EntitlementError(f"Claude account usage is unavailable (HTTP {response.status}).")
        data = response.read(128 * 1024 + 1)
        if len(data) > 128 * 1024:
            raise EntitlementError("Claude returned an unexpected account usage response.")
        return json.loads(data)
    finally:
        connection.close()


def _summary(plan, usage):
    enabled = (usage.get("extra_usage") or {}).get("is_enabled")
    if type(enabled) is not bool:
        raise EntitlementError("Claude did not confirm whether paid extra usage is enabled.")
    windows = {}
    for key in ("five_hour", "seven_day", "seven_day_sonnet"):
        details = usage.get(key)
        if details is None and key == "seven_day_sonnet":
            continue
        percent = (details or {}).get("utilization")
        if type(percent) not in (float, int) or not math.isfinite(percent) or percent < 0:
            raise EntitlementError(f"Claude did not confirm its included {key} utilization.")
        windows[key] = {"utilization": percent, "resets_at": details.get("resets_at")}
    return {"auth_method": "claude.ai", "api_provider": "firstParty", "plan": plan,
            "paid_extra_usage_enabled": enabled, "windows": windows}


def claude_subscription_status(binary, env):
    """Return only plan, extra-usage switch and relevant included quota windows."""
    try:
        account, plan = _native_account(binary, env)
        return _summary(plan, _usage(account, plan))
    except EntitlementError:
        raise
    except (OSError, subprocess.SubprocessError, ValueError, TypeError,
            AttributeError, KeyError, http.client.HTTPException):
        raise EntitlementError("Claude subscription status could not be verified securely.") from None


def assert_claude_subscription(binary, env):
    """Stop before inference if paid overages are possible or an included window is exhausted."""
    status = claude_subscription_status(binary, env)
    if status["paid_extra_usage_enabled"]:
        raise EntitlementError("Disable paid extra usage for this Claude subscription before running inference.")
    for window, details in status["windows"].items():
        if details["utilization"] >= 100:
            reset = f" Reset at {details['resets_at']}." if details["resets_at"] else ""
            raise EntitlementError(f"Claude's included {window} quota is exhausted.{reset}")
    return status
