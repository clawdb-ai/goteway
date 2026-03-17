# goteway

OpenClaw Gateway 的高性能 Go 替代实现工作区（兼容优先，性能增强）。

## 当前内容

- 完整替代蓝图: `docs/blueprint/openclaw-go-gateway-replacement-blueprint.md`
- 协议兼容契约: `docs/contracts/ws-protocol-compat.md`
- HTTP 兼容契约: `docs/contracts/http-api-compat.md`
- 插件兼容契约: `docs/contracts/plugin-compat-contract.md`
- 压测规范: `docs/testing/perf-benchmark-spec.md`
- 兼容测试矩阵: `docs/testing/compatibility-test-matrix.md`
- 运维手册: `docs/ops/runbook.md`
- SLO/告警: `docs/ops/slo-and-alerts.md`
- 配置样例: `docs/config/openclaw.config.compat.sample.yaml`

## Go 工程骨架

- 启动入口: `cmd/goteway/main.go`
- 运行时装配: `internal/runtime/app.go`
- 协议协商: `internal/protocol/compat.go`
- 认证骨架: `internal/auth/service.go`
- 会话管理骨架: `internal/session/manager.go`
- 插件注册与校验骨架: `internal/plugin/registry.go`, `internal/plugin/validator.go`
- HTTP 兼容端点骨架: `internal/transport/httpapi/server.go`
- WS 抽象契约: `internal/transport/ws/contract.go`
- 幂等存储骨架: `internal/idempotency/store.go`

## 验证

```bash
go test ./...
```

## 状态说明

本仓库已完成“蓝图与契约层”全项落地，并给出可编译的执行骨架。  
完整替代官方网关仍需继续实现 WS 真实传输、插件运行时桥接、全插件回归、性能调优与灰度发布流程。
