# 任务清单: 企业微信发送消息日志增强

目录: `helloagents/plan/202601130955_wecom_send_logging/`

---

## 1. 日志与可观测性
- [√] 1.1 在 `internal/wecom/client.go` 增加 `message/send`、`gettoken`、`update_template_card` 的结构化日志（不输出敏感 token/secret）
- [√] 1.2 `message/send` 在 errcode=0 但返回 `invaliduser/invalidparty/invalidtag/unlicenseduser` 时输出告警并返回错误
- [√] 1.3 在 `internal/core/router.go` 增加发送无权限提示/更新卡片按钮失败的错误日志

## 2. 知识库同步
- [√] 2.1 更新 `helloagents/CHANGELOG.md`（Unreleased）记录日志增强与不可达告警行为

## 3. 测试
- [√] 3.1 运行 `go test ./...`（Docker: `golang:1.22`）
