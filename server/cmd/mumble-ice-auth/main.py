#!/usr/bin/env python3
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "zeroc-ice",
# ]
# ///
"""
AmiyaEden Mumble ICE Authenticator.

This process registers a Murmur ICE authenticator and delegates credential
checks to AmiyaEden's internal `/api/v1/voice/mumble/ice-auth` endpoint.
Channel ACLs stay managed in Mumble; AmiyaEden only returns role-derived groups.
"""

from __future__ import annotations

import json
import logging
import os
import signal
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any

try:
    import Ice  # type: ignore
except ImportError as exc:  # pragma: no cover - runtime dependency check
    raise SystemExit(
        "Missing Python package `zeroc-ice`. Install it with: pip install zeroc-ice"
    ) from exc


def env(name: str, default: str = "") -> str:
    return os.environ.get(name, default).strip()


@dataclass
class Settings:
    amiya_base_url: str
    amiya_auth_secret: str
    amiya_timeout: float
    murmur_proxy: str
    murmur_secret: str
    murmur_server_id: int
    adapter_name: str
    adapter_endpoints: str
    slice_file: str

    @classmethod
    def from_env(cls) -> "Settings":
        return cls(
            amiya_base_url=env("AMIYA_EDEN_BASE_URL", "http://127.0.0.1:8080/api/v1").rstrip("/"),
            amiya_auth_secret=env("AMIYA_MUMBLE_AUTH_SECRET"),
            amiya_timeout=float(env("AMIYA_EDEN_TIMEOUT", "3")),
            murmur_proxy=env("MURMUR_ICE_PROXY", "Meta:tcp -h 127.0.0.1 -p 6502"),
            murmur_secret=env("MURMUR_ICE_SECRET"),
            murmur_server_id=int(env("MURMUR_SERVER_ID", "1")),
            adapter_name=env("MUMBLE_AUTH_ADAPTER_NAME", "AmiyaEdenMumbleAuth"),
            adapter_endpoints=env("MUMBLE_AUTH_ENDPOINTS", "tcp -h 127.0.0.1 -p 6503"),
            slice_file=env("MUMBLE_SLICE_FILE", ""),
        )


def find_slice_file(configured: str) -> str:
    if configured:
        return configured

    candidates = [
        "/usr/share/slice/Murmur.ice",
        "/usr/share/slice/mumble/Murmur.ice",
        "/usr/share/mumble-server/Murmur.ice",
        "/usr/share/slice/MumbleServer.ice",
        "/usr/share/slice/mumble/MumbleServer.ice",
        "/usr/share/mumble-server/MumbleServer.ice",
    ]
    for path in candidates:
        if os.path.exists(path):
            return path
    raise SystemExit(
        "Unable to locate Murmur/MumbleServer ICE slice. Set MUMBLE_SLICE_FILE=/path/to/MumbleServer.ice"
    )


def load_murmur_slice(slice_file: str) -> Any:
    include_dirs = []
    if hasattr(Ice, "getSliceDir"):
        include_dirs.append("-I" + Ice.getSliceDir())
    include_dirs.append("-I" + os.path.dirname(slice_file))
    Ice.loadSlice("", [*include_dirs, slice_file])

    try:
        import Murmur  # type: ignore
    except ImportError as exc:  # pragma: no cover - runtime dependency check
        raise SystemExit(f"Loaded {slice_file}, but Python module `Murmur` was not generated") from exc
    return Murmur


class AmiyaClient:
    def __init__(self, settings: Settings):
        if not settings.amiya_auth_secret:
            raise SystemExit("AMIYA_MUMBLE_AUTH_SECRET is required")
        self.settings = settings

    def authenticate(self, username: str, password: str) -> dict[str, Any]:
        url = self.settings.amiya_base_url + "/voice/mumble/ice-auth"
        body = json.dumps({"username": username, "password": password}).encode("utf-8")
        req = urllib.request.Request(
            url,
            data=body,
            headers={
                "Content-Type": "application/json",
                "X-Mumble-Auth-Secret": self.settings.amiya_auth_secret,
            },
            method="POST",
        )
        try:
            with urllib.request.urlopen(req, timeout=self.settings.amiya_timeout) as resp:
                payload = json.loads(resp.read().decode("utf-8"))
        except urllib.error.HTTPError as exc:
            logging.warning("AmiyaEden auth endpoint rejected request: status=%s", exc.code)
            return {"allowed": False}
        except Exception:
            logging.exception("AmiyaEden auth endpoint call failed")
            return {"allowed": False}

        if payload.get("code") != 200:
            logging.warning("AmiyaEden auth endpoint returned code=%s msg=%s", payload.get("code"), payload.get("msg"))
            return {"allowed": False}
        data = payload.get("data") or {}
        if not data.get("allowed"):
            return {"allowed": False}
        return data


