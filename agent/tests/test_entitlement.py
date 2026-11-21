import json
import unittest
from unittest.mock import Mock, patch

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


class RateLimitedSubscriptionCheck(unittest.TestCase):
    def setUp(self):
        self.enterContext(patch("profitctl_agent.entitlement._native_account", return_value=("user", "max")))
        credential = {"claudeAiOauth": {"accessToken": "example", "subscriptionType": "max", "scopes": ["user:profile"]}}
        self.enterContext(patch("profitctl_agent.entitlement.subprocess.run",
                                return_value=Mock(returncode=0, stdout=json.dumps(credential))))
        self.connections = self.enterContext(patch("profitctl_agent.entitlement.http.client.HTTPSConnection"))
        self.sleep = self.enterContext(patch("profitctl_agent.entitlement.time.sleep"))

    @staticmethod
    def connection(status, *, retry=None, body=None):
        connection = Mock()
        response = connection.getresponse.return_value
        response.status = status
        response.getheader.return_value = retry
        response.read.return_value = json.dumps(body).encode()
        return connection

    def test_short_retry_recovers_and_next_model_call_checks_again(self):
        limited = self.connection(429, retry="4")
        self.connections.side_effect = [limited, self.connection(200, body=usage()),
                                        self.connection(200, body=usage(enabled=True))]
        self.assertIs(assert_claude_subscription("claude", {})["paid_extra_usage_enabled"], False)
        self.sleep.assert_called_once_with(4)
        limited.close.assert_called_once()
        limited.getresponse.return_value.getheader.assert_called_once_with("Retry-After")
        with self.assertRaisesRegex(EntitlementError, "Disable paid extra usage"):
            assert_claude_subscription("claude", {})
        self.assertEqual(self.connections.call_count, 3)

    def test_http_date_and_second_rate_limit_stop_after_one_retry(self):
        self.connections.side_effect = [self.connection(429, retry="Thu, 01 Jan 1970 00:00:07 GMT"),
                                        self.connection(429, retry="12")]
        with patch("profitctl_agent.entitlement.time.time", return_value=4.1), \
                self.assertRaisesRegex(EntitlementError, "HTTP 429.*12s.*retry the same run"):
            assert_claude_subscription("claude", {})
        self.sleep.assert_called_once_with(3)
        self.assertEqual(self.connections.call_count, 2)

    def test_long_missing_and_invalid_retry_times_return_to_operator_immediately(self):
        for header, expected in [("4000000", "4000000s"), (None, "one minute"), ("broken", "one minute")]:
            with self.subTest(retry_after=header):
                self.connections.reset_mock()
                self.connections.side_effect = [self.connection(429, retry=header)]
                with self.assertRaisesRegex(EntitlementError, f"HTTP 429.*{expected}.*retry the same run"):
                    assert_claude_subscription("claude", {})
                self.connections.assert_called_once()
                self.sleep.assert_not_called()
