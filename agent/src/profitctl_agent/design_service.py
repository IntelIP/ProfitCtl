"""A private POSIX socket for deterministic design reviews and durable investigations."""

import errno
import fcntl
import json
import os
import re
import shutil
import signal
import socket
import stat
import subprocess
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from contextlib import contextmanager
from pathlib import Path


MAX_REQUEST = 128 * 1024
MAX_RESPONSE = 4 * 1024 * 1024
READ_TIMEOUT = 3
WORKERS = 8
_PROFILE = re.compile(r"[a-z0-9][a-z0-9-]{0,63}")
_INVESTIGATION_FIELDS = {
    "start": {"run_id", "case_id", "question"},
    "status": {"run_id"},
    "resume": {"run_id", "expected_revision", "action"},
}


class ServiceError(ValueError):
    """A message safe to return to a local client."""


def _encode(value, limit):
    try:
        data = json.dumps(value, allow_nan=False, separators=(",", ":"), ensure_ascii=True).encode("ascii")
    except (ValueError, TypeError, RecursionError):
        raise ServiceError("Supply a finite JSON object.") from None
    if len(data) > limit:
        raise ServiceError("The local service message exceeds the size limit.")
    return data


def _decode(data):
    def distinct_keys(pairs):
        value = dict(pairs)
        if len(value) != len(pairs):
            raise ValueError
        return value

    def invalid_constant(_):
        raise ValueError

    try:
        value = json.loads(data.decode("utf-8"), object_pairs_hook=distinct_keys, parse_constant=invalid_constant)
    except (ValueError, RecursionError):
        raise ServiceError("Supply a valid UTF-8 JSON object.") from None
    if not isinstance(value, dict):
        raise ServiceError("Supply a JSON object.")
    return value


def _write(conn, value, limit):
    data = _encode(value, limit)
    conn.sendall(len(data).to_bytes(4) + data)


def _read(conn, limit, deadline):
    def exact(size):
        chunk = bytearray()
        while len(chunk) < size:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError
            conn.settimeout(remaining)
            data = conn.recv(min(size - len(chunk), 65536))
            if not data:
                raise ServiceError("The local service message was incomplete.")
            chunk.extend(data)
        return chunk

    size = int.from_bytes(exact(4))
    if not 0 < size <= limit:
        raise ServiceError("The local service message exceeds the size limit.")
    return _decode(exact(size))


def _private(info, kind):
    if not kind(info.st_mode) or info.st_uid != os.getuid() or info.st_mode & 0o077:
        raise ServiceError("The service needs a private location owned by the current user.")


def request(socket_path, payload, timeout=420):
    """Send one bounded request over the operator's private Unix socket and return its result."""
    path = Path(socket_path)
    deadline = time.monotonic() + timeout
    try:
        _private(path.parent.lstat(), stat.S_ISDIR)
        _private(path.lstat(), stat.S_ISSOCK)
        with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as conn:
            conn.settimeout(timeout)
            conn.connect(os.fspath(path))
            _write(conn, payload, MAX_REQUEST)
            reply = _read(conn, MAX_RESPONSE, deadline)
    except OSError:
        raise OSError("The local ProfitCtl service is unavailable or timed out.") from None
    if isinstance(reply.get("error"), str):
        raise ServiceError(reply["error"])
    if not isinstance(reply.get("result"), dict):
        raise ServiceError("The local ProfitCtl service returned an invalid reply.")
    return reply["result"]


class SocketClient:
    def __init__(self, path):
        self.path = path

    def _call(self, payload):
        try:
            return request(self.path, payload)
        except OSError as exc:
            raise ServiceError(str(exc)) from None

    def start(self, **arguments):
        return self._call({"action": "start", "arguments": arguments})

    def status(self, **arguments):
        return self._call({"action": "status", "arguments": arguments})

    def resume(self, **arguments):
        return self._call({"action": "resume", "arguments": arguments})

    def review(self, *, profile=None, proposal=None):
        if (profile is None) == (proposal is None):
            raise ServiceError("Supply exactly one registered profile or inline proposal.")
        subject = {"profile": profile} if profile is not None else {"proposal": proposal}
        return self._call({"action": "review", **subject})


