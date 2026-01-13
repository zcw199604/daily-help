# 任务清单（轻量迭代）

> 目标：补齐企业微信“更新模版卡片消息（ResponseCode）”能力：文档可复制示例 + 代码实现可用。

- [ ] 获取官方接口信息（SSOT：更新模版卡片消息 https://developer.work.weixin.qq.com/document/90000/90135/94888）
- [√] 获取官方接口信息（SSOT：更新模版卡片消息 https://developer.work.weixin.qq.com/document/90000/90135/94888）
- [√] 回调入站结构补齐 `ResponseCode` 字段解析（MsgType=event, Event=template_card_event）
- [√] wecom client 实现 `message/update_template_card` 调用（至少支持 `button.replace_name` 场景）
- [√] core 路由在收到模板卡片回调后消费 `ResponseCode`，将按钮更新为不可点击（避免重复点击/与 TaskId 去重一致）
- [√] 知识库文档补充：`helloagents/wiki/wecom_interaction.md` 增加“更新模版卡片消息”章节与可复制请求示例（curl/JSON）
- [√] 补充/更新单元测试覆盖 `update_template_card` 请求构造与回调字段解析
- [√] Docker 环境运行 `go test ./...` 通过
- [√] 知识库同步：更新 `helloagents/CHANGELOG.md`，并迁移方案包到 `helloagents/history/2026-01/`
- [√] 提交并 push
