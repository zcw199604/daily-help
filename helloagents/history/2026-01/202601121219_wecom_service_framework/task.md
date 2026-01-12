# 任务清单: 企业微信多服务插件框架 + 青龙(QL)对接

目录: `helloagents/plan/202601121219_wecom_service_framework/`

---

## 1. Provider 框架（core）
- [√] 1.1 在 `internal/core/router.go` 中引入 Provider 注册与分发（服务选择/直达入口），验证 why.md#需求-多服务路由框架-场景-入口菜单选择服务
- [√] 1.2 在 `internal/core/state.go` 中扩展会话状态以支持 ServiceKey/InstanceID/目标任务等，验证 why.md#需求-多服务路由框架-场景-入口菜单选择服务

## 2. Unraid 迁移兼容（unraid/core）
- [√] 2.1 将现有 Unraid 交互封装为 Provider（或等价适配层），保持“容器/unraid”直达，验证 why.md#需求-多服务路由框架-场景-兼容现有-unraid-入口

## 3. 青龙 OpenAPI 客户端（qinglong）
- [√] 3.1 新增 `internal/qinglong/client.go` 实现 token 获取与缓存刷新，验证 why.md#需求-青龙多实例管理-场景-选择实例
- [√] 3.2 在 `internal/qinglong/client.go` 实现任务列表/搜索与日志获取，验证 why.md#需求-青龙多实例管理-场景-查询-搜索任务 与 why.md#需求-青龙多实例管理-场景-查看最近日志
- [√] 3.3 在 `internal/qinglong/client.go` 实现运行/启用/禁用任务 API 调用，验证 why.md#需求-青龙多实例管理-场景-运行任务 与 why.md#需求-青龙多实例管理-场景-启用-禁用任务

## 4. 配置与装配（config/app）
- [√] 4.1 扩展 `internal/config/config.go` 增加 `qinglong.instances` 配置结构并校验，更新 `config.example.yaml`，验证 why.md#需求-青龙多实例管理-场景-选择实例
- [√] 4.2 在 `internal/app/server.go` 中完成 Provider 组装与注入（包含多实例青龙与 Unraid），验证 why.md#需求-多服务路由框架-场景-入口菜单选择服务

## 5. 企业微信交互（wecom）
- [√] 5.1 在 `internal/wecom/message.go` 中新增服务选择卡片与青龙卡片（实例选择/动作菜单/确认），验证 why.md#需求-多服务路由框架-场景-入口菜单选择服务
- [√] 5.2 实现青龙任务选择交互（列表/搜索→选择→确认→执行回显），验证 why.md#需求-青龙多实例管理-场景-查询-搜索任务

## 6. 安全检查
- [√] 6.1 执行安全检查（按G9: 输入验证、敏感信息处理、权限控制、越权防护）

## 7. 文档更新（知识库）
- [√] 7.1 新增 `helloagents/wiki/modules/qinglong.md` 并更新 `helloagents/wiki/overview.md`、`helloagents/wiki/api.md`、`helloagents/wiki/arch.md`（如需要）
- [√] 7.2 更新 `helloagents/CHANGELOG.md`

## 8. 测试
- [√] 8.1 为 `internal/qinglong/client.go` 增加 `httptest` 单元测试（token/列表/运行/启用禁用/日志）
- [√] 8.2 为 Provider 分发与关键状态流转增加单元测试（包含 Unraid 兼容入口）
