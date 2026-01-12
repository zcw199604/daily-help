# 技术设计: Token 并发刷新治理与 StateStore 过期清理

## 技术方案

### Token 刷新并发治理
- 在 `internal/wecom/client.go` 与 `internal/qinglong/client.go` 中引入 `golang.org/x/sync/singleflight`。
- 获取 token 的流程改为：
  1. 先加锁检查缓存是否有效；有效则直接返回。
  2. 无效则进入 `singleflight.Do`（同 key 合并为单次刷新）。
  3. 在 `Do` 的函数内再次“双重检查”缓存；避免并发等待结束后重复刷新。
  4. 发起上游 token 请求并写回缓存；所有等待协程复用该结果。

### StateStore 主动过期清理
- 在 `internal/core/state.go` 的 `StateStore` 内部启动后台 janitor goroutine：
  - 定时 tick（间隔 `min(ttl, 1min)`）
  - 扫描并删除 `ExpiresAt` 已过期的 key
- 提供 `Close()` 用于停止 janitor（测试与未来优雅退出可复用）。

## 安全与性能
- 不记录 token 明文到日志；错误信息避免包含敏感字段。
- `singleflight` 抑制击穿可显著降低上游 token 请求数量，避免限流风险。

## 测试
- wecom：并发调用 `SendText`，断言 `/gettoken` 仅命中一次。
- qinglong：并发调用带鉴权的 API（如 `ListCrons`），断言 `/open/auth/token` 仅命中一次。
- core：写入过期状态后不调用 `Get()`，等待 janitor 运行后断言 map 中条目被清理。
