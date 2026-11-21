import unittest
from unittest.mock import patch

from profitctl_agent.entitlement import EntitlementError, _summary, assert_claude_subscription


def usage(enabled=False, five_hour=70):
    return {"extra_usage": {"is_enabled": enabled},
            "five_hour": {"utilization": five_hour, "resets_at": "2026-11-21T03:30:00Z"},
            "seven_day": {"utilization": 24}, "seven_day_sonnet": None}


class SubscriptionPreflight(unittest.TestCase):
    def test_unknown_billing_or_quota_fails_closed(self):
        for enabled in (None, 0, "false"):
            with self.assertRaisesRegex(EntitlementError, "paid extra usage"):
                _summary("max", usage(enabled=enabled))
        for value in (None, False, -1, float("nan")):
            with self.assertRaisesRegex(EntitlementError, "five_hour"):
                _summary("max", usage(five_hour=value))

    def test_extra_usage_off_and_all_included_windows_available(self):
        with patch("profitctl_agent.entitlement._native_account", return_value=("user", "max")), \
                patch("profitctl_agent.entitlement._usage") as endpoint:
            endpoint.return_value = usage(enabled=True)
            with self.assertRaisesRegex(EntitlementError, "Disable paid extra usage"):
                assert_claude_subscription("claude", {})
            endpoint.return_value = usage(five_hour=100)
            with self.assertRaisesRegex(EntitlementError, "03:30:00"):
                assert_claude_subscription("claude", {})
            endpoint.return_value = usage() | {"seven_day_sonnet": {"utilization": 100}}
            with self.assertRaisesRegex(EntitlementError, "seven_day_sonnet"):
                assert_claude_subscription("claude", {})
            endpoint.return_value = usage()
            status = assert_claude_subscription("claude", {})
            self.assertIs(status["paid_extra_usage_enabled"], False)
            self.assertEqual(status["windows"]["seven_day"]["utilization"], 24)
