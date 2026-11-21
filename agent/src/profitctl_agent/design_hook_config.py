"""Install or remove the two native ProfitCtl command hooks in user-owned JSON configuration."""
import argparse
import fcntl
import hashlib
import json
import os
from pathlib import Path
import shlex
import stat
import tempfile


EVENTS = ("UserPromptSubmit", "Stop")
MAX_CONFIG = 4 * 1024 * 1024
MISSING = object()


def _object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("Duplicate JSON keys in the configuration.")
        result[key] = value
    return result


def _directory(path, private=False):
    if not path.is_absolute() or path.is_symlink() or ".." in path.parts:
        raise ValueError("Use an absolute, owner-controlled directory.")
    path.mkdir(mode=0o700, parents=True, exist_ok=True)
    st = path.stat()
    if st.st_uid != os.getuid() or st.st_mode & (0o077 if private else 0o022):
        raise ValueError("The target directory has unsafe ownership or permissions.")


def _read(path):
    try:
        fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
    except FileNotFoundError:
        return None
    with os.fdopen(fd, "rb") as stream:
        st = os.fstat(fd)
        if (not stat.S_ISREG(st.st_mode) or st.st_uid != os.getuid()
                or st.st_mode & 0o022 or st.st_size > MAX_CONFIG):
            raise ValueError("Expected a bounded owner-controlled regular file.")
        data = stream.read(MAX_CONFIG + 1)
        if len(data) > MAX_CONFIG:
            raise ValueError("Configuration is too large.")
        return data


def _write(path, old, new):
    if _read(path) != old:
        raise ValueError("Configuration changed during the edit. Re-read it and retry.")
    if new is None:
        path.unlink()
        return
    with tempfile.NamedTemporaryFile(dir=path.parent, delete=False) as out:
        candidate = Path(out.name)
        try:
            os.fchmod(out.fileno(), path.stat().st_mode & 0o777 if old is not None else 0o600)
            out.write(new)
            out.flush()
            os.fsync(out.fileno())
        except BaseException:
            candidate.unlink(missing_ok=True)
            raise
    try:
        if _read(path) != old:
            raise ValueError("Configuration changed during the edit. Re-read it and retry.")
        os.replace(candidate, path)
    finally:
        candidate.unlink(missing_ok=True)


def _set_hooks(source, hooks):
    """Splice only the top-level hooks member, keeping unrelated settings byte-for-byte."""
    decoder = json.JSONDecoder()
    pos = source.index("{") + 1
    previous_comma = None

    def skip(index):
        while source[index].isspace():
            index += 1
        return index

    while True:
        start = skip(pos)
        if source[start] == "}":
            if hooks is MISSING:
                return source
            comma = "," if source[source.index("{") + 1:start].strip() else ""
            return source[:start] + comma + '\n  "hooks": ' + json.dumps(hooks, indent=2) + "\n" + source[start:]
        key, key_end = decoder.raw_decode(source, start)
        val_start = skip(skip(key_end) + 1)
        _, val_end = decoder.raw_decode(source, val_start)
        after = skip(val_end)
        if key == "hooks":
            if hooks is not MISSING:
                return source[:val_start] + json.dumps(hooks, indent=2, ensure_ascii=False) + source[val_end:]
            if source[after] == ",":
                return source[:start] + source[after+1:]
            return source[:previous_comma if previous_comma is not None else start] + source[val_end:]
        if source[after] == "}":
            pos = after
        else:
            previous_comma = after
            pos = after + 1


def _edit(raw, entry, install, baseline):
    source = raw.decode("utf-8") if raw is not None else "{}\n"
    doc = json.loads(source, object_pairs_hook=_object)
    if not isinstance(doc, dict) or not isinstance(doc.get("hooks", {}), dict):
        raise ValueError("Expected a JSON object with an optional hooks object.")
    hooks = doc.get("hooks", {}).copy()
    before = json.loads(baseline) if baseline is not None else {}
    changed = False
    for event in EVENTS:
        groups = hooks.get(event, [])
        if not isinstance(groups, list):
            raise ValueError(f"Expected a list of {event} hook groups.")
        found = False
        remaining = []
        for group in groups:
            if not isinstance(group, dict) or not isinstance(group.get("hooks"), list):
                raise ValueError(f"Invalid {event} hook group.")
            handlers = group["hooks"]
            matches = [h for h in handlers if isinstance(h, dict) and h.get("command") == entry["command"]]
            if install:
                for other in handlers:
                    command = other.get("command") if isinstance(other, dict) else None
                    if isinstance(command, str) and command != entry["command"] and "profitctl_agent.design_hook" in command:
                        if shlex.split(command)[1:4] == ["-I", "-m", "profitctl_agent.design_hook"]:
                            raise ValueError("A different ProfitCtl hook is active. Uninstall it with its original arguments first.")
            if any(h != entry for h in matches):
                raise ValueError(f"The {event} ProfitCtl handler was edited. Inspect it before changing configuration.")
            found |= bool(matches)
            if matches and not install:
                changed = True
                others = [h for h in handlers if h != entry]
                if others:
                    remaining.append({**group, "hooks": others})
            else:
                remaining.append(group)
        if install and not found:
            remaining.append({"hooks": [entry]})
            changed = True
        if remaining or event in before.get("hooks", {}):
            hooks[event] = remaining
        else:
            hooks.pop(event, None)
    if not changed:
        return raw
    if not install and not hooks and "hooks" not in before:
        if set(doc) == {"hooks"} and baseline is None:
            return None
        return _set_hooks(source, MISSING).encode()
    return _set_hooks(source, hooks).encode()


