#!/usr/bin/env bash
set -euo pipefail

UNIT="openclaw-gateway.service"
UPSTREAM_UNIT="openclaw-gateway-upstream.service"
SERVICE_DIR="$HOME/.config/systemd/user"
DROPIN_FILE="$SERVICE_DIR/${UNIT}.d/10-go-replacement.conf"
UPSTREAM_FILE="$SERVICE_DIR/$UPSTREAM_UNIT"

echo "[1/4] stop upstream passthrough service"
systemctl --user disable --now "$UPSTREAM_UNIT" >/dev/null 2>&1 || true

echo "[2/4] remove local override files"
rm -f "$DROPIN_FILE"
rm -f "$UPSTREAM_FILE"

echo "[3/4] reload units"
systemctl --user daemon-reload

echo "[4/4] restart $UNIT"
systemctl --user restart "$UNIT"
systemctl --user is-active --quiet "$UNIT"

echo "Rollback complete. $UNIT is active with the original configuration."
