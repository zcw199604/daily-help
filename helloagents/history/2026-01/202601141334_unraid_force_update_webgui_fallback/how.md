# 技术设计: Unraid 容器强制更新 WebGUI 兜底

## 技术方案

### 核心技术
- Go（wecom-home-ops 现有实现）
- HTTP：GraphQL（JSON）+ WebGUI StartCommand（x-www-form-urlencoded）

### 实现要点
- **GraphQL 首选**：继续通过 `internal/unraid/client.go` 的 `callDockerForceUpdateMutation` 发送 `docker { <mutation>(<arg>:$v) { ... } }`。
- **兜底触发条件**：仅当错误包含以下特征之一时视为“不支持/Schema 不匹配”，才允许进入 WebGUI 兜底，避免重复执行真实更新操作：
  - `Cannot query field`
  - `Unknown argument`
  - `Unknown type`
  - `graphql error:`（GraphQL errors 数组）
- **WebGUI 兜底执行**：
  - URL：`unraid.webgui_command_url`（默认由 `unraid.endpoint` 推导：`/graphql` → `/webGui/include/StartCommand.php`）
  - Body（form）：`cmd=update_container <name>&start=0&csrf_token=<token>`
  - Header（可选）：`Cookie: <cookie>`（当 WebGUI 需要登录时）
- **错误识别**：
  - HTTP 非 2xx：返回 `unraid webgui http status ...`
  - Body 含 `invalid csrf`：提示 csrf_token 无效或过期
  - Content-Type 为 HTML 且出现 login/password：提示可能未登录或 csrf_token 失效

## 架构设计
无结构性调整，仅在 unraid Client 内增加一条“强制更新”的备用执行路径。

## 架构决策 ADR
### ADR-001: 强制更新增加 WebGUI StartCommand 兜底
**上下文:** GraphQL schema 在部分 Unraid 版本中不提供容器更新 mutation，但 WebGUI 存在可用内部命令接口。需要在不破坏现有 GraphQL 行为的前提下提升“强制更新”的成功率。
**决策:** GraphQL 作为首选；当 GraphQL 明确提示“不支持”时，使用 WebGUI StartCommand.php 执行 `update_container` 兜底。
**理由:** 变更最小、对现有部署兼容；兜底仅在“未执行更新”的校验错误场景触发，避免重复操作风险。
**替代方案:** 仅依赖 GraphQL 配置覆盖 → 拒绝原因: 在目标环境缺失 mutation 时无法解决；且用户已观察到 WebGUI 可用路径。
**影响:** 新增可选敏感配置（Cookie/CSRF），需避免日志泄露；WebGUI 内部接口可能随版本变化，需要文档说明与错误提示。

## API设计
无对外 API 变更。

## 数据模型
无数据模型变更；新增配置项：
- `unraid.webgui_command_url`（可选）
- `unraid.webgui_csrf_token`（启用兜底必填）
- `unraid.webgui_cookie`（可选）

## 安全与性能
- **安全:** 不记录 Cookie/CSRF；兜底仅在 GraphQL 校验错误时触发；错误信息仅提示配置项名称，不回显敏感值。
- **性能:** 兜底仅在失败场景追加一次 HTTP 调用；默认路径无额外开销。

## 测试与部署
- **测试:** 增加单测覆盖“GraphQL 不支持 → WebGUI 兜底成功”的路径。
- **部署:** 如需启用兜底，按需在 `config.yaml` 增加上述 webgui 配置项；否则不受影响。
