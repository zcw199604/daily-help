# 任务清单: 企业微信自建应用会话交互速查表

目录: `helloagents/plan/202601130629_wecom_interaction_kb/`

---

## 1. 官方文档调研
- [√] 1.1 查阅并摘取企业微信官方文档关键信息：回调配置（GET/POST 验证与响应时限/重试语义）、事件推送（enter_agent、template_card_event）、发送应用消息（template_card/button_interaction）、获取access_token

## 2. 知识库整理
- [√] 2.1 新增 `helloagents/wiki/wecom_interaction.md`：开发者速查表 + 可复制示例 + 与本项目代码映射
- [√] 2.2 更新 `helloagents/wiki/modules/wecom.md`：添加专题文档入口链接

## 3. 知识库同步
- [√] 3.1 更新 `helloagents/CHANGELOG.md`（Unreleased）记录文档新增

## 4. 自检
- [√] 4.1 校验文档链接可用、示例与现有实现一致（回调事件名、字段名、task_id 约束等）
