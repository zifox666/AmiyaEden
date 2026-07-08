# AmiyaEden Mumble ICE Authenticator

This process connects to Murmur through ICE, registers an external authenticator, and delegates login checks to AmiyaEden.

Scope:

- AmiyaEden authenticates `username = user_id` and the generated Mumble password.
- AmiyaEden returns the display name and role-derived Mumble groups.
- Mumble channel ACLs remain managed inside Mumble.

## Requirements

```bash
pip install zeroc-ice
```

The host also needs the Murmur slice file, usually one of:

- `/usr/share/slice/Murmur.ice`
- `/usr/share/slice/mumble/MumbleServer.ice`
- `/usr/share/mumble-server/MumbleServer.ice`

If it is elsewhere, set `MUMBLE_SLICE_FILE`.

## AmiyaEden Setup

Set the same shared secret in the AmiyaEden Mumble admin page:

- Voice Center -> Mumble -> Mumble Configuration -> ICE Shared Secret

The authenticator sends this value as `X-Mumble-Auth-Secret` to:

```text
POST /api/v1/voice/mumble/ice-auth
```

## Murmur Setup

Enable ICE in `murmur.ini` / `mumble-server.ini`:

```ini
ice="tcp -h 127.0.0.1 -p 6502"
icesecret="change-this-murmur-ice-secret"
```

Restart Murmur after changing ICE settings.

## Run

```bash
cd server/cmd/mumble-ice-auth
AMIYA_EDEN_BASE_URL="http://127.0.0.1:8080/api/v1" \
AMIYA_MUMBLE_AUTH_SECRET="same-secret-as-admin-page" \
MURMUR_ICE_PROXY="Meta:tcp -h 127.0.0.1 -p 6502" \
MURMUR_ICE_SECRET="change-this-murmur-ice-secret" \
MURMUR_SERVER_ID="1" \
MUMBLE_AUTH_ENDPOINTS="tcp -h 127.0.0.1 -p 6503" \
python3 main.py
```

## Environment

| Variable | Default | Description |
| --- | --- | --- |
| `AMIYA_EDEN_BASE_URL` | `http://127.0.0.1:8080/api/v1` | AmiyaEden API base URL |
| `AMIYA_MUMBLE_AUTH_SECRET` | required | Shared secret configured in AmiyaEden |
| `AMIYA_EDEN_TIMEOUT` | `3` | Backend request timeout in seconds |
| `MURMUR_ICE_PROXY` | `Meta:tcp -h 127.0.0.1 -p 6502` | Murmur ICE Meta proxy |
| `MURMUR_ICE_SECRET` | empty | Murmur `icesecret` |
| `MURMUR_SERVER_ID` | `1` | Murmur virtual server ID |
| `MUMBLE_AUTH_ADAPTER_NAME` | `AmiyaEdenMumbleAuth` | ICE object adapter name |
| `MUMBLE_AUTH_ENDPOINTS` | `tcp -h 127.0.0.1 -p 6503` | Callback endpoint Murmur can reach |
| `MUMBLE_SLICE_FILE` | auto-detect | Path to `Murmur.ice` / `MumbleServer.ice` |

## Mumble Group Mapping

Configure role -> group mapping in AmiyaEden's Mumble page. During login the authenticator returns those groups from `authenticate(...)`. Then manage channel ACLs in Mumble using those group names.
