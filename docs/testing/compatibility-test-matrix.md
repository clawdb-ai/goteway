# 兼容性测试矩阵

日期: 2026-03-17

## 1. WebSocket 协议

- 握手协商: minProtocol/maxProtocol 边界
- req/res 关联: 丢包、重试、超时
- event 顺序: seq 单调、重放补偿

## 2. 认证授权

- Token 认证成功/失败
- Password 认证成功/失败
- Device Pairing 全链路
- Allowlist 与 Group Policy 命中

## 3. 消息与会话

- DM/Group/Thread 路由隔离
- 会话恢复与上下文续接
- Identity Links 行为一致

## 4. HTTP API

- /v1/chat/completions 非流式
- /v1/chat/completions 流式 SSE
- /tools/invoke 正常与异常路径
- /health 依赖异常降级

## 5. 插件系统

- manifest 校验
- 发现与加载
- 热重载成功/失败回滚
- Provider/Channel/Tool 关键用例

## 6. 稳定性

- 高并发长稳 24h
- 慢消费者与背压
- 插件超时、崩溃隔离

## 7. 发布门禁

- 协议兼容失败用例 = 0
- P0/P1 缺陷 = 0
- 关键性能指标达标
