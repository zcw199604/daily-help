# unraid_mobile_ui（Unraid Mobile UI）

## 来源
- 仓库: https://github.com/s3ppo/unraid_mobile_ui
- 调研基线: `e323b28ba2763218dd50a82db1892efcfeb48714`（main，检索时间：2026-01-14）
- 技术栈: Flutter（Dart）+ GraphQL client（Query/Mutation + Subscription）

## 项目定位（上游描述摘要）
上游 README 描述其为移动端 GraphQL 客户端，用于展示 Unraid 服务器数据；待 Unraid Connect 插件提供更多能力后再考虑发布到应用商店。

## 功能清单（按页面/模块）
- 登录/连接:
  - 录入服务器局域网 IP（Lan IP）与 Token
  - 支持选择协议（http/https），订阅场景使用 ws/wss
  - 多服务器（Multiserver Settings）管理与切换
- Dashboard:
  - 服务器卡片（Server Card）
  - Array/Info/Parity/UPS 卡片
  - 通知未读与通知列表摘要
  - CPU 指标订阅（`systemMetricsCpu`）
- Dockers:
  - Docker 容器列表
  - 容器 Start/Stop
  - ⚠️ 未提供“更新容器/拉取镜像”操作（上游 GraphQL mutation 列表中未包含 update/updateContainer）
- VMs:
  - 虚拟机列表
  - Start/Stop/Pause/Resume/ForceStop/Reboot/Reset
- Array:
  - 阵列状态展示
  - 设置阵列状态（ArrayStateInput）
- Shares / System / Plugins / Notifications:
  - 共享列表、系统信息（含 Baseboard/CPU/Memory/OS 子页）、插件信息、通知列表（支持 filter）

## API 使用方法（GraphQL）

### Endpoint 约定
- HTTP（Query/Mutation）: `http(s)://<ip>/graphql`
- WebSocket（Subscription）: `ws(s)://<ip>/graphql`

### 鉴权与 Header
- `x-api-key: <token>`：核心鉴权方式（Token 来自用户输入并持久化）
- `Origin: <packageName>`：客户端会附带 Origin（值为 packageName 变量；上游代码以表达式形式注入）

### 客户端侧持久化（多服务器/连接信息）
上游代码使用 SharedPreferences 存储（便于多服务器与重连）：
- `ip`
- `prot`（协议选择：`http`/`https`）
- `token`
- `multiservers`（多服务器列表）

## GraphQL 操作清单（与页面关联）

### 页面 → 调用关系（按上游代码扫描）
- Dashboard:
  - Queries: `getServerCard` / `getArrayCard` / `getInfoCard` / `getParityCard` / `getUpsCard` / `getNotificationsUnread` / `getNotifications`
  - Subscriptions: `getCpuMetrics`
- Dockers:
  - Queries: `getDockers`
  - Mutations: `startDocker` / `stopDocker`
- VMs:
  - Queries: `getVms`
  - Mutations: `startVM` / `stopVM` / `pauseVM` / `resumeVM` / `forceStopVM` / `rebootVM` / `resetVM`
- Array:
  - Queries: `getArray`
  - Mutations: `setArrayState`
- Shares:
  - Queries: `getShares`
- System:
  - Queries: `getInfo`
- Plugins:
  - Queries: `getPlugins`
- Notifications:
  - Queries: `getNotifications`

### 变量类型与关键约定（按上游 GraphQL 文本定义）
- Docker/VM 操作：`PrefixedID!`
- Array 状态变更：`ArrayStateInput!`
- 通知查询：`NotificationFilter!`（变量名 `filter`）
- Parity Check mutation（上游定义但不一定在 UI 中暴露）：`correct: Boolean!`

## 抓包差异排查提示（重要）
如果你抓包看到类似：
- `http://<unraid>/webGui/include/StartCommand.php`
- `Content-Type: application/x-www-form-urlencoded`
- 表单参数包含 `csrf_token=...`、`cmd=update_container ...`

这更像是 **Unraid WebGUI 的内部命令接口**（依赖 Cookie/CSRF），而非 `/graphql` + `x-api-key` 的 GraphQL 调用。

在上述基线版本源码中未检索到 `StartCommand.php`/`update_container` 的直接调用；因此更可能是：
- 实际操作发生在 Unraid WebGUI 页面（或 App 内嵌/跳转的 WebGUI），而不是 App 自身的 GraphQL 层发起
- 结论：**unraid_mobile_ui 本身不支持更新容器**；更新容器通常走 WebGUI 内部接口（StartCommand.php）

## 对 wecom-home-ops 的参考价值（可选）
- 可作为 Unraid GraphQL 使用形态的参考样例：`/graphql` + `x-api-key` +（可选）`Origin`，订阅使用 `ws(s)` 同路径。
- 覆盖的业务域（VM/Array/Parity/UPS/Notifications/Plugins）可用于 wecom-home-ops 后续扩展时的需求清单参考。
