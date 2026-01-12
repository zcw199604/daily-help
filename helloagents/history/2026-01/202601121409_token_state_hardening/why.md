# 变更提案: Token 并发刷新治理与 StateStore 过期清理

## 需求背景
当前系统会同时对接企业微信与青龙 OpenAPI，并在运行过程中缓存 access token。

在高并发（或瞬时并发）场景下，当 token 过期时会出现“缓存击穿”问题：多个协程同时判定 token 无效并并发向上游拉取新 token，造成不必要的资源浪费并可能触发上游 API 限流。

同时，core 的 `StateStore` 目前仅在 `Get()` 时做懒惰删除；当用户状态写入后长期不再交互时，过期状态可能长期驻留在内存 map 中，存在内存增长风险。

## 变更内容
1. wecom 与 qinglong 的 token 获取改为“单飞”刷新：同一时刻仅允许一个协程刷新 token，其余协程等待结果并复用缓存。
2. StateStore 增加后台定时清理：周期性扫描并删除过期会话状态，避免过期数据长期占用内存。
3. 补充并发场景单元测试，确保行为在并发下稳定可验证。

## 影响范围
- **模块:** wecom / qinglong / core
- **文件:** `internal/wecom/client.go`、`internal/qinglong/client.go`、`internal/core/state.go` 及对应测试
- **依赖:** 新增 `golang.org/x/sync/singleflight`
- **对外行为:** 无新增功能；仅提升稳定性与资源效率

## 核心场景

### 需求: Token 并发刷新抑制
**模块:** wecom / qinglong
当 token 过期或未初始化时，多协程并发触发 API 调用，仅产生一次上游 token 请求。

### 需求: 过期状态主动清理
**模块:** core
用户状态写入后，即使不再调用 `Get()`，过期状态也会被后台清理回收。

## 风险评估
- **风险:** 引入后台 goroutine（StateStore janitor）增加实现复杂度
  - **缓解:** 提供 `Close()` 停止清理协程（便于测试与未来优雅退出）；清理逻辑受互斥锁保护
- **风险:** 引入新依赖 `x/sync/singleflight`
  - **缓解:** 依赖稳定、体积小；单元测试覆盖并发场景
