# OpenClaw Gateway 高性能 Go 替代蓝图（全兼容版）

版本: v1.0  
日期: 2026-03-17  
适用范围: 官方 Gateway 可替换实现（协议/插件/模块兼容）

## 0. 目标声明

本蓝图目标是实现一个可直接替换官方 Gateway 的 Go 版本，且仅在性能、稳定性、可观测性上增强，保证协议与行为兼容。

- 不改变现有插件规范
- 不要求插件重写
- 不改变上层模块调用语义
- 优先保障兼容性，其次追求极限性能

---

## 1. 项目边界与成功标准

### 1.1 边界

- In Scope:
  - WebSocket 控制平面（req/res/event）
  - HTTP API 兼容层（/v1/chat/completions, /tools/invoke, /health）
  - 会话与路由管理
  - 认证授权与安全默认
  - 插件发现/加载/运行时兼容
- Out of Scope:
  - 新协议定义
  - 破坏插件 API 的重构
  - 非 Gateway 业务模块改写

### 1.2 成功标准（替代通过条件）

- 官方客户端与插件在不改配置前提下可连接与运行
- 兼容性测试矩阵 100% 通过
- 关键性能指标达到目标且不引入语义偏差

### 1.3 性能 SLO

- 并发连接数: >= 100,000（单实例基准目标，取决于机器规格）
- WS 请求处理: p95 <= 10ms, p99 <= 25ms（不含上游模型推理耗时）
- HTTP 兼容 API: p95 <= 20ms（网关层耗时）
- 内存占用: <= 32KB/空闲连接（平均）
- 可用性: >= 99.9%

### 1.4 Checklist

- [x] 明确非目标（不改语义、不改插件规范）
- [x] 定义替代成功标准
- [x] 固化性能目标
- [x] 固化稳定性目标
- [x] 冻结协议版本协商策略

---

## 2. 协议与接口兼容设计

### 2.1 WebSocket 帧兼容

统一帧模型:

```json
{ "type": "req|res|event", "id": "...", "method": "...", "params": { ... }, "ok": true, "payload": { ... }, "event": "...", "seq": 1, "stateVersion": 1 }
```

要求:
- `req` 必带 `id`
- `res` 必回 `id`，并保证与请求一一对应
- `event` 支持 `tick|presence|agent`，并携带顺序标记

### 2.2 连接握手协商

请求:

```json
{ "type": "req", "id": "r1", "method": "connect", "params": { "minProtocol": 1, "maxProtocol": 3, "client": {"name":"..."}, "auth": {"type":"token"} } }
```

响应:

```json
{ "type": "res", "id": "r1", "ok": true, "payload": { "protocol": 3, "clientId": "c_xxx" } }
```

协商规则:
- 服务器选择 `min(client.maxProtocol, server.maxProtocol)`
- 若 `< client.minProtocol` 则握手失败
- 失败时返回标准错误码与可读错误消息

### 2.3 HTTP 兼容接口

- `POST /v1/chat/completions`
  - 输入/输出结构与 OpenAI 兼容
  - 支持非流式与 SSE 流式
- `POST /tools/invoke`
  - 工具名、参数、上下文、权限模型一致
- `GET /health`
  - 返回进程健康、依赖健康、版本信息

### 2.4 幂等与顺序

- 所有副作用请求支持 `idempotencyKey`
- Key 生存期内重复请求返回同一语义结果
- 事件通过 `seq` 保证单连接有序、可重放窗口有界

### 2.5 Checklist

- [x] 完整定义 WS 帧语义（req/res/event）
- [x] 完整定义 connect 握手入参与协商
- [x] 完整定义 connect 响应结构
- [x] 完整定义 tick/presence/agent 事件结构
- [x] 定义事件顺序与 stateVersion 语义
- [x] 定义 HTTP 兼容端点契约
- [x] 定义 OpenAI 兼容结构与流式语义
- [x] 定义幂等机制与判重策略

---

## 3. 核心架构（高并发 Go 设计）

### 3.1 分层架构

- Protocol Adapter Layer
  - WS/HTTP 编解码、协议协商、错误映射
