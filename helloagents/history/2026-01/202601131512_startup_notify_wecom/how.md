# 技术设计: 启动成功通知（企业微信应用消息）

## 技术方案

### 核心技术
- Go `net.Listener`：通过显式 `net.Listen` 确认端口绑定成功，作为“启动成功/就绪”的判定点
- 复用 `internal/wecom.Client` 的 `SendText`：发送企业微信自建应用文本消息（`message/send`）

### 实现要点
- **启动就绪判定**
  - 将监听行为从 `http.Server.ListenAndServe()` 调整为先 `net.Listen` 再 `http.Server.Serve(listener)`，避免在端口占用等场景误报“启动成功”。
- **通知发送**
  - 在监听成功后异步发送（独立 goroutine），对每个 `auth.allowed_userids` 发送一次文本消息。
  - 使用 `context.WithTimeout` 控制单次通知的最长期限，避免启动阶段被外部网络阻塞。
- **消息内容（脱敏）**
  - 关键字段：服务名、启动时间/耗时、hostname/pid、配置的 `server.listen_addr` 与实际监听地址、`server.base_url`（如有）、`/healthz` `/readyz`、`/wecom/callback`。
  - 避免输出：`wecom.secret`、`wecom.token`、`wecom.encoding_aes_key`、`unraid.api_key`、`qinglong.client_secret` 等任何密钥信息。

## 安全与性能
- **安全:** 仅发送非敏感诊断信息；失败日志不包含敏感字段；不引入新的外部依赖或权限变更。
- **性能:** 启动阶段通知发送异步执行；按用户列表顺序发送，避免并发放大上游限流风险。

## 测试与部署
- **测试:** 运行 `go test ./...`；重点关注编译、现有 wecom/client 与 core/router 相关测试是否通过。
- **部署:** 无额外部署步骤；更新后二进制/容器启动后将自动发送一次启动通知。
