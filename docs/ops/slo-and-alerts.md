# SLO 与告警规则

日期: 2026-03-17

## 1. SLO

- 可用性 SLO: 99.9%
- 网关层延迟 SLO:
  - WS p95 <= 10ms
  - WS p99 <= 25ms
  - HTTP p95 <= 20ms
- 错误率 SLO: <= 0.1%

## 2. 告警规则

### P1
- ws_handshake_success_rate < 99.5% 持续 5 分钟
- req_error_rate > 1% 持续 5 分钟
- process_restart_total 异常增长

### P2
- ws_send_queue_depth p95 超阈值 10 分钟
- plugin_call_timeout_total 持续增长
- memory_rss 增速异常（泄漏风险）

### P3
- /health 依赖降级但核心服务可用
- 部分插件不可用

## 3. 值班处置

- P1: 5 分钟内响应，必要时触发回滚
- P2: 15 分钟内响应，限流或降级
- P3: 观察并排期修复