class AmiyaAuthenticator:
    def __init__(self, murmur: Any, client: AmiyaClient):
        self._murmur = murmur
        self._client = client
        self._server = None
        self._cache: dict[int, dict[str, Any]] = {}

    def setServer(self, server: Any, current: Any = None) -> None:
        self._server = server
        logging.info("Murmur server reference received")

    def authenticate(
        self,
        name: str,
        pw: str,
        certlist: list[str],
        certhash: str,
        *args: Any,
    ) -> tuple[int, str, list[str]]:
        data = self._client.authenticate(name, pw)
        if not data.get("allowed"):
            logging.info("Rejected Mumble login for username=%s", name)
            return (-1, "", [])

        user_id = int(data["user_id"])
        display_name = str(data.get("display_name") or name)
        groups = [str(group) for group in data.get("groups", []) if group]
        self._cache[user_id] = {
            "display_name": display_name,
            "groups": groups,
            "last_auth_at": time.time(),
        }
        logging.info("Accepted Mumble login user_id=%s display_name=%s groups=%s", user_id, display_name, groups)
        return (user_id, display_name, groups)

    def getInfo(self, user_id: int, current: Any = None) -> tuple[bool, dict[int, str]]:
        cached = self._cache.get(int(user_id))
        if not cached:
            return (False, {})
        info_key = getattr(self._murmur.UserInfo, "UserName", 0)
        return (True, {info_key: cached["display_name"]})

    def idToName(self, user_id: int, current: Any = None) -> str:
        cached = self._cache.get(int(user_id))
        return str(cached.get("display_name", "")) if cached else ""

    def nameToId(self, name: str, current: Any = None) -> int:
        try:
            return int(name)
        except ValueError:
            return -2

    def idToTexture(self, user_id: int, current: Any = None) -> bytes:
        return b""

    # Registration APIs are intentionally disabled. AmiyaEden is the identity source.
    def registerUser(self, info: dict[int, str], current: Any = None) -> int:
        return -1

    def unregisterUser(self, user_id: int, current: Any = None) -> int:
        return -1

    def getRegisteredUsers(self, filter: str, current: Any = None) -> dict[int, str]:
        return {}

    def setInfo(self, user_id: int, info: dict[int, str], current: Any = None) -> int:
        return -1

    def setTexture(self, user_id: int, texture: bytes, current: Any = None) -> int:
        return -1


def main() -> int:
    logging.basicConfig(
        level=os.environ.get("LOG_LEVEL", "INFO").upper(),
        format="%(asctime)s %(levelname)s %(message)s",
    )
    settings = Settings.from_env()
    murmur = load_murmur_slice(find_slice_file(settings.slice_file))

    communicator = Ice.initialize(sys.argv)
    shutdown_requested = False
    try:
        if settings.murmur_secret:
            communicator.getImplicitContext().put("secret", settings.murmur_secret)

        meta = murmur.MetaPrx.checkedCast(communicator.stringToProxy(settings.murmur_proxy))
        if not meta:
            raise RuntimeError(f"Unable to connect Murmur ICE proxy: {settings.murmur_proxy}")

        server = meta.getServer(settings.murmur_server_id)
        if not server:
            raise RuntimeError(f"Murmur virtual server not found: {settings.murmur_server_id}")

        adapter = communicator.createObjectAdapterWithEndpoints(settings.adapter_name, settings.adapter_endpoints)
        servant_base = getattr(
            murmur,
            "ServerUpdatingAuthenticator",
            getattr(murmur, "ServerAuthenticator", object),
        )
        servant_cls = type("AmiyaServerUpdatingAuthenticator", (AmiyaAuthenticator, servant_base), {})
        servant = servant_cls(murmur, AmiyaClient(settings))

        proxy = adapter.addWithUUID(servant)
        auth_proxy = murmur.ServerAuthenticatorPrx.uncheckedCast(proxy)
        adapter.activate()
        server.setAuthenticator(auth_proxy)

        logging.info(
            "AmiyaEden Mumble ICE authenticator registered: server_id=%s adapter=%s endpoints=%s",
            settings.murmur_server_id,
            settings.adapter_name,
            settings.adapter_endpoints,
        )

        def shutdown(signum: int, frame: Any) -> None:
            nonlocal shutdown_requested
            if not shutdown_requested:
                shutdown_requested = True
                logging.info("Shutdown requested")
                communicator.shutdown()

        signal.signal(signal.SIGINT, shutdown)
        signal.signal(signal.SIGTERM, shutdown)
        communicator.waitForShutdown()
        return 0
    finally:
        communicator.destroy()


if __name__ == "__main__":
    raise SystemExit(main())
