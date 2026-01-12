# 任务清单: 企业微信对接 Unraid 容器管理（MVP）

目录: `helloagents/plan/202601120816_wecom_unraid/`

---

## 1. 项目骨架
- [√] 1.1 初始化 Go Module 与目录结构（`cmd/`、`internal/`），实现基础 HTTP Server 与配置加载，验证 why.md#需求-unraid-容器管理-场景-重启容器
- [√] 1.2 增加健康检查接口与结构化日志骨架，验证 why.md#需求-安全与可追溯-场景-非授权用户访问

## 2. 企业微信回调（wecom）
- [√] 2.1 实现回调 URL 校验（GET）与签名校验、消息解密（POST），并为验签/解密编写单元测试，验证 why.md#需求-安全与可追溯
- [√] 2.2 实现 access_token 获取与缓存、发送文本消息/卡片消息封装，并编写最小单测（HTTP mock），验证 why.md#需求-unraid-容器管理

## 3. 会话状态机与交互卡片（core）
- [√] 3.1 实现“选择动作 → 录入参数 → 二次确认 → 执行/取消/超时”的状态机，验证 why.md#需求-unraid-容器管理
- [√] 3.2 实现卡片/按钮事件解析与路由到状态机（包含取消与超时提示），验证 why.md#需求-unraid-容器管理

## 4. Unraid 适配器
- [√] 4.1 定义适配器接口与 Unraid GraphQL 客户端抽象，实现容器重启/停止，验证 why.md#需求-unraid-容器管理-场景-重启容器
- [√] 4.2 基于 GraphQL introspection 确认“强制更新”相关 mutation 能力（update/pull/recreate 等），并实现强制更新；如能力缺失则返回可诊断的降级提示，验证 why.md#需求-unraid-容器管理-场景-强制更新容器

## 5. 安全检查
- [√] 5.1 输入验证、敏感信息处理、权限控制、危险操作二次确认策略（按G9），验证 why.md#需求-安全与可追溯

## 6. 文档更新
- [√] 6.1 同步更新知识库：`wiki/arch.md`、`wiki/api.md`、`wiki/modules/*.md`、`CHANGELOG.md`

## 7. 测试
- [√] 7.1 补充关键单元测试并确保 `go test ./...` 通过（回调验签、状态机、Unraid 执行器 mock）
