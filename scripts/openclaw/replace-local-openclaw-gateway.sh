#!/usr/bin/env bash
set -euo pipefail

UNIT="openclaw-gateway.service"
UPSTREAM_UNIT="openclaw-gateway-upstream.service"
UPSTREAM_PORT="${OPENCLAW_UPSTREAM_PORT:-18790}"
LISTEN_HOST="${GATEWAY_HOST:-127.0.0.1}"
LISTEN_PORT="${GATEWAY_PORT:-18789}"
GO_VERSION="${GO_VERSION:-1.26.1}"
GO_BIN="${GO_BIN:-$HOME/.local/toolchains/go${GO_VERSION}/bin/go}"
PROXY_BIN="${OPENCLAW_PROXY_BIN:-$HOME/.local/bin/openclaw-gateway-go-proxy}"
PROXY_LOG_FILE="${OPENCLAW_PROXY_LOG_FILE:-$HOME/.local/state/openclaw/openclaw-gateway-go-proxy.log}"
PROXY_LOG_HEADERS="${OPENCLAW_PROXY_LOG_HEADERS:-false}"
SERVICE_DIR="$HOME/.config/systemd/user"
DROPIN_DIR="$SERVICE_DIR/${UNIT}.d"
DROPIN_FILE="$DROPIN_DIR/10-go-replacement.conf"
UPSTREAM_FILE="$SERVICE_DIR/$UPSTREAM_UNIT"
STATE_DIR="$HOME/.local/state/openclaw-go-replacement"
BACKUP_DIR="$STATE_DIR/backups/$(date +%Y%m%d-%H%M%S)"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required command: $1" >&2
    exit 1
  fi
}

need_cmd systemctl
need_cmd curl

install_local_go() {
  if [[ -x "$GO_BIN" ]]; then
    return
  fi

  local go_arch
  case "$(uname -m)" in
    x86_64) go_arch="amd64" ;;
    aarch64|arm64) go_arch="arm64" ;;
    *)
      echo "unsupported CPU arch for local Go install: $(uname -m)" >&2
      exit 1
      ;;
  esac

  local dest_dir="$HOME/.local/toolchains/go${GO_VERSION}"
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' RETURN

  echo "Go toolchain not found, installing go${GO_VERSION} locally..."
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${go_arch}.tar.gz" -o "$tmp_dir/go.tgz"
  mkdir -p "$dest_dir"
  tar -xzf "$tmp_dir/go.tgz" -C "$dest_dir" --strip-components=1

  if [[ ! -x "$GO_BIN" ]]; then
    echo "failed to install Go toolchain at $GO_BIN" >&2
    exit 1
  fi
}

if [[ ! -x "$GO_BIN" ]]; then
  install_local_go
fi

FRAGMENT_PATH="$(systemctl --user show "$UNIT" -p FragmentPath --value)"
if [[ -z "$FRAGMENT_PATH" || ! -f "$FRAGMENT_PATH" ]]; then
  echo "unable to locate $UNIT fragment path" >&2
  exit 1
fi

echo "[1/7] backup current unit to $BACKUP_DIR"
mkdir -p "$BACKUP_DIR"
cp "$FRAGMENT_PATH" "$BACKUP_DIR/$UNIT"
if [[ -f "$UPSTREAM_FILE" ]]; then
  cp "$UPSTREAM_FILE" "$BACKUP_DIR/$UPSTREAM_UNIT.previous"
fi
if [[ -f "$DROPIN_FILE" ]]; then
  cp "$DROPIN_FILE" "$BACKUP_DIR/10-go-replacement.conf.previous"
fi

echo "[2/7] build go proxy binary"
mkdir -p "$(dirname "$PROXY_BIN")"
(
  cd "$ROOT_DIR"
  "$GO_BIN" build -o "$PROXY_BIN" ./cmd/openclaw-gateway-proxy
)

