import json
import os
import subprocess
import unittest
from unittest.mock import patch

from langchain_core.messages import HumanMessage

from profitctl_agent.subscription import ClaudeSubscriptionModel, _message


class SubscriptionTest(unittest.TestCase):
    def test_two_reads_return_to_the_graph_without_granting_the_cli_tools(self):
        names = ["inspect_case_evidence", "read_outcome_comparison"]
        calls = [{"tool": name, "arguments": "{}"} for name in names]
        cli_reply = [
            {"type": "system", "subtype": "init", "apiKeySource": "none", "mcp_servers": [], "model": "sonnet"},
            {"type": "result", "structured_output": {"kind": "tools", "calls": calls, "text": ""}},
        ]
        val = subprocess.CompletedProcess([], 0, json.dumps(cli_reply), "")
        model = ClaudeSubscriptionModel()
        with patch.dict(os.environ, {"ANTHROPIC_API_KEY": "example"}), \
             patch("profitctl_agent.subscription.shutil.which", return_value="/test/claude"), \
             patch("profitctl_agent.subscription.assert_claude_subscription") as check, \
             patch("profitctl_agent.subscription.subprocess.run", return_value=val) as run:
            msg = model.invoke([HumanMessage("Inspect this case.")], tools=[{"name": n} for n in names])
        self.assertEqual([c["name"] for c in msg.tool_calls], names)
        argv = run.call_args.args[0]
        env = run.call_args.kwargs["env"]
        self.assertNotIn("ANTHROPIC_API_KEY", env)
        self.assertNotIn("ANTHROPIC_BASE_URL", env)
        self.assertEqual(argv[argv.index("--tools") + 1], "")
        self.assertEqual(json.loads(argv[argv.index("--mcp-config") + 1]), {"mcpServers": {}})
        self.assertEqual(check.call_args.args, ("/test/claude", env))

    def test_unavailable_capacity_stops_before_model_and_unknown_tools_are_rejected(self):
        with patch("profitctl_agent.subscription.shutil.which", return_value="/test/claude"), \
             patch("profitctl_agent.subscription.assert_claude_subscription", side_effect=RuntimeError("quota exhausted")), \
             patch("profitctl_agent.subscription.subprocess.run") as run:
            with self.assertRaisesRegex(RuntimeError, "quota exhausted"):
                ClaudeSubscriptionModel().invoke([HumanMessage("Inspect this case.")])
            run.assert_not_called()
        with self.assertRaisesRegex(ValueError, "unavailable"):
            _message({"kind": "tools", "calls": [{"tool": "other", "arguments": "{}"}]}, ["inspect_case_evidence"])
