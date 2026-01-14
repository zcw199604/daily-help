# Unraid 官方 API（GraphQL）要点整理

## 来源
本页内容基于 Unraid 官方文档 `docs.unraid.net` 的 API 专题整理（检索时间：2026-01-14）：
- https://docs.unraid.net/API/
- https://docs.unraid.net/API/how-to-use-the-api/
- https://docs.unraid.net/API/cli/
- https://docs.unraid.net/API/programmatic-api-key-management/
- https://docs.unraid.net/API/api-key-app-developer-authorization-flow/
- https://docs.unraid.net/API/oidc-provider-setup/
- https://docs.unraid.net/API/upcoming-features/

## API 概述
- Unraid API 提供 **GraphQL** 接口，用于以程序化方式访问/管理 Unraid 服务器能力（自动化、监控、集成）。
- 支持多种鉴权方式：**API Key / WebGUI Session Cookie / SSO(OIDC)**。
- 提供开发者工具（GraphQL Sandbox / CLI 等）。

## 可用性（版本/安装方式）
- **Unraid 7.2+**：API 原生集成在操作系统中，无需额外插件；入口在 WebGUI：
  - `Settings → Management Access → API`
- **Unraid 7.2 之前 / 或需要更新功能**：可安装 **Unraid Connect 插件** 获取 API 能力：
  - 插件为 pre-7.2 提供 API
  - 本地使用 API **不要求登录 Unraid Connect**
  - 在 7.2+ 上安装插件，可提前使用尚未随 OS 发布的新 API 特性

## Endpoint 与开发工具

### GraphQL 访问入口
- GraphQL（含 Playground/Sandbox）地址示例：
  - `http://<YOUR_SERVER_IP>/graphql`

### 启用 GraphQL Sandbox
官方文档给出两种方式：
- WebGUI（推荐）：在 `Settings → Management Access → API` 启用 GraphQL Sandbox 开关
- CLI：
  - `unraid-api developer --sandbox true`（开启）
  - `unraid-api developer --sandbox false`（关闭）

## 鉴权方式与 API Key 管理

### 鉴权方式（官方文档列举）
多数 Query/Mutation 需要鉴权，官方文档列出：
- **API Keys**：面向程序化接入
- **Cookies**：在 WebGUI 登录后的浏览器会话自动携带
- **SSO/OIDC**：配置第三方身份提供商后使用

### API Key 在 GraphQL 请求中的用法
官方文档给出 GraphQL Header 示例：
- `x-api-key: <YOUR_API_KEY>`

### WebGUI 管理 API Key
官方文档路径：
- `Settings → Management Access → API Keys`
  - 查看/创建/配置权限与角色
  - 撤销或重新生成

### CLI 管理 API Key（unraid-api）
官方文档：
- 命令概览：`unraid-api <command> [options]`
- API key 相关：`unraid-api apikey [options]`
  - 关键参数（示例）：`--name`、`--create`、`--roles`、`--permissions`、`--description`
- 开发者模式：`unraid-api developer ...`（可启用 GraphQL Sandbox）
- SSO 管理：`unraid-api sso ...`（add/remove/list/validate-token）

### 程序化创建/删除 API Key（CLI + JSON）
官方文档给出可用于自动化/脚本的示例：
- 创建（JSON 输出）：
  - `unraid-api apikey --create --name "workflow key" --roles ADMIN --json`
- 使用细粒度权限创建：
  - `unraid-api apikey --create --name "limited access key" --permissions "DOCKER:READ_ANY,ARRAY:READ_ANY" --description "Read-only access for monitoring" --json`
- 覆盖同名 key（会使旧 key 立即失效）：
  - `unraid-api apikey --create --name "existing key" --roles ADMIN --overwrite --json`
- 删除：
  - `unraid-api apikey --delete --name "workflow key"`
  - `unraid-api apikey --delete --name "workflow key" --json`

### 常见报错（官方文档举例）
文档给出 CLI 常见错误提示（示例）：
- API key 名称字符限制（仅字母数字与空格，不允许特殊字符）
- 同名 key 已存在（使用 `--overwrite` 或改名）
- 未指定 role/permission（至少一个）

