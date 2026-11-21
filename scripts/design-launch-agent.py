#!/usr/bin/env python3
"""Install and inspect the per-user macOS launchd service for ProfitCtl design reviews."""

import argparse
import json
import os
from pathlib import Path
import plistlib
import pwd
import stat
import subprocess
import sys
import tempfile
import time

from profitctl_agent.design_service import request


LABEL = "co.intelip.profitctl.design-agent"
DOMAIN = f"gui/{os.getuid()}"
TARGET = f"{DOMAIN}/{LABEL}"
LAUNCHCTL = "/bin/launchctl"


def private_directory(path, *, create):
    if create:
        path.mkdir(mode=0o700, parents=True, exist_ok=True)
    info = path.lstat()
    if not stat.S_ISDIR(info.st_mode) or info.st_uid != os.getuid() or info.st_mode & 0o077:
        raise ValueError(f"Expected an owner-only directory at {path}")


def read_unit(path):
    if not path.exists() and not path.is_symlink():
        return None
    info = path.lstat()
    if not stat.S_ISREG(info.st_mode) or info.st_uid != os.getuid() or info.st_nlink != 1:
        raise ValueError("Unexpected file at the launch agent path")
    data = path.read_bytes()
    if plistlib.loads(data).get("Label") != LABEL:
        raise ValueError("A different launch agent owns this file")
    return data


def write_unit(path, data):
    with tempfile.NamedTemporaryFile(dir=path.parent, prefix=".profitctl-", delete=False) as stream:
        temp = Path(stream.name)
        try:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
            os.replace(temp, path)
        finally:
            temp.unlink(missing_ok=True)


def loaded():
    return subprocess.run([LAUNCHCTL, "print", TARGET], stdout=subprocess.DEVNULL,
                          stderr=subprocess.DEVNULL).returncode == 0


def bootout():
    if loaded():
        subprocess.run([LAUNCHCTL, "bootout", TARGET], check=True)
        # bootout can return before launchd releases the service name.
        for _ in range(100):
            if not loaded():
                return
            time.sleep(.05)
        raise RuntimeError("launchd has not unloaded the previous service")


def start(path):
    subprocess.run([LAUNCHCTL, "enable", TARGET], check=True)
    subprocess.run([LAUNCHCTL, "bootstrap", DOMAIN, str(path)], check=True)


def health(socket_path, *, wait=False):
    deadline = time.monotonic() + (20 if wait else 0)
    while True:
        try:
            result = request(socket_path, {"action": "health"}, timeout=1)
            if result.get("status") == "ok" and type(result.get("pid")) is int:
                return result
            raise ValueError("Unexpected health response")
        except (OSError, ValueError, RuntimeError):
            if time.monotonic() >= deadline:
                raise RuntimeError("The ProfitCtl design agent did not report healthy") from None
            time.sleep(.15)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("install", "status", "uninstall"))
    parser.add_argument("--runtime", type=Path, default=Path.home() / "Library/Application Support/ProfitCtl/design")
    parser.add_argument("--workspace", type=Path, help="Existing private investigation workspace")
    parser.add_argument("--profitctl-binary", type=Path, help="Go executable from this same release")
    parser.add_argument("--model", choices=("claude-subscription", "scripted"), default="claude-subscription",
                        help="Existing investigator setting. Design reviews perform no inference.")
    args = parser.parse_args()
    if sys.platform != "darwin":
        parser.error("This service manager requires macOS")
    if not args.runtime.is_absolute():
        parser.error("--runtime must be absolute")
    runtime = args.runtime
    unit = Path.home() / "Library/LaunchAgents" / (LABEL + ".plist")
    sock = runtime / "agent.sock"
    before = read_unit(unit)
    if args.action == "status":
        print(json.dumps({"label": LABEL, "loaded": loaded(), "health": health(sock)}))
        return
    if args.action == "uninstall":
        if before is None and loaded():
            raise ValueError("No installed ProfitCtl plist owns the running service")
        bootout()
        if before is not None:
            unit.unlink()
        print(json.dumps({"label": LABEL, "loaded": loaded()}))
        return
    if args.workspace is None or args.profitctl_binary is None:
        parser.error("install requires --workspace and --profitctl-binary")
    if len(os.fsencode(sock)) >= 104:
        parser.error("The macOS socket path must fit in 103 bytes")
    workspace, binary = args.workspace.resolve(strict=True), args.profitctl_binary.resolve(strict=True)
    private_directory(workspace, create=False)
    private_directory(runtime, create=True)
    profiles = runtime / "profiles"
    private_directory(profiles, create=True)
    # The interpreter in the frozen venv identifies its matching agent entry point.
    entry = Path(sys.executable).parent / "profitctl-agent"
    if not (entry.is_file() and os.access(entry, os.X_OK) and binary.is_file() and os.access(binary, os.X_OK)):
        parser.error("Use the installed agent's venv Python and its matching ProfitCtl binary")
    owner = pwd.getpwuid(os.getuid())
    path = os.pathsep.join((str(Path.home()/".local/bin"), str(Path.home()/".npm-global/bin"),
                            "/opt/homebrew/bin", "/usr/local/bin", os.defpath, "/usr/sbin", "/sbin"))
    program = ["/usr/bin/env", "-i", f"HOME={owner.pw_dir}", f"USER={owner.pw_name}", f"LOGNAME={owner.pw_name}",
               f"PATH={path}", "LANG=en_US.UTF-8", "DISABLE_TELEMETRY=1", f"PROFITCTL_AGENT_MODEL={args.model}",
               str(entry), "--workspace", str(workspace), "--profitctl-binary", str(binary),
               "serve", "--socket", str(sock), "--profiles", str(profiles)]
    after = plistlib.dumps({"Label": LABEL, "ProgramArguments": program, "RunAtLoad": True, "KeepAlive": True,
                            "ProcessType": "Background", "ThrottleInterval": 5, "Umask": 0o077,
                            "WorkingDirectory": str(runtime)})
    if before == after and loaded():
        print(json.dumps(health(sock, wait=True)))
        return
    if before is None and loaded():
        raise ValueError("A running launch agent already owns this label without our plist")
    unit.parent.mkdir(parents=True, mode=0o700, exist_ok=True)
    was_running = loaded()
    if was_running:
        bootout()
    try:
        write_unit(unit, after)
        start(unit)
        result = health(sock, wait=True)
    except BaseException:
        bootout()
        if before is None:
            unit.unlink(missing_ok=True)
        else:
            write_unit(unit, before)
            if was_running:
                start(unit)
        raise
    print(json.dumps(result))


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, RuntimeError, subprocess.CalledProcessError) as exc:
        sys.exit(str(exc))
