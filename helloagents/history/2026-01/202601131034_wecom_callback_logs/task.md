# 任务清单: 企业微信回调链路日志增强

目录: `helloagents/plan/202601131034_wecom_callback_logs/`

---

## 1. 回调可观测性
- [√] 1.1 在 `internal/wecom/callback.go` 增加 GET/POST 回调的验签/解密/解析阶段结构化日志，便于定位不回复问题
- [√] 1.2 在 `internal/app/server.go` 的请求日志中补充 `status_code/response_bytes`

## 2. 知识库同步
- [√] 2.1 更新 `helloagents/CHANGELOG.md`（Unreleased）记录日志增强

## 3. 测试
- [√] 3.1 运行 `go test ./...`（Docker: `golang:1.22`）
