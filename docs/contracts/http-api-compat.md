# HTTP API 兼容说明书

日期: 2026-03-17

## 1. `POST /v1/chat/completions`

### 1.1 兼容要求

- 请求体字段、默认值、错误码语义对齐 OpenAI 兼容层
- `stream=false` 返回 JSON
- `stream=true` 返回 SSE

### 1.2 错误响应结构

```json
{
  "error": {
    "message": "...",
    "type": "invalid_request_error",
    "param": "model",
    "code": "invalid_model"
  }
}
```

### 1.3 SSE 要求

- `Content-Type: text/event-stream`
- 事件以 `data: ...\n\n` 输出
- 正常结束输出 `[DONE]`

## 2. `POST /tools/invoke`

请求:

```json
{
  "tool": "web.search",
  "arguments": {"q": "..."},
  "context": {
    "sessionId": "...",
    "scope": "dm:123"
  },
  "idempotencyKey": "..."
}
```

响应:

```json
{
  "ok": true,
  "result": {},
  "meta": {
    "durationMs": 12
  }
}
```

## 3. `GET /health`

响应:

```json
{
  "status": "ok",
  "version": "go-gateway/1.0.0",
  "uptimeSec": 12345,
  "deps": {
    "store": "ok",
    "pluginRuntime": "ok"
  }
}
```

## 4. 兼容验收

- 现有 OpenAI 客户端无需改动可调用
- 现有工具调用路径无需改动可调用
- 健康检查可被现有监控探针复用
