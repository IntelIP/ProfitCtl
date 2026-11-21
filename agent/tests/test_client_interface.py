"""Public CLI and MCP contracts for cross-client investigation handoffs."""

import asyncio
import json
from tempfile import TemporaryDirectory
from pathlib import Path
import unittest
from unittest.mock import Mock

from mcp.server.fastmcp.exceptions import ToolError

from profitctl_agent.cli import _evidence, _parser, make_server


class ClientContract(unittest.TestCase):
    def test_mcp_handoff_preserves_ids_and_validates_evidence(self):
        async def exercise():
            service = Mock()
            server = make_server(service)
            tools = {tool.name: tool for tool in await server.list_tools()}
            self.assertEqual(set(tools), {
                "profitctl_start_investigation", "profitctl_investigation_status",
                "profitctl_resume_investigation",
            })
            status = tools["profitctl_investigation_status"]
            self.assertTrue(status.annotations.readOnlyHint)
            self.assertEqual(set(status.inputSchema["properties"]), {"run_id"})
            service.status.return_value = {"run_id": "depth-01", "base_revision": 4}
            _, state = await server.call_tool(status.name, {"run_id": "depth-01"})
            self.assertEqual(state, service.status.return_value)

            resume = tools["profitctl_resume_investigation"]
            self.assertFalse(resume.annotations.readOnlyHint)
            self.assertIn("expected_revision", resume.inputSchema["required"])
            args = {"run_id": "depth-01", "expected_revision": 4, "action": "supply"}
            with self.assertRaises(ToolError):
                await server.call_tool(resume.name, {
                    **args, "evidence": {"kind": "observed", "updates": [{"option": "A", "accepted": 2}]}
                })
            service.resume.assert_not_called()
            observation = {"kind": "assumed", "source": "scenario", "updates": [{"option": "A", "accepted": 2}]}
            service.resume.return_value = {"run_id": "depth-01", "decision_revision": 5}
            _, final_state = await server.call_tool(resume.name, {**args, "evidence": observation})
            self.assertEqual(final_state, service.resume.return_value)
            service.resume.assert_called_once_with(**args, evidence=observation)

        asyncio.run(exercise())

    def test_cli_reads_bounded_evidence_file_and_decline(self):
        evidence = {"kind": "assumed", "source": "trial", "updates": [{"option": "A", "accepted": 1}]}
        with TemporaryDirectory() as tmp:
            path = Path(tmp) / "evidence.json"
            path.write_text(json.dumps(evidence))
            args = _parser().parse_args([
                "--workspace", tmp, "resume", "depth-01", "--expected-revision", "3",
                "--action", "supply", "--evidence-file", str(path),
            ])
            self.assertEqual(_evidence(args), evidence)
            path.write_bytes(b" " * 65537)
            with self.assertRaisesRegex(ValueError, "64 KiB"):
                _evidence(args)
            args.action = "decline"
            with self.assertRaisesRegex(ValueError, "no evidence"):
                _evidence(args)
            args.evidence_file = None
            self.assertIsNone(_evidence(args))