@contextmanager
def _private_directory(path):
    fd = os.open(path, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        _private(os.fstat(fd), stat.S_ISDIR)
        yield fd
    finally:
        os.close(fd)


def _profile(directory, name):
    if not isinstance(name, str) or not _PROFILE.fullmatch(name):
        raise ServiceError("Supply a registered profile slug of 1 to 64 lowercase letters, digits or hyphens.")
    try:
        fd = os.open(name + ".json", os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=directory)
        with os.fdopen(fd, "rb") as source:
            info = os.fstat(fd)
            _private(info, stat.S_ISREG)
            if info.st_nlink != 1 or info.st_size > MAX_REQUEST:
                raise ServiceError("The registered profile must be a bounded private regular file without links.")
            data = source.read(MAX_REQUEST + 1)
            if len(data) > MAX_REQUEST:
                raise ServiceError("The registered profile exceeds the size limit.")
    except OSError:
        raise ServiceError("The registered design profile is unavailable.") from None
    return _decode(data)


def _review(binary, workspace, proposal):
    if not isinstance(proposal, dict):
        raise ServiceError("Supply a self-contained design proposal object.")
    data = _encode(proposal, MAX_REQUEST)
    try:
        run = subprocess.run([binary, "system", "design", "-"], cwd=workspace, input=data,
                             stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, timeout=20,
                             env={"PATH": os.defpath, "LANG": "C"}, check=False)
    except (OSError, subprocess.TimeoutExpired):
        raise ServiceError("The local ProfitCtl design calculator could not complete.") from None
    if run.returncode not in (0, 1):
        raise ServiceError("ProfitCtl rejected the proposal. Supply a valid self-contained profitctl.design/v1 object.")
    if len(run.stdout) > MAX_RESPONSE:
        raise ServiceError("The design report exceeds the local service size limit.")
    try:
        report = _decode(run.stdout)
    except ServiceError:
        raise ServiceError("The configured calculator returned an invalid design report.") from None
    expected = "incomplete" if run.returncode else "complete"
    if (report.get("schema_version") != "profitctl.design-report/v1" or report.get("status") != expected
            or not isinstance(report.get("summary"), str) or not report["summary"].strip()):
        raise ServiceError("The configured calculator returned an invalid design report.")
    return report


def _dispatch(payload, workspace, binary, profiles):
    action = payload.get("action")
    if action == "health" and set(payload) == {"action"}:
        return {"status": "ok", "pid": os.getpid()}
    if action == "review":
        if set(payload) == {"action", "profile"}:
            proposal = _profile(profiles, payload["profile"])
        elif set(payload) == {"action", "proposal"}:
            proposal = payload["proposal"]
        else:
            raise ServiceError("Supply exactly one registered profile or inline proposal.")
        return _review(binary, workspace, proposal)
    if not isinstance(action, str) or action not in _INVESTIGATION_FIELDS:
        raise ServiceError("Choose a supported local service action.")
    args = payload.get("arguments")
    required = _INVESTIGATION_FIELDS[action]
    allowed = required | ({"evidence"} if action == "resume" else set())
    if (set(payload) != {"action", "arguments"} or not isinstance(args, dict)
            or not required <= args.keys() <= allowed):
        raise ServiceError("Supply the required structured investigation arguments.")
    from .economics import CaseNotFound, RevisionConflict
    from .runtime import InvestigationError, InvestigationService

    try:
        investigator = InvestigationService(workspace, profitctl_binary=binary)
        return getattr(investigator, action)(**args)
    except (InvestigationError, CaseNotFound, RevisionConflict) as exc:
        raise ServiceError(str(exc)) from None
    except Exception:
        raise ServiceError("Investigation could not finish. Check the case, evidence and model connection. "
                           "Read its saved status before retrying with the same run ID.") from None


@contextmanager
def _listener(path):
    # The persistent lock serializes bind, stale socket replacement and unlink across restarts.
    path = Path(path).absolute()
    with _private_directory(path.parent) as parent:
        lock = os.open(path.name + ".lock", os.O_RDWR | os.O_CREAT | os.O_NOFOLLOW | os.O_NONBLOCK,
                       0o600, dir_fd=parent)
        try:
            info = os.fstat(lock)
            _private(info, stat.S_ISREG)
            if info.st_nlink != 1:
                raise ServiceError("The local service lock must be a private file without links.")
            try:
                fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                raise ServiceError("A local service already owns this socket.") from None
            if os.path.lexists(path):
                info = path.lstat()
                if not stat.S_ISSOCK(info.st_mode) or info.st_uid != os.getuid():
                    raise ServiceError("The requested socket path is already occupied.")
                with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as probe:
                    probe.settimeout(0.25)
                    if probe.connect_ex(os.fspath(path)) != errno.ECONNREFUSED:
                        raise ServiceError("The requested socket path already serves another listener.")
                if not os.path.samestat(info, path.lstat()):
                    raise ServiceError("The requested socket path changed during startup.")
                path.unlink()
            with socket.socket(socket.AF_UNIX, socket.SOCK_STREAM) as listener:
                listener.bind(os.fspath(path))
                bound = path.lstat()
                try:
                    path.chmod(0o600)
                    listener.listen(WORKERS)
                    listener.settimeout(0.4)
                    yield listener
                finally:
                    try:
                        if os.path.samestat(bound, path.lstat()):
                            path.unlink()
                    except FileNotFoundError:
                        pass
        finally:
            os.close(lock)


def _handle(conn, workspace, binary, profiles):
    with conn:
        try:
            payload = _read(conn, MAX_REQUEST, time.monotonic() + READ_TIMEOUT)
            reply = {"result": _dispatch(payload, workspace, binary, profiles)}
            data = _encode(reply, MAX_RESPONSE)
        except ServiceError as exc:
            data = _encode({"error": str(exc)}, MAX_RESPONSE)
        except (TimeoutError, OSError):
            return
        except Exception:
            data = _encode({"error": "The local service could not complete the request."}, MAX_RESPONSE)
        try:
            conn.settimeout(READ_TIMEOUT)
            conn.sendall(len(data).to_bytes(4) + data)
        except OSError:
            pass


def serve(workspace, profitctl_binary, socket_path, profiles):
    """Run the local service in the foreground until its supervisor requests shutdown."""
    binary = shutil.which(os.fspath(profitctl_binary))
    if binary is None:
        raise ServiceError("The configured ProfitCtl executable is unavailable.")
    workspace = Path(workspace).resolve(strict=True)
    if not workspace.is_dir():
        raise ServiceError("The configured workspace must be a directory.")
    binary = str(Path(binary).resolve())
    stopping = threading.Event()
    slots = threading.BoundedSemaphore(WORKERS)
    with _private_directory(profiles) as directory, _listener(socket_path) as listener, ThreadPoolExecutor(WORKERS) as workers:
        previous = {sig: signal.signal(sig, lambda *_: stopping.set()) for sig in (signal.SIGTERM, signal.SIGINT)}
        try:
            while not stopping.is_set():
                try:
                    conn, _ = listener.accept()
                except TimeoutError:
                    continue
                if slots.acquire(blocking=False):
                    workers.submit(_handle, conn, workspace, binary, directory).add_done_callback(lambda _: slots.release())
                else:
                    with conn:
                        try:
                            conn.settimeout(0.1)
                            _write(conn, {"error": "The local service is busy. Retry shortly."}, MAX_RESPONSE)
                        except OSError:
                            pass
        finally:
            for sig, handler in previous.items():
                signal.signal(sig, handler)