def configure(mode, *, codex_config, claude_config, python, socket, routes, backup_dir):
    paths = [Path(p) for p in (codex_config, claude_config, python, socket, routes, backup_dir)]
    if any(not p.is_absolute() or ".." in p.parts or any(ord(c) < 32 for c in str(p)) for p in paths):
        raise ValueError("Supply absolute operator-selected paths.")
    codex, claude, interpreter, sock, route_file, backups = paths
    if codex == claude or mode not in ("install", "uninstall"):
        raise ValueError("Supply separate native configurations and an install or uninstall action.")
    if mode == "install":
        from .design_hook import _trusted_routes
        if not interpreter.is_file() or not os.access(interpreter, os.X_OK):
            raise ValueError("The isolated hook interpreter must be an installed executable.")
        _trusted_routes(route_file)
    command = shlex.join([str(interpreter), "-I", "-m", "profitctl_agent.design_hook",
                          "--socket", str(sock), "--routes", str(route_file)])
    entry = {"type": "command", "command": command, "timeout": 4}
    _directory(backups, private=True)
    for target in (codex, claude):
        _directory(target.parent)
    # Serialize installer operations for this exact pair of clients and hook definition.
    identity = hashlib.sha256(json.dumps([str(codex), str(claude), command]).encode()).hexdigest()[:24]
    state = backups / (identity + ".json")
    lockfd = os.open(backups/(identity + ".lock"), os.O_RDWR | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    with os.fdopen(lockfd, "rb") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        stored = _read(state)
        prior = json.loads(stored) if stored is not None else {}
        old_folder = prior.get("backup")
        if old_folder and (Path(old_folder).name != old_folder or not old_folder.startswith("install-")):
            raise ValueError("Invalid installation receipt.")
        edits = []
        for label, target in [("codex", codex), ("claude", claude)]:
            current = _read(target)
            baseline = _read(backups/old_folder/(label + ".before")) if old_folder else current
            updated = _edit(current, entry, mode == "install", baseline)
            if mode == "uninstall" and old_folder and prior.get(label) == hashlib.sha256(current or b"").hexdigest():
                updated = baseline
            if updated != current:
                edits.append((label, target, current, updated))
        if not edits:
            if mode == "uninstall" and stored is not None:
                _write(state, stored, None)
            return None
        snapshot = Path(tempfile.mkdtemp(prefix=mode+"-", dir=backups))
        receipt = {"backup": snapshot.name}
        for label, target in [("codex", codex), ("claude", claude)]:
            current = _read(target)
            if current is not None:
                _write(snapshot/(label+".before"), None, current)
            result = next((new for key, _, _, new in edits if key == label), current)
            receipt[label] = hashlib.sha256(result or b"").hexdigest()
        changed = []
        try:
            for _, target, old, new in edits:
                _write(target, old, new)
                changed.append((target, old, new))
            if mode == "install" and stored is None:
                _write(state, None, json.dumps(receipt).encode())
            elif mode == "uninstall" and stored is not None:
                _write(state, stored, None)
        except (ValueError, OSError):
            for target, old, new in reversed(changed):
                _write(target, new, old)
            raise
        return snapshot


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=["install", "uninstall"])
    for name in ("codex-config", "claude-config", "python", "socket", "routes", "backup-dir"):
        parser.add_argument("--" + name, type=Path, required=True)
    try:
        snapshot = configure(**vars(parser.parse_args()))
    except (OSError, ValueError, RecursionError) as exc:
        parser.exit(1, f"ProfitCtl could not update the hooks. {exc}\n")
    print(f"Saved configuration backups in {snapshot}" if snapshot else "Configuration already matches.")


if __name__ == "__main__":
    main()
