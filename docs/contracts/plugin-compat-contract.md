# 插件兼容契约（全类型）

日期: 2026-03-17

## 1. Manifest 契约

必须存在 `openclaw.plugin.json`，至少包含:

```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "type": "provider|channel|tool|memory|speech|hook|command|cron|http-route|interactive",
  "entry": "index.js",
  "capabilities": []
}
```

校验要求:
- 零代码验证模式下只做静态校验，不执行插件
- `entry` 必须在插件目录内（路径穿越检查）

## 2. 运行时兼容面

Go Gateway 通过 Runtime Bridge 暴露以下能力:
- 会话读写
- 事件发布
- 工具执行
- 配置读取
- 审计上报

## 3. Provider 契约

- 输入: model/messages/tools/stream/options
- 输出: completion chunks 或 final completion
- 错误: 标准错误码、是否可重试

## 4. Channel 契约

- 必须实现入站消息转换与出站发送
- 必须支持身份绑定和断线重连
- 必须保证消息去重键可配置

## 5. Tool 契约

- 声明 schema
- 执行前权限检查
- 超时、重试、熔断可配置

## 6. 隔离与资源控制

- 每插件并发上限
- 每调用超时时间
- 内存/CPU 预算（按运行时能力）

## 7. 热重载

- manifest 或配置变更触发插件重载
- 重载失败不影响已运行插件

## 8. 兼容验收

- 插件仓库全量扫描通过
- 插件加载失败率低于阈值
- 插件调用错误码与官方一致
