"""Private socket, deterministic calculator and cross-process MCP contracts."""

import asyncio
import json
import os
from pathlib import Path
import socket
import stat
import subprocess
import sys
import time
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

from case_fixture import EconomicCaseFixture, REPO
from profitctl_agent.design_service import (
    MAX_REQUEST, MAX_RESPONSE, SocketClient, _listener, _read, _review, request,
)


class PersistentService(EconomicCaseFixture):
    def setUp(self):
        super().setUp()
        self.path = self.root / "agent.sock"
        self.profiles = self.root / "profiles"
        self.profiles.mkdir(mode=0o700)
        self.env = {**os.environ, "PYTHONPATH": str(REPO / "agent/src"), "PROFITCTL_AGENT_MODEL": "scripted"}

    def launch(self, binary):
        command = [sys.executable, "-m", "profitctl_agent.cli", "--workspace", str(self.root),
                   "--profitctl-binary", str(binary), "serve", "--socket", str(self.path),
                   "--profiles", str(self.profiles)]
        proc = subprocess.Popen(command, env=self.env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        self.addCleanup(self.stop, proc)
        deadline = time.monotonic() + 20
        while time.monotonic() < deadline and proc.poll() is None:
            try:
                if request(self.path, {"action": "health"}, timeout=0.5)["pid"] == proc.pid:
                    return proc
            except OSError:
                pass
            time.sleep(0.05)
        self.fail("Service did not become ready.")

    @staticmethod
    def stop(proc):
        if proc.poll() is None:
            proc.terminate()
            try:
                proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                proc.kill()
                proc.wait(timeout=10)

    async def invoke_mcp(self, name, arguments):
        params = StdioServerParameters(command=sys.executable, args=[
            "-m", "profitctl_agent.cli", "--workspace", str(self.root), "mcp", "--socket", str(self.path),
        ], env={"PYTHONPATH": self.env["PYTHONPATH"]})
        async with stdio_client(params) as (reader, writer), ClientSession(reader, writer) as session:
            await session.initialize()
            advertised = {t.name: t for t in (await session.list_tools()).tools}
            self.assertEqual(set(advertised), {"profitctl_start_investigation", "profitctl_investigation_status",
                                              "profitctl_resume_investigation", "profitctl_review_design"})
            self.assertTrue(advertised["profitctl_review_design"].annotations.readOnlyHint)
            result = await session.call_tool(name, arguments)
            self.assertFalse(result.isError)
            return result.structuredContent

    def test_two_stdio_sessions_share_durable_run_across_restart(self):
        first = self.launch(self.binary)
        start = asyncio.run(self.invoke_mcp("profitctl_start_investigation", {
            "run_id": "handoff", "case_id": "research", "question": "Which depth is economical?",
        }))
        self.assertEqual(start["status"], "awaiting_evidence")
        self.stop(first)
        self.assertFalse(self.path.exists())
        second = self.launch(self.binary)
        self.assertNotEqual(first.pid, second.pid)
        restored = SocketClient(self.path).status(run_id="handoff")
        self.assertEqual(restored["checkpoint_id"], start["checkpoint_id"])
        finish = asyncio.run(self.invoke_mcp("profitctl_resume_investigation", {
            "run_id": "handoff", "expected_revision": 1, "action": "decline",
        }))
        self.assertEqual(finish["status"], "declined")
        self.assertEqual(finish["base_revision"], 1)

    def test_registered_and_inline_review_slow_connections_and_input_bounds(self):
        self.env["PROFITCTL_AGENT_MODEL"] = ""
        self.launch(self.binary)
        self.assertEqual(stat.S_IMODE(self.path.stat().st_mode), 0o600)
        proposal = json.loads((REPO / "examples/design/condere-queue.json").read_bytes())
        profile = self.profiles / "condere.json"
        profile.write_text(json.dumps(proposal))
        profile.chmod(0o600)
        client = SocketClient(self.path)
        with socket.socket(socket.AF_UNIX) as slow:
            slow.connect(str(self.path))
            slow.sendall((128).to_bytes(4) + b"{")
            named = client.review(profile="condere")
            self.assertEqual(client.review(proposal=proposal), named)
            self.assertEqual(named["status"], "complete")
            self.assertTrue(named["summary"])
        gap = {"question": "What is the usual monthly volume?", "source": "Admission counts", "owner": "Operator"}
        incomplete = client.review(proposal={**proposal, "expected_monthly": {"value": None, "missing": gap}})
        self.assertEqual(incomplete["status"], "incomplete")
        self.assertTrue(incomplete["missing_inputs"])
        self.assertEqual(asyncio.run(self.invoke_mcp("profitctl_review_design", {"profile": "condere"})), named)
        with socket.socket(socket.AF_UNIX) as oversized:
            oversized.connect(str(self.path))
            oversized.sendall((MAX_REQUEST + 1).to_bytes(4))
            error = _read(oversized, MAX_RESPONSE, time.monotonic() + 2)
            self.assertIn("size limit", error["error"])
        with self.assertRaisesRegex(ValueError, "exactly one"):
            request(self.path, {"action": "review", "profile": "condere", "proposal": proposal})
        with self.assertRaisesRegex(ValueError, "slug"):
            client.review(profile="../condere")
        with self.assertRaisesRegex(ValueError, "registered design profile"):
            client.review(profile="missing")
        with self.assertRaisesRegex(ValueError, "valid self-contained") as invalid:
            client.review(proposal={"raw_prompt": "private error content"})
        self.assertNotIn("private error content", str(invalid.exception))
        with self.assertRaisesRegex(ValueError, "size limit"):
            client.review(proposal={"x": "a" * MAX_REQUEST})
        (self.profiles / "link.json").symlink_to(profile)
        with self.assertRaisesRegex(ValueError, "unavailable"):
            client.review(profile="link")
        os.mkfifo(self.profiles / "pipe.json", mode=0o600)
        with self.assertRaises(ValueError):
            client.review(profile="pipe")
        profile.chmod(0o644)
        with self.assertRaisesRegex(ValueError, "private"):
            client.review(profile="condere")
        self.assertEqual(request(self.path, {"action": "health"})["status"], "ok")


class ListenerSafety(unittest.TestCase):
    def test_preserves_active_socket_and_files_and_reclaims_stale_socket(self):
        with TemporaryDirectory() as tmp:
            root = Path(tmp)
            path = root / "p.sock"
            path.write_text("operator file")
            with self.assertRaisesRegex(ValueError, "occupied"), _listener(path):
                pass
            self.assertEqual(path.read_text(), "operator file")
            path.unlink()
            with socket.socket(socket.AF_UNIX) as other:
                other.bind(str(path))
                other.listen()
                with self.assertRaisesRegex(ValueError, "another listener"), _listener(path):
                    pass
                self.assertTrue(path.is_socket())
            with _listener(path):
                with self.assertRaisesRegex(ValueError, "already owns"), _listener(path):
                    pass
            self.assertFalse(path.exists())
            root.chmod(0o755)
            with self.assertRaisesRegex(ValueError, "private"), _listener(path):
                pass

    def test_malformed_or_inconsistent_go_reply_is_not_an_incomplete_report(self):
        for output in (b"", b"{}", b'{"schema_version":"profitctl.design-report/v1","status":"complete","summary":"x"}'):
            operation = subprocess.CompletedProcess([], 1, stdout=output)
            with patch("profitctl_agent.design_service.subprocess.run", return_value=operation):
                with self.assertRaises(ValueError):
                    _review("profitctl", "/", {"schema_version": "profitctl.design/v1"})