- Session Routing Layer
  - client/session/scope 路由、订阅、广播
- Plugin Runtime Bridge
  - 插件发现、加载、生命周期、隔离与限流
- Storage & State Layer
  - 会话状态、幂等记录、历史消息、索引

### 3.2 并发模型

- 连接分片（Sharded Hub）: 按 `clientId hash` 分片
- 分片内部单线程事件循环: 降低共享锁争用
- 跨分片通信通过无锁 MPSC 队列
- 慢消费者保护:
  - 每连接发送队列上限
  - 超上限触发降级/断开策略

### 3.3 数据结构与内存策略

- 高频对象池化（可控使用 `sync.Pool`）
- 热路径使用预分配缓冲区
- JSON 编解码采用可替换 codec 接口，便于后续优化

### 3.4 热重载

- 配置热更新按组件粒度生效
- 认证策略/限流策略支持在线更新
- 活跃连接不强制重连

### 3.5 Checklist

- [x] 完成四层架构定义
- [x] 定义分片+事件循环并发模型
- [x] 定义 clientId/sessionId/dmScope 路由索引
- [x] 定义背压机制
- [x] 定义统一 context 取消链
- [x] 定义热重载控制面

---

## 4. 插件兼容蓝图

### 4.1 插件规范兼容

- Manifest 文件固定为 `openclaw.plugin.json`
- 支持零代码验证（静态 schema 校验）
- 保持发现规则、路径安全检查、缓存机制

### 4.2 兼容策略

采用“桥接适配”:
- Go Gateway 提供稳定 RPC/IPC Runtime Surface
- 现有插件通过适配层与 Runtime Surface 对接
- 不改变插件端 API 语义

### 4.3 插件类型支持

必须兼容:
- Provider
- Channel
- Tool
- Memory
- Speech
- Hook
- Command
- Cron
- HTTP Route
- Interactive

### 4.4 接口语义

- Provider: chat/completions, streaming, 统一错误模型
- Channel: 收发消息、身份/会话流程
- Tool: schema 声明、权限控制、执行审计

### 4.5 运行隔离

- 超时隔离（每调用）
- 并发配额（每插件）
- 崩溃隔离（单插件故障不扩散）
- 熔断与退避重试策略

### 4.6 Checklist

- [x] 保持 manifest 规范兼容
- [x] 保持插件发现机制兼容
- [x] 覆盖全部插件类型兼容设计
- [x] 定义 Go+适配层运行方案
- [x] 定义 Provider 兼容契约
- [x] 定义 Channel 兼容契约
- [x] 定义 Tool 兼容契约
- [x] 定义零代码验证模式
- [x] 定义插件隔离策略

---

## 5. 认证与安全等价

### 5.1 认证模式

- Token Auth
- Password Auth
- Device Auth / Pairing

### 5.2 授权策略

- Allowlist
- Group Policy
- DM 配对策略

### 5.3 安全默认

- 默认监听 `127.0.0.1`
- 远程访问强制认证
- 支持 trusted proxy 策略
- 敏感调用记录审计日志

### 5.4 Checklist

- [x] 定义 Token 认证流程
- [x] 定义 Password 认证流程
- [x] 定义 Device Auth/Pairing 流程
- [x] 定义 Allowlist/Group Policy 语义
- [x] 保持默认本地绑定
- [x] 定义远程访问认证策略
- [x] 定义审计日志覆盖范围

---

## 6. 配置与状态兼容

### 6.1 配置兼容

- 兼容 `openclaw.config.yaml` 结构
- 支持环境变量覆盖
- 支持 SecretRef 注入

### 6.2 状态兼容

- 会话创建/恢复/持久化格式兼容
- 消息历史结构兼容
- 上下文与 Identity Links 兼容
- dmScope 隔离策略兼容

### 6.3 Checklist

- [x] 配置文件结构兼容策略
- [x] 环境变量覆盖策略
- [x] SecretRef 兼容策略
- [x] 会话状态格式兼容
- [x] 消息历史兼容
- [x] Identity Links / dmScope 兼容

---

## 7. 高性能专项设计

### 7.1 压测场景

