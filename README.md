# goteway

[![OpenClaw Skill](https://img.shields.io/badge/OpenClaw-Skill-2f855a)](https://openclaw.ai)
[![Gateway](https://img.shields.io/badge/OpenClaw-Go%20Gateway-111827)](https://github.com/mac/goteway)
[![Local Swap](https://img.shields.io/badge/Mode-Machine%20Local-0ea5e9)](https://github.com/mac/goteway)

> **Where Go meets Gateway, and `gote` sounds like `goat`.**

[中文](#goteway-中文) | [English](#goteway-english)

---

## goteway (English)

### Philosophy

`goteway` is a compatibility-first Go replacement track for OpenClaw Gateway.
The goal is simple: **keep protocol behavior stable**, improve operational control, and make debugging easier.

### What goteway offers

| Feature | Description |
|---------|-------------|
| **Protocol compatibility contracts** | WS/HTTP/plugin compatibility specs for replacement-safe behavior |
| **Go gateway runtime scaffold** | Structured runtime layers for auth, sessions, protocol, and HTTP compatibility |
| **Machine-local Go swap path** | Safe `systemd --user` replacement flow with backup and rollback |
| **Debug-focused log tracking** | Live journald + proxy file logs to trace receive/dispatch/reply behavior |
| **Rollback-first operations** | One-command restore to the original Node gateway service |

### The Name

**gote** + **gateway** = **goteway**  
And yes, **`gote` sounds like `goat`**: stubbornly reliable on rough terrain.

### Quick Start

```bash
# 1) Install as an OpenClaw skill from GitHub
openclaw skills install https://github.com/mac/goteway.git

# 2) Ask the skill to install the local Go gateway replacement
openclaw skills run goteway-goat-gateway --task "replace my local openclaw gateway with goteway and keep debug logs"

# 3) Verify status
./scripts/openclaw/status-openclaw-gateway.sh

# 4) Follow logs
./scripts/openclaw/follow-openclaw-logs.sh
```

### Rollback

```bash
./scripts/openclaw/rollback-local-openclaw-gateway.sh
```

### Repository Map

- Blueprint: `docs/blueprint/openclaw-go-gateway-replacement-blueprint.md`
- Contracts: `docs/contracts/ws-protocol-compat.md`, `docs/contracts/http-api-compat.md`, `docs/contracts/plugin-compat-contract.md`
- Ops: `docs/ops/runbook.md`, `docs/ops/slo-and-alerts.md`
- Config sample: `docs/config/openclaw.config.compat.sample.yaml`
- Skill entry: `SKILL.md`

---

## goteway 中文

### 设计理念

`goteway` 是 OpenClaw Gateway 的 Go 替代实现路线，核心原则是：
**兼容优先，行为不漂移，排障更直接**。

### goteway 提供什么

| 能力 | 描述 |
|------|------|
| **协议兼容契约** | 提供 WS/HTTP/插件兼容规范，确保可替换性 |
| **Go 运行时骨架** | 认证、会话、协议协商、HTTP 兼容层分层清晰 |
| **本机替换流程** | `systemd --user` 下可回滚的机器本地替换 |
| **日志追踪能力** | journald + 代理日志，便于定位接收/分发/回复链路 |
| **回滚优先运维** | 单命令恢复原 Node 网关服务 |

### 名字的由来

**gote + gateway = goteway**。  
`gote` 的发音接近 `goat`，意思也很直白：在复杂地形里也要稳。

### 快速开始

```bash
# 1) 从 GitHub 安装技能
openclaw skills install https://github.com/mac/goteway.git

# 2) 让技能执行本机 Go 网关替换
openclaw skills run goteway-goat-gateway --task "替换当前 openclaw gateway 为 goteway，并持续记录日志"

# 3) 检查状态
./scripts/openclaw/status-openclaw-gateway.sh

# 4) 跟踪日志
./scripts/openclaw/follow-openclaw-logs.sh
```

### 回滚

```bash
./scripts/openclaw/rollback-local-openclaw-gateway.sh
```

### 文档索引

- 蓝图：`docs/blueprint/openclaw-go-gateway-replacement-blueprint.md`
- 契约：`docs/contracts/ws-protocol-compat.md`、`docs/contracts/http-api-compat.md`、`docs/contracts/plugin-compat-contract.md`
- 运维：`docs/ops/runbook.md`、`docs/ops/slo-and-alerts.md`
- 配置样例：`docs/config/openclaw.config.compat.sample.yaml`
- 技能入口：`SKILL.md`

---

## Verify

```bash
go test ./...
```

*Build the goat path. Keep the gateway calm.*
