#!/usr/bin/env bash
set -euo pipefail

UNIT="openclaw-gateway.service"
UPSTREAM_UNIT="openclaw-gateway-upstream.service"
LISTEN_HOST="${GATEWAY_HOST:-127.0.0.1}"
LISTEN_PORT="${GATEWAY_PORT:-18789}"
PROXY_LOG_FILE="${OPENCLAW_PROXY_LOG_FILE:-$HOME/.local/state/openclaw/openclaw-gateway-go-proxy.log}"

exit_code=0

echo "== service activity =="
if systemctl --user is-active --quiet "$UNIT"; then
  echo "$UNIT: active"
else
  echo "$UNIT: inactive"
  exit_code=1
fi

if systemctl --user is-active --quiet "$UPSTREAM_UNIT"; then
  echo "$UPSTREAM_UNIT: active"
else
  echo "$UPSTREAM_UNIT: inactive (expected only after replacement)"
fi

echo
echo "== health endpoint =="
if curl -fsS "http://$LISTEN_HOST:$LISTEN_PORT/health" >/dev/null 2>&1; then
  echo "http://$LISTEN_HOST:$LISTEN_PORT/health: ok"
else
  echo "http://$LISTEN_HOST:$LISTEN_PORT/health: failed"
  exit_code=1
fi

echo
echo "== recent journal ($UNIT, 20 lines) =="
journalctl --user -u "$UNIT" -n 20 --no-pager || true

echo
echo "== recent journal ($UPSTREAM_UNIT, 20 lines) =="
journalctl --user -u "$UPSTREAM_UNIT" -n 20 --no-pager || true

if [[ -f "$PROXY_LOG_FILE" ]]; then
  echo
  echo "== recent proxy log file (20 lines) =="
  tail -n 20 "$PROXY_LOG_FILE" || true
fi

exit "$exit_code"
