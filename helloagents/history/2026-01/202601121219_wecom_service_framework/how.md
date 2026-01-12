# 技术设计: 企业微信多服务插件框架 + 青龙(QL)对接

## 技术方案

### 核心技术
- Go 1.22 / `net/http`
- 复用现有企业微信回调验签解密与消息发送（`internal/wecom`）
- 新增青龙 OpenAPI HTTP 客户端（不引入额外第三方依赖，优先标准库）

### 实现要点
- **Provider 抽象与注册表**
  - 定义 `ServiceProvider`（或等价命名）接口：声明服务 key、展示名、入口关键词、处理文本/事件的能力
  - `core.Router` 负责：白名单鉴权、统一会话状态存取、把消息分发给选中的 Provider
  - Provider 负责：构建卡片、解析事件 key、执行业务动作（调用后端 API）
- **会话状态机通用化**
  - 在 `ConversationState` 中增加通用字段：`ServiceKey`、`InstanceID`、`Action`、`TargetID`、`ExpiresAt`
  - 为不同 Provider 保留扩展载荷（如可用 `map[string]string` 或结构化子状态），但优先保持轻量
- **事件 key 编码策略**
  - 统一命名空间：`svc.*` / `unraid.*` / `qinglong.*`
  - 动态参数（cron id、instance id）采用“前缀+分隔符”编码，解析时做白名单校验与类型校验
- **青龙 OpenAPI 客户端**
  - 每实例维护 token 缓存：`token + expiration`，过期前提前刷新
  - 统一响应解码：青龙 API 通常返回 `{code, data, message, errors}`；封装为可诊断错误
  - 方法集合（按需求覆盖）：
    - 获取 token：`GET /open/auth/token?client_id=...&client_secret=...`
    - 任务列表/搜索：`GET /open/crons?...`
    - 运行任务：`PUT /open/crons/run`
    - 启用任务：`PUT /open/crons/enable`
    - 禁用任务：`PUT /open/crons/disable`
    - 最近日志：`GET /open/crons/{id}/log`
- **企业微信交互**
  - 新增“服务选择卡片”：展示 Unraid / 青龙
  - 青龙 Provider 内：实例选择卡片 → 动作卡片 →（列表/搜索结果卡片）→ 确认卡片
  - 对长日志/长列表：做摘要与分页；必要时引导用户输入关键词或 ID

## 架构决策 ADR

### ADR-003: 引入 Provider 插件框架（已采纳）
**上下文:** 需要在现有 Unraid 基础上新增青龙对接，并支持未来更多服务，且要支持多实例。  
**决策:** 将服务接入抽象为 Provider，核心路由只做鉴权、状态与分发。  
**理由:** 降低核心路由复杂度；新服务新增文件即可接入；多实例与交互流程更易复用。  
**替代方案:** 在现有 `core.Router` 内继续堆叠分支逻辑 → 拒绝原因: 随服务数量增长可维护性显著下降。  
**影响:** 短期改动面扩大，需要为兼容性与分发逻辑补充测试。

## 安全与性能
- **安全:**
  - 白名单鉴权继续作为第一道防线
  - 青龙 `client_secret`、token 不写入日志与回显
  - 事件 key 与用户输入严格校验，避免越权操作（如伪造 cron id/instance id）
- **性能:**
  - token 缓存减少频繁鉴权调用
  - 列表查询默认分页/限制返回数量，避免一次回显过大

## 测试与部署
- **测试:**
  - `internal/qinglong` 使用 `httptest` 覆盖 token 刷新、列表/运行/启用/禁用/日志等请求与响应解析
  - `internal/core` 增加 Provider 分发与关键状态流转的单元测试（包含兼容 Unraid 入口）
- **部署:**
  - 更新 `config.yaml` 增加 `qinglong.instances`
  - 保持 HTTP 回调路径不变（`/wecom/callback`），升级后直接替换二进制/容器即可
