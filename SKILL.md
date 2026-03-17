---
name: goteway-goat-gateway
description: Use this skill when users want to replace the local OpenClaw gateway with goteway's Go implementation, verify runtime health, follow debug logs, or rollback safely.
---

# goteway-goat-gateway

This skill performs machine-local OpenClaw gateway replacement using goteway scripts in this repository.

## Use When

- User asks to replace OpenClaw gateway with the Go gateway.
- User asks to verify gateway working status after replacement.
- User asks to collect/follow gateway logs for debugging.
- User asks to rollback to the original Node gateway service.

## Preconditions

- Linux with `systemd --user`.
- `openclaw-gateway.service` exists.
- Network access available for first-time local Go toolchain download.

## Standard Workflow

1. Install or refresh Go replacement:
```bash
./scripts/openclaw/replace-local-openclaw-gateway.sh
```

2. Verify current status:
```bash
./scripts/openclaw/status-openclaw-gateway.sh
```

3. Follow logs for debugging:
```bash
./scripts/openclaw/follow-openclaw-logs.sh
```

4. Rollback if requested:
```bash
./scripts/openclaw/rollback-local-openclaw-gateway.sh
```

## Environment Overrides

- `GO_VERSION` (default `1.26.1`)
- `GO_BIN`
- `GATEWAY_HOST` (default `127.0.0.1`)
- `GATEWAY_PORT` (default `18789`)
- `OPENCLAW_UPSTREAM_PORT` (default `18790`)
- `OPENCLAW_PROXY_LOG_FILE`
- `OPENCLAW_PROXY_LOG_HEADERS` (`true|false`)

## Safety Rules

- Operate only on `openclaw-gateway.service` and `openclaw-gateway-upstream.service`.
- Do not edit unrelated services or global system configuration.
- Keep backup/rollback path intact before restarts.
