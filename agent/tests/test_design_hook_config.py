import json
from pathlib import Path
import shlex
import sys
import tempfile
import unittest
from unittest.mock import patch

from profitctl_agent import design_hook_config as config


class NativeConfiguration(unittest.TestCase):
    def setUp(self):
        root = Path(self.enterContext(tempfile.TemporaryDirectory())).resolve()
        self.codex = root / "codex" / "hooks.json"
        self.claude = root / "claude" / "settings.json"
        self.codex.parent.mkdir()
        self.claude.parent.mkdir()
        self.routes = root / "route's archive.json"
        self.routes.write_text('{"version":1,"routes":[]}')
        self.socket = root / "service with spaces.sock"
        self.backups = root / "backups"
        self.args = dict(codex_config=self.codex, claude_config=self.claude, python=Path(sys.executable),
                         socket=self.socket, routes=self.routes, backup_dir=self.backups)

    def test_targeted_install_preserves_native_config_and_uninstall_restores_original_bytes(self):
        old_codex = b'{ "description"   :   "keep bytes", "hooks": {"Stop":[{"hooks":[{"type":"command","command":"other"}]}]}}\n'
        old_claude = b'{ "model": "from operator", "permissions"  : { "allow": [] } }\n'
        self.codex.write_bytes(old_codex)
        self.claude.write_bytes(old_claude)
        backup = config.configure("install", **self.args)
        self.assertEqual((backup/"codex.before").read_bytes(), old_codex)
        self.assertEqual((backup/"claude.before").read_bytes(), old_claude)
        self.assertEqual((backup/"claude.before").stat().st_mode & 0o777, 0o600)
        installed = self.codex.read_bytes()
        self.assertIn(b'"description"   :   "keep bytes"', installed)
        self.assertIn(b'"permissions"  : { "allow": [] }', self.claude.read_bytes())
        hooks = json.loads(installed)["hooks"]
        self.assertEqual(hooks["Stop"][0]["hooks"][0]["command"], "other")
        for event in config.EVENTS:
            entry = hooks[event][-1]["hooks"][0]
            self.assertEqual(shlex.split(entry["command"]),
                             [sys.executable, "-I", "-m", "profitctl_agent.design_hook", "--socket", str(self.socket),
                              "--routes", str(self.routes)])
            self.assertEqual(entry["timeout"], 4)
        self.assertIsNone(config.configure("install", **self.args))
        with self.assertRaisesRegex(ValueError, "different ProfitCtl hook"):
            config.configure("install", **(self.args | {"socket": self.socket.with_name("another.sock")}))
        self.assertEqual(self.codex.read_bytes(), installed)
        config.configure("uninstall", **self.args)
        self.assertEqual(self.codex.read_bytes(), old_codex)
        self.assertEqual(self.claude.read_bytes(), old_claude)
        self.assertIsNone(config.configure("uninstall", **self.args))

    def test_changed_unrelated_settings_and_cohabiting_hooks_survive(self):
        config.configure("install", **self.args)
        settings = json.loads(self.claude.read_text())
        settings["new_user_setting"] = [1, 2, 3]
        settings["hooks"]["Stop"][0]["hooks"].append({"type": "command", "command": "another"})
        self.claude.write_text(json.dumps(settings))
        config.configure("uninstall", **self.args)
        self.assertFalse(self.codex.exists())
        self.assertEqual(json.loads(self.claude.read_text()),
                         {"new_user_setting": [1,2,3], "hooks": {"Stop": [{"hooks": [{"type":"command","command":"another"}]}]}})

    def test_malformed_duplicate_and_symlinked_configs_fail_before_any_change(self):
        for data in ('{} broken', '{"model":"a","model":"b"}'):
            self.claude.write_text(data)
            with self.assertRaises(ValueError):
                config.configure("install", **self.args)
            self.assertFalse(self.codex.exists())
            self.assertEqual(self.claude.read_text(), data)
        self.claude.unlink()
        self.claude.symlink_to(self.routes)
        with self.assertRaises(OSError):
            config.configure("install", **self.args)
        self.assertFalse(self.codex.exists())
        self.assertTrue(self.claude.is_symlink())

    def test_conflicting_manually_edited_hook_is_preserved(self):
        config.configure("install", **self.args)
        data = json.loads(self.claude.read_text())
        data["hooks"]["Stop"][0]["hooks"][0]["timeout"] = 15
        self.claude.write_text(json.dumps(data))
        old = self.codex.read_bytes()
        with self.assertRaisesRegex(ValueError, "edited"):
            config.configure("uninstall", **self.args)
        self.assertEqual(json.loads(self.claude.read_text()), data)
        self.assertEqual(self.codex.read_bytes(), old)

    def test_second_file_concurrent_change_rolls_back_first(self):
        self.codex.write_text('{"existing": 1}')
        self.claude.write_text('{"model": "before"}')
        real = config._write
        def write(path, old, new):
            if path == self.claude:
                self.claude.write_text('{"model": "changed elsewhere"}')
            return real(path, old, new)
        with patch.object(config, "_write", side_effect=write), self.assertRaisesRegex(ValueError,"changed during"):
            config.configure("install", **self.args)
        self.assertEqual(self.codex.read_text(), '{"existing": 1}')
        self.assertEqual(self.claude.read_text(), '{"model": "changed elsewhere"}')
