# 任务清单: 启动成功通知（企业微信应用消息）

目录: `helloagents/plan/202601131512_startup_notify_wecom/`

---

## 1. 启动通知（企业微信应用消息）
- [√] 1.1 在 `internal/app/server.go` 中支持使用 `net.Listener` 启动 HTTP 服务（`Serve(listener)` 或等效方式），确保端口绑定成功后再进入服务循环，验证 why.md#需求-启动成功通知-场景-正常启动后发送一次通知
- [√] 1.2 在 `cmd/wecom-home-ops/main.go` 中在监听成功后向 `auth.allowed_userids` 发送一次启动成功文本消息（发送失败仅记录日志），验证 why.md#需求-启动成功通知-场景-正常启动后发送一次通知
- [√] 1.3 补齐启动消息的“详细信息”构造逻辑，并确保不包含任何敏感字段（仅输出脱敏/非敏感信息）

## 2. 安全检查
- [√] 2.1 执行安全检查（按G9：敏感信息处理、权限控制、外部调用失败处理；确认不输出 secret/token/api_key 等）

## 3. 文档更新
- [√] 3.1 更新 `helloagents/wiki/modules/wecom.md`：补充“启动成功通知”行为说明
- [√] 3.2 更新 `helloagents/CHANGELOG.md`：记录新增启动通知能力

## 4. 测试
- [√] 4.1 运行 `go test ./...`