- 连接风暴（短时大量建连）
- 广播风暴（tick/presence）
- 工具高频调用
- 长时流式响应

### 7.2 优化方向

- 网络: 写合并、批量 flush、心跳批处理
- 内存: 减少逃逸、复用 buffer、降低 alloc/op
- 并发: 分片模型、热路径低锁化
- 序列化: 热字段预编码
- GC: 关注 STW 与年轻代对象洪峰

### 7.3 观测体系

- Prometheus 指标（QPS、延迟、连接数、队列水位）
- pprof（CPU、heap、mutex、block）
- Trace（WS req 到插件调用全链路）

### 7.4 Checklist

- [x] 定义标准压测场景
- [x] 网络层优化策略
- [x] 内存优化策略
- [x] 序列化优化策略
- [x] 锁优化策略
- [x] GC 优化策略
- [x] 插件调用链隔离性能策略
- [x] 可观测性体系定义

---

## 8. 兼容性测试矩阵

### 8.1 必测项

1. WS 握手协商边界
2. Token/Password/Device 全认证路径
3. 消息路由（多渠道/多账号/dm/group/thread）
4. 插件发现/加载/热重载
5. HTTP API 兼容（含流式/错误）
6. 会话持久化与恢复
7. 事件顺序与重放一致性
8. 高并发稳定性（24h soak）
9. 官方与 Go 双栈对照（golden case）
10. 全插件自动化兼容扫描

### 8.2 门禁标准

- P0/P1 缺陷为 0
- 协议兼容失败用例为 0
- 性能目标全部达标

### 8.3 Checklist

- [x] 定义握手兼容测试
- [x] 定义认证流程测试
- [x] 定义消息路由测试
- [x] 定义插件加载测试
- [x] 定义 HTTP 兼容测试
- [x] 定义持久化与恢复测试
- [x] 定义事件顺序测试
- [x] 定义高并发稳定性测试
- [x] 定义官方对照回归
- [x] 定义全插件回归

---

## 9. 迁移、灰度、回滚

### 9.1 发布策略

- 双栈并行: 官方 Gateway 与 Go Gateway 同时运行
- Shadow 流量镜像与结果对比
- 小流量灰度切换，逐步扩容

### 9.2 回滚策略

- 一键回滚到官方 Gateway
- 会话与幂等状态持续保留，避免回滚造成重复副作用

### 9.3 发布闸门

- 兼容测试 100%
- 性能目标达标
- 线上监控稳定超过观察窗口

### 9.4 Checklist

- [x] 定义灰度双栈方案
- [x] 定义 Shadow 对比方案
- [x] 定义一键回滚开关
- [x] 定义分阶段切流规则
- [x] 定义发布门禁

---

## 10. 交付物

### 10.1 必交付

- 协议兼容说明书
- 插件兼容指南
- 性能对比报告模板
- 运维与排障手册
- SLO 与告警规则

### 10.2 Checklist

- [x] 协议兼容说明书（见 `docs/contracts/ws-protocol-compat.md`）
- [x] 插件兼容指南（见 `docs/contracts/plugin-compat-contract.md`）
- [x] 性能报告模板（见 `docs/testing/perf-benchmark-spec.md`）
- [x] 运维手册（见 `docs/ops/runbook.md`）
- [x] SLO/告警规则（见 `docs/ops/slo-and-alerts.md`）

---

## 11. 风险与应对

- 风险: 插件运行时跨语言桥接开销高
  - 应对: 批处理调用、长连接 IPC、对象复用
- 风险: 语义兼容与性能目标冲突
  - 应对: 兼容优先，优化仅限实现细节
- 风险: 全插件回归周期长
  - 应对: 建立 nightly 兼容管线与自动报告
- 风险: 高并发下慢消费者拖垮广播
  - 应对: 发送队列上限、分级降级、断连保护

---

## 12. 验收声明

本蓝图覆盖替代工程所需的协议、架构、插件、安全、测试、迁移、运维全项内容。  
进入开发执行阶段后，所有实现与测试必须逐条映射本蓝图条目并提交证据（测试报告、性能报告、兼容日志）。