echo "[3/7] generate upstream node service at $UPSTREAM_FILE"
cp "$FRAGMENT_PATH" "$UPSTREAM_FILE"
sed -E -i "s/^Description=.*/Description=OpenClaw Gateway Upstream (node passthrough)/" "$UPSTREAM_FILE"
if grep -Eq '^ExecStart=.*--port(=|[[:space:]]+)[0-9]+' "$UPSTREAM_FILE"; then
  sed -E -i "s/(^ExecStart=.*--port(=|[[:space:]]+))[0-9]+/\1$UPSTREAM_PORT/" "$UPSTREAM_FILE"
else
  sed -E -i "s|^ExecStart=(.*)$|ExecStart=\1 --port $UPSTREAM_PORT|" "$UPSTREAM_FILE"
fi
if grep -q '^Environment=OPENCLAW_GATEWAY_PORT=' "$UPSTREAM_FILE"; then
  sed -E -i "s/^Environment=OPENCLAW_GATEWAY_PORT=.*/Environment=OPENCLAW_GATEWAY_PORT=$UPSTREAM_PORT/" "$UPSTREAM_FILE"
fi
if grep -q '^Environment=OPENCLAW_SYSTEMD_UNIT=' "$UPSTREAM_FILE"; then
  sed -E -i "s/^Environment=OPENCLAW_SYSTEMD_UNIT=.*/Environment=OPENCLAW_SYSTEMD_UNIT=$UPSTREAM_UNIT/" "$UPSTREAM_FILE"
fi
if grep -q '^Environment=OPENCLAW_SERVICE_KIND=' "$UPSTREAM_FILE"; then
  sed -E -i 's/^Environment=OPENCLAW_SERVICE_KIND=.*/Environment=OPENCLAW_SERVICE_KIND=gateway-upstream/' "$UPSTREAM_FILE"
fi

echo "[4/7] write local override for $UNIT"
mkdir -p "$DROPIN_DIR" "$(dirname "$PROXY_LOG_FILE")"
cat >"$DROPIN_FILE" <<EOF
[Unit]
After=$UPSTREAM_UNIT
Requires=$UPSTREAM_UNIT

[Service]
ExecStart=
ExecStart=$PROXY_BIN
Environment=GATEWAY_HOST=$LISTEN_HOST
Environment=GATEWAY_PORT=$LISTEN_PORT
Environment=OPENCLAW_UPSTREAM_URL=http://127.0.0.1:$UPSTREAM_PORT
Environment=OPENCLAW_PROXY_LOG_FILE=$PROXY_LOG_FILE
Environment=OPENCLAW_PROXY_LOG_HEADERS=$PROXY_LOG_HEADERS
EOF

echo "[5/7] reload and start services"
systemctl --user daemon-reload
systemctl --user enable --now "$UPSTREAM_UNIT"
systemctl --user restart "$UNIT"

echo "[6/7] verify service health"
if ! systemctl --user is-active --quiet "$UNIT"; then
  echo "failed: $UNIT is not active" >&2
  exit 1
fi
if ! systemctl --user is-active --quiet "$UPSTREAM_UNIT"; then
  echo "failed: $UPSTREAM_UNIT is not active" >&2
  exit 1
fi
for _ in {1..15}; do
  if curl -fsS "http://127.0.0.1:$UPSTREAM_PORT/health" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done
if ! curl -fsS "http://$LISTEN_HOST:$LISTEN_PORT/health" >/dev/null 2>&1; then
  echo "warning: health endpoint check failed at http://$LISTEN_HOST:$LISTEN_PORT/health" >&2
fi

echo "[7/7] done"
echo "Gateway now runs through Go proxy on $LISTEN_HOST:$LISTEN_PORT with upstream node on 127.0.0.1:$UPSTREAM_PORT."
echo "Follow logs:"
echo "  journalctl --user -f -u $UNIT -u $UPSTREAM_UNIT -o short-iso"
echo "  tail -f $PROXY_LOG_FILE"
