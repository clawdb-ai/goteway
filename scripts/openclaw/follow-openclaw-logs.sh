#!/usr/bin/env bash
set -euo pipefail

UNIT="openclaw-gateway.service"
UPSTREAM_UNIT="openclaw-gateway-upstream.service"
PROXY_LOG_FILE="${OPENCLAW_PROXY_LOG_FILE:-$HOME/.local/state/openclaw/openclaw-gateway-go-proxy.log}"

echo "Following journal logs for $UNIT and $UPSTREAM_UNIT (Ctrl+C to stop)..."
journalctl --user -f -u "$UNIT" -u "$UPSTREAM_UNIT" -o short-iso &
JOURNAL_PID=$!

if [[ -f "$PROXY_LOG_FILE" ]]; then
  echo "Also tailing proxy file log: $PROXY_LOG_FILE"
  tail -f "$PROXY_LOG_FILE" &
  TAIL_PID=$!
else
  TAIL_PID=""
fi

cleanup() {
  kill "$JOURNAL_PID" >/dev/null 2>&1 || true
  if [[ -n "$TAIL_PID" ]]; then
    kill "$TAIL_PID" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

wait
