#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH="${OPENCLAW_CONFIG_PATH:-$HOME/.openclaw/openclaw.json}"
BACKUP_DIR="${OPENCLAW_CONFIG_BACKUP_DIR:-$HOME/.local/state/openclaw-go-replacement/config-backups}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "openclaw config not found: $CONFIG_PATH" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"
cp "$CONFIG_PATH" "$BACKUP_DIR/openclaw.json.$TIMESTAMP.bak"

TMP_FILE="$(mktemp)"
trap 'rm -f "$TMP_FILE"' EXIT

CODEX_BIN="${CODEX_BIN:-codex}"
ASK_ARG_SUPPORTED="false"
if command -v "$CODEX_BIN" >/dev/null 2>&1; then
  if "$CODEX_BIN" exec --help 2>/dev/null | grep -q -- "--ask-for-approval"; then
    ASK_ARG_SUPPORTED="true"
  fi
fi

node - "$CONFIG_PATH" "$TMP_FILE" "$ASK_ARG_SUPPORTED" <<'NODE'
const fs = require("fs");

const [, , configPath, outPath, askArgSupportedRaw] = process.argv;
const raw = fs.readFileSync(configPath, "utf8");
const cfg = JSON.parse(raw);
const askArgSupported = askArgSupportedRaw === "true";

cfg.agents = cfg.agents && typeof cfg.agents === "object" ? cfg.agents : {};
cfg.agents.defaults =
  cfg.agents.defaults && typeof cfg.agents.defaults === "object"
    ? cfg.agents.defaults
    : {};
cfg.agents.defaults.cliBackends =
  cfg.agents.defaults.cliBackends &&
  typeof cfg.agents.defaults.cliBackends === "object"
    ? cfg.agents.defaults.cliBackends
    : {};

const backend = cfg.agents.defaults.cliBackends["codex-cli"] || {};
backend.command =
  typeof backend.command === "string" && backend.command.trim()
    ? backend.command
    : "codex";
backend.env = backend.env && typeof backend.env === "object" ? backend.env : {};

backend.args = [
  "exec",
  "--json",
  "--color",
  "never",
  "--sandbox",
  "danger-full-access",
  "--skip-git-repo-check",
];
if (askArgSupported) {
  backend.args.splice(6, 0, "--ask-for-approval", "never");
}

backend.resumeArgs = [
  "exec",
  "resume",
  "{sessionId}",
  "--color",
  "never",
  "--sandbox",
  "danger-full-access",
  "--skip-git-repo-check",
];
if (askArgSupported) {
  backend.resumeArgs.splice(7, 0, "--ask-for-approval", "never");
}

cfg.agents.defaults.cliBackends["codex-cli"] = backend;

fs.writeFileSync(outPath, `${JSON.stringify(cfg, null, 2)}\n`, "utf8");
NODE

mv "$TMP_FILE" "$CONFIG_PATH"

if [[ "$ASK_ARG_SUPPORTED" == "true" ]]; then
  echo "Patched codex-cli backend sandbox mode in $CONFIG_PATH (with --ask-for-approval)"
else
  echo "Patched codex-cli backend sandbox mode in $CONFIG_PATH (without --ask-for-approval)"
fi
echo "Backup: $BACKUP_DIR/openclaw.json.$TIMESTAMP.bak"
