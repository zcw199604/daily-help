# 任务清单: Token 并发刷新治理与 StateStore 过期清理

目录: `helloagents/plan/202601121409_token_state_hardening/`

---

## 1. Token 刷新并发治理
- [√] 1.1 在 `internal/wecom/client.go` 的 `getAccessToken` 中引入 singleflight，避免并发刷新击穿
- [√] 1.2 在 `internal/qinglong/client.go` 的 `getToken` 中引入 singleflight，避免并发刷新击穿
- [√] 1.3 补充并发单测：`internal/wecom/client_test.go` 与 `internal/qinglong/client_test.go` 验证 token 请求仅一次

## 2. StateStore 主动清理
- [√] 2.1 在 `internal/core/state.go` 为 `StateStore` 增加后台 janitor（ticker）定期清理过期 key
- [√] 2.2 补充单测：`internal/core/state_test.go` 验证不调用 `Get()` 也能清理过期条目

## 3. 安全检查
- [√] 3.1 安全检查：避免 token 泄露、避免竞态、避免 goroutine 泄漏

## 4. 文档与变更记录
- [√] 4.1 更新 `helloagents/wiki/modules/wecom.md`、`helloagents/wiki/modules/qinglong.md`、`helloagents/wiki/modules/core.md`
- [√] 4.2 更新 `helloagents/CHANGELOG.md`

## 5. 测试
- [√] 5.1 运行 `go test ./...`
