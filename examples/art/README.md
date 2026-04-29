# art — agrirouter G4 CLI

`art` is a small command-line tool built on top of [`agrirouter-sdk-go`](../..)
that exercises the agrirouter G4 API. It is useful as a working example of
the SDK and as a debugging client during development.

## Build & run

```bash
# from this directory (examples/art/)
go build -o art .
./art --help

# or run directly
go run . --help
```

## Authentication

`art` authenticates against agrirouter using the OAuth 2.0 client-credentials
flow. Set the client ID and secret in the environment before running any
command that talks to the API:

```bash
export AGRIROUTER_OAUTH_CLIENT_ID=...
export AGRIROUTER_OAUTH_CLIENT_SECRET=...
```

The token endpoint and the API server URL default to QA
(`https://oauth.qa.agrirouter.farm/token`, `https://api.qa.agrirouter.farm`).
Override them via `ART_OAUTH_TOKEN_URL` and `ART_API_URL` to point at a
different environment (e.g. a local gateway at `http://localhost:8081`).

## Environment variables

| Variable | Purpose | Used by |
|---|---|---|
| `AGRIROUTER_OAUTH_CLIENT_ID` | OAuth 2.0 client ID for the agrirouter application. Required. | All commands that call the API; `serve` also uses it to construct the user-facing authorize URL. |
| `AGRIROUTER_OAUTH_CLIENT_SECRET` | OAuth 2.0 client secret. Required. | All commands that call the API. |
| `ART_TENANT_ID` | Default tenant UUID, used when `--tenant-id` / `-t` is not passed. | `put-endpoint`, `delete-endpoint`, `send-messages`, `confirm-messages`, `list-tenant-endpoints`. The `repl` builtin `set-tenant <uuid>` updates this. |
| `ART_APPLICATION_ID` | Default application UUID, used when `--application-id` is not passed. | `put-endpoint`, `serve`. |
| `ART_SOFTWARE_VERSION_ID` | Default software-version UUID, used when `--software-version-id` is not passed. | `put-endpoint`, `serve`. |
| `ART_API_URL` | Override the agrirouter API base URL. Defaults to `https://api.qa.agrirouter.farm`. Set to e.g. `http://localhost:8081` to test against a local gateway. | All API calls. |
| `ART_OAUTH_TOKEN_URL` | Override the OAuth 2.0 token URL. Defaults to `https://oauth.qa.agrirouter.farm/token`. | Token acquisition. |

For each `ART_*` variable the corresponding flag wins if it is provided;
otherwise the env var is read. If neither is set and the value is required,
the command errors out.

### `.env` file

On startup, `art` loads a `.env` file from the current working directory if
one exists (using [`subosito/gotenv`](https://github.com/subosito/gotenv),
the same parser viper uses). Variables already present in the process
environment are **not** overridden, so real env vars always win over the
file. A missing `.env` is silent.

Copy `.env.example` to `.env` and fill in the values to get started:

```bash
cp .env.example .env
$EDITOR .env
```

The `repl` subcommand expands `$ART_*` references (e.g. `$ART_TENANT_ID`) in
its input lines before executing them and offers tab completion for any
`ART_*` env var currently set.

## Commands

Run `art <command> --help` for full flag listings. At a glance:

### Endpoint management
- `put-endpoint` — create or update an endpoint by external ID.
- `delete-endpoint` — delete an endpoint by external ID.
- `list-authorized-tenants` (alias `lat`) — list all tenants the application
  is authorized for, together with their visible endpoints.
- `list-tenant-endpoints` (alias `lte`) — list endpoints of a single tenant,
  including capabilities and route-derived send/receive maps for
  application-owned endpoints.

### Messaging
- `send-messages` — send a message payload to agrirouter.
- `receive-messages` — stream `MESSAGE_RECEIVED` events; optionally save
  payloads to disk with `--save-payloads-to <dir>`.
- `confirm-messages` — confirm one or more received messages.

### Events
- `receive-events` (alias `re`) — generic SSE listener. Without `--types`
  receives every event kind; with `--types` (comma-separated or repeated)
  filters server-side. Valid types: `MESSAGE_RECEIVED`, `FILE_RECEIVED`,
  `ENDPOINT_DELETED`, `ENDPOINTS_LIST_CHANGED`, `AUTHORIZATION_ADDED`,
  `AUTHORIZATION_REVOKED`.
- `receive-endpoint-deleted-events` (alias `rede`) — `ENDPOINT_DELETED` only.
- `receive-endpoints-list-changed-events` (alias `relc`) — `ENDPOINTS_LIST_CHANGED` only.
- `receive-authorization-added-events` (alias `raa`) — `AUTHORIZATION_ADDED` only.
- `receive-authorization-revoked-events` (alias `rar`) — `AUTHORIZATION_REVOKED` only.

The dedicated single-type commands are thin wrappers around the same SSE
stream `receive-events` uses; pick whichever reads more clearly in your
workflow.

### Interactive / web
- `repl` — sandboxed interactive shell that re-invokes `art` for each line,
  with tab completion for subcommands, flags and `$ART_*` env vars. Builtins:
  `clear`, `exit`, `set-tenant <uuid>` (sets `$ART_TENANT_ID` for the session).
- `serve` — minimal web UI on `http://localhost:8080` with the agrirouter
  authorization flow, an endpoint-validation form, and an embedded xterm.js
  terminal that runs `art repl`. Flags: `--port`, `--application-id`,
  `--software-version-id`.

## Typical session

```bash
export AGRIROUTER_OAUTH_CLIENT_ID=...
export AGRIROUTER_OAUTH_CLIENT_SECRET=...
export ART_APPLICATION_ID=00000000-0000-0000-0000-000000000000
export ART_SOFTWARE_VERSION_ID=00000000-0000-0000-0000-000000000000
export ART_TENANT_ID=00000000-0000-0000-0000-000000000000

# Inspect what we currently see
art list-authorized-tenants
art list-tenant-endpoints

# Stream every event type in one terminal
art receive-events

# Or stream just a subset
art receive-events --types ENDPOINTS_LIST_CHANGED,AUTHORIZATION_ADDED

# In another terminal, create an endpoint
art put-endpoint \
  --external-id urn:my-app:endpoint:demo \
  --endpoint-type cloud_software \
  --with-capability iso:11783:-10:taskdata:zip=SEND_RECEIVE
```