## 应用开发者授权流程（ApiKeyAuthorize）
官方文档提供一个“授权页”流程（类似 OAuth 授权交互，但返回的是 API key）：
- 入口：
  - `https://<unraid-server>/ApiKeyAuthorize?name=MyApp&scopes=docker:read,vm:*&redirect_uri=https://myapp.com/callback&state=abc123`
- Query 参数：
  - `name`（必填）
  - `description`（可选）
  - `scopes`（必填，逗号分隔）
  - `redirect_uri`（可选）
  - `state`（可选）
- 授权后回调示例：
  - `https://myapp.com/callback?api_key=xxx&state=abc123`
- scope 格式（官方文档描述）：
  - `resource:action`（例如 `docker:read`、`vm:*`、`system:update`、`role:viewer`、`role:admin`）
  - 动作：`create/read/update/delete/*`
  - `redirect_uri` 要求 HTTPS（localhost 开发例外），API key 只展示一次需安全保存

## OIDC/SSO 配置要点
官方文档路径：
- `Settings → Management Access → API → OIDC`

### 授权模式
- Simple mode（推荐）：按邮箱域名/邮箱白名单授权（配置简单）
- Advanced mode：基于 JWT claim 规则（equals/contains/endsWith/startsWith 等），支持 OR/AND 规则组合

### 必须配置的 Redirect URI
官方文档要求所有 provider 使用固定格式：
- `http://<YOUR_UNRAID_IP>/graphql/api/auth/oidc/callback`

### Issuer URL（安全建议）
官方建议优先使用 **base URL**（系统会自动拼接 `/.well-known/openid-configuration`），避免直接填 discovery URL 导致 issuer 校验降级风险。

## GraphQL Schema 与示例 Query

### 可覆盖的业务域（官方文档概括）
官方文档列举 API 可覆盖范围（摘要）：
- System information：系统/硬件/健康信息
- Array management：阵列状态、磁盘健康、Parity check 等
- Docker management：容器列表/状态、网络管理等
- Remote access：远程访问配置、SSO 配置、Allowed origins 管理等

### 示例 Query（官方文档提供）
- 系统信息：
  - `query { info { os { platform distro release uptime } cpu { manufacturer brand cores threads } } }`
- 阵列状态：
  - `query { array { state capacity { disks { free used total } } disks { name size status temp } } }`
- Docker 容器列表：
  - `query { dockerContainers { id names state status autoStart } }`

## 错误处理与限流
- 官方文档说明 API 具备 rate limiting，需要客户端正确处理限流响应。
- GraphQL 错误返回格式示例：
  - `{"errors":[{"message":"...","locations":[...],"path":[...]}]}`

## 对 wecom-home-ops 的影响（核对结论）
- ✅ 当前实现的基础接入方式与官方文档一致：`/graphql` + `x-api-key` + 可选 `Origin`（见 `internal/unraid/client.go`）
- ⚠️ Schema 可能存在版本差异：官方示例 Query 使用 `dockerContainers` 顶层字段；而 wecom-home-ops 当前使用 `docker { containers { ... } }` 形态。若目标环境为 Unraid 7.2+ 官方 API 且不兼容该结构，可能需要为“容器列表/查找 ID”增加 Query 形态回退或做 schema 探测/配置化切换。
- ⚠️ 官方文档页未直接列出“更新容器/拉取镜像”相关 mutation；若目标环境的 `DockerMutations` 不提供 `updateContainer/update` 等字段，可：
  - 以 Apollo Live Documentation（官方文档给出的链接）为准确认是否存在替代字段
  - 或改用 WebGUI 内部接口 `webGui/include/StartCommand.php` 执行 `update_container <name>`（需要 Cookie/CSRF；wecom-home-ops 已提供可选兜底配置：`unraid.webgui_csrf_token` / `unraid.webgui_cookie` / `unraid.webgui_command_url`）

## 目标实例差异快照（10.10.10.100）
- Schema 摘要见：`helloagents/wiki/unraid_schema_10.10.10.100.md`
- 关键差异（与“官方示例/常见预期”对照）：
  - ✅ 容器列表走 `docker { containers(...) { ... } }`
  - ⚠️ `DockerMutations` 仅包含 `start/stop`，不支持 `updateContainer/update`（因此更新容器需走 WebGUI 兜底）
