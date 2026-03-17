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
EXEC_COLOR_SUPPORTED="false"
EXEC_SANDBOX_SUPPORTED="false"
EXEC_ASK_ARG_SUPPORTED="false"
EXEC_SKIP_GIT_REPO_CHECK_SUPPORTED="false"
RESUME_COLOR_SUPPORTED="false"
RESUME_SANDBOX_SUPPORTED="false"
RESUME_ASK_ARG_SUPPORTED="false"
RESUME_SKIP_GIT_REPO_CHECK_SUPPORTED="false"
if command -v "$CODEX_BIN" >/dev/null 2>&1; then
  EXEC_HELP="$("$CODEX_BIN" exec --help 2>/dev/null || true)"
  RESUME_HELP="$("$CODEX_BIN" exec resume --help 2>/dev/null || true)"
  if printf '%s\n' "$EXEC_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--color([[:space:]]|$)'; then EXEC_COLOR_SUPPORTED="true"; fi
  if printf '%s\n' "$EXEC_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--sandbox([[:space:]]|$)'; then EXEC_SANDBOX_SUPPORTED="true"; fi
  if printf '%s\n' "$EXEC_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--ask-for-approval([[:space:]]|$)'; then EXEC_ASK_ARG_SUPPORTED="true"; fi
  if printf '%s\n' "$EXEC_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--skip-git-repo-check([[:space:]]|$)'; then EXEC_SKIP_GIT_REPO_CHECK_SUPPORTED="true"; fi
  if printf '%s\n' "$RESUME_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--color([[:space:]]|$)'; then RESUME_COLOR_SUPPORTED="true"; fi
  if printf '%s\n' "$RESUME_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--sandbox([[:space:]]|$)'; then RESUME_SANDBOX_SUPPORTED="true"; fi
  if printf '%s\n' "$RESUME_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--ask-for-approval([[:space:]]|$)'; then RESUME_ASK_ARG_SUPPORTED="true"; fi
  if printf '%s\n' "$RESUME_HELP" | grep -Eq -- '^[[:space:]]+(-[[:alnum:]],[[:space:]]+)?--skip-git-repo-check([[:space:]]|$)'; then RESUME_SKIP_GIT_REPO_CHECK_SUPPORTED="true"; fi
fi

node - "$CONFIG_PATH" "$TMP_FILE" \
  "$EXEC_COLOR_SUPPORTED" \
  "$EXEC_SANDBOX_SUPPORTED" \
  "$EXEC_ASK_ARG_SUPPORTED" \
  "$EXEC_SKIP_GIT_REPO_CHECK_SUPPORTED" \
  "$RESUME_COLOR_SUPPORTED" \
  "$RESUME_SANDBOX_SUPPORTED" \
  "$RESUME_ASK_ARG_SUPPORTED" \
  "$RESUME_SKIP_GIT_REPO_CHECK_SUPPORTED" <<'NODE'
const fs = require("fs");

const [
  ,
  ,
  configPath,
  outPath,
  execColorSupportedRaw,
  execSandboxSupportedRaw,
  execAskArgSupportedRaw,
  execSkipGitRepoCheckSupportedRaw,
  resumeColorSupportedRaw,
  resumeSandboxSupportedRaw,
  resumeAskArgSupportedRaw,
  resumeSkipGitRepoCheckSupportedRaw,
] = process.argv;
const raw = fs.readFileSync(configPath, "utf8");
const cfg = JSON.parse(raw);
const execColorSupported = execColorSupportedRaw === "true";
const execSandboxSupported = execSandboxSupportedRaw === "true";
const execAskArgSupported = execAskArgSupportedRaw === "true";
const execSkipGitRepoCheckSupported = execSkipGitRepoCheckSupportedRaw === "true";
const resumeColorSupported = resumeColorSupportedRaw === "true";
const resumeSandboxSupported = resumeSandboxSupportedRaw === "true";
const resumeAskArgSupported = resumeAskArgSupportedRaw === "true";
const resumeSkipGitRepoCheckSupported = resumeSkipGitRepoCheckSupportedRaw === "true";

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

const execArgs = ["exec", "--json"];
if (execColorSupported) execArgs.push("--color", "never");
if (execSandboxSupported) execArgs.push("--sandbox", "danger-full-access");
if (execAskArgSupported) execArgs.push("--ask-for-approval", "never");
if (execSkipGitRepoCheckSupported) execArgs.push("--skip-git-repo-check");
backend.args = execArgs;

const resumeArgs = ["exec", "resume", "{sessionId}"];
resumeArgs.push("--dangerously-bypass-approvals-and-sandbox");
if (resumeSkipGitRepoCheckSupported) resumeArgs.push("--skip-git-repo-check");
backend.resumeArgs = resumeArgs;

cfg.agents.defaults.cliBackends["codex-cli"] = backend;

fs.writeFileSync(outPath, `${JSON.stringify(cfg, null, 2)}\n`, "utf8");
NODE

mv "$TMP_FILE" "$CONFIG_PATH"

echo "Patched codex-cli backend args in $CONFIG_PATH"
echo "Detected flags: exec(color=$EXEC_COLOR_SUPPORTED sandbox=$EXEC_SANDBOX_SUPPORTED ask=$EXEC_ASK_ARG_SUPPORTED skip=$EXEC_SKIP_GIT_REPO_CHECK_SUPPORTED) resume(color=$RESUME_COLOR_SUPPORTED sandbox=$RESUME_SANDBOX_SUPPORTED ask=$RESUME_ASK_ARG_SUPPORTED skip=$RESUME_SKIP_GIT_REPO_CHECK_SUPPORTED bypass=forced)"
echo "Backup: $BACKUP_DIR/openclaw.json.$TIMESTAMP.bak"
