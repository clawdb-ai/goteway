# WS 协议兼容说明书

日期: 2026-03-17

## 1. 帧结构

### 1.1 请求帧 `req`

```json
{
  "type": "req",
  "id": "string",
  "method": "string",
  "params": {}
}
```

约束:
- `id` 在连接内唯一
- `method` 必须在服务端方法注册表中
- `params` 可以为空对象，不可为数组

### 1.2 响应帧 `res`

```json
{
  "type": "res",
  "id": "string",
  "ok": true,
  "payload": {}
}
```

错误响应:

```json
{
  "type": "res",
  "id": "string",
  "ok": false,
  "payload": {
    "error": {
      "code": "ERR_*",
      "message": "human readable",
      "retryable": false
    }
  }
}
```

约束:
- `id` 必须与对应 `req.id` 一致
- 每个 `req` 必有且仅有一个终态 `res`

### 1.3 事件帧 `event`

```json
{
  "type": "event",
  "event": "tick|presence|agent",
  "payload": {},
  "seq": 123,
  "stateVersion": 7
}
```

约束:
- `seq` 在连接上下文严格递增
- `stateVersion` 表示状态快照版本，仅在状态变更时递增

## 2. 握手协议 `connect`

请求:

```json
{
  "type": "req",
  "id": "c1",
  "method": "connect",
  "params": {
    "minProtocol": 1,
    "maxProtocol": 3,
    "client": {
      "name": "client-name",
      "version": "1.2.3"
    },
    "auth": {
      "type": "token",
      "token": "***"
    }
  }
}
```

成功响应:

```json
{
  "type": "res",
  "id": "c1",
  "ok": true,
  "payload": {
    "protocol": 3,
    "clientId": "cli_123"
  }
}
```

失败场景:
- 无可用协议版本
- 认证失败
- 客户端参数非法

## 3. 幂等规则

副作用方法（如发送消息、工具执行）可携带:

```json
{ "idempotencyKey": "uuid-v4" }
```

服务端行为:
- 在 TTL 内重复请求直接返回已完成结果（或处理中状态）
- 相同 key 且参数哈希不同则返回冲突错误

## 4. 顺序与重放

- 单连接事件顺序由 `seq` 保证
- 服务端保留最近 N 条事件用于断线短窗口重放
- 客户端可通过 `lastSeq` 请求补偿

## 5. 兼容验收

- 官方客户端握手通过率 100%
- 官方事件消费者无顺序异常
- 所有 `res.id` 与 `req.id` 对应率 100%
