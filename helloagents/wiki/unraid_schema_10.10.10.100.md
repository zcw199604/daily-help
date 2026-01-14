# Unraid GraphQL Schema（10.10.10.100）

## 来源与抓取方式
- Endpoint: `http://10.10.10.100/graphql`
- 抓取时间: 2026-01-14
- 抓取方式: GraphQL Introspection（Apollo Server CSRF 保护：需使用 `POST`，并带 `Content-Type: application/json` + `x-apollo-operation-name`）

### 最小验证（示例）
```bash
curl -sS -H 'Content-Type: application/json' -H 'x-apollo-operation-name: Introspection' \
  --data '{"query":"query { __schema { queryType { name } mutationType { name } subscriptionType { name } } }"}' \
  http://10.10.10.100/graphql
```

## Schema 概览
- 根类型:
  - `Query`
  - `Mutation`
  - `Subscription`
- 类型数量（introspection 统计）: `164`

---

## Query（46）

### 鉴权 / API Key / OIDC
- `apiKeys: [ApiKey!]!`
- `apiKey(id: PrefixedID!): ApiKey`
- `apiKeyPossibleRoles: [Role!]!`
- `apiKeyPossiblePermissions: [Permission!]!`
- `getPermissionsForRoles(roles: [Role!]!): [Permission!]!`
- `previewEffectivePermissions(roles: [Role!], permissions: [AddPermissionInput!]): [Permission!]!`
- `getAvailableAuthActions: [AuthAction!]!`
- `getApiKeyCreationFormSchema: ApiKeyFormSettings!`
- `me: UserAccount!`
- `isSSOEnabled: Boolean!`
- `publicOidcProviders: [PublicOidcProvider!]!`
- `oidcProviders: [OidcProvider!]!`
- `oidcProvider(id: PrefixedID!): OidcProvider`
- `oidcConfiguration: OidcConfiguration!`
- `validateOidcSession(token: String!): OidcSessionValidation!`

### 系统 / 服务器 / 配置 / 日志 / 指标
- `config: Config!`
- `flash: Flash!`
- `online: Boolean!`
- `owner: Owner!`
- `registration: Registration`
- `server: Server`
- `servers: [Server!]!`
- `services: [Service!]!`
- `vars: Vars!`
- `isInitialSetup: Boolean!`
- `customization: Customization`
- `publicPartnerInfo: PublicPartnerInfo`
- `publicTheme: Theme!`
- `rclone: RCloneBackupSettings!`
- `info: Info!`
- `logFiles: [LogFile!]!`
- `logFile(path: String!, lines: Int, startLine: Int): LogFileContent!`
- `settings: Settings!`
- `metrics: Metrics!`

### 存储 / 阵列 / 校验 / 共享
- `array: UnraidArray!`
- `parityHistory: [ParityCheck!]!`
- `disks: [Disk!]!`
- `disk(id: PrefixedID!): Disk!`
- `shares: [Share!]!`

### Docker / VM / 通知 / UPS / 插件
- `docker: Docker!`
- `vms: Vms!`
- `notifications: Notifications!`
- `upsDevices: [UPSDevice!]!`
- `upsDeviceById(id: String!): UPSDevice`
- `upsConfiguration: UPSConfiguration!`
- `plugins: [Plugin!]!`

---

## Mutation（22）

### 通知
- `createNotification(input: NotificationData!): Notification!`
- `deleteNotification(id: PrefixedID!, type: NotificationType!): NotificationOverview!`
- `deleteArchivedNotifications: NotificationOverview!`
- `archiveNotification(id: PrefixedID!): Notification!`
- `archiveNotifications(ids: [PrefixedID!]!): NotificationOverview!`
- `archiveAll(importance: NotificationImportance): NotificationOverview!`
- `unreadNotification(id: PrefixedID!): Notification!`
- `unarchiveNotifications(ids: [PrefixedID!]!): NotificationOverview!`
- `unarchiveAll(importance: NotificationImportance): NotificationOverview!`
- `recalculateOverview: NotificationOverview!`

### 域操作（Mutations 容器）
- `array: ArrayMutations!`
- `docker: DockerMutations!`
- `vm: VmMutations!`
- `parityCheck: ParityCheckMutations!`
- `apiKey: ApiKeyMutations!`
- `customization: CustomizationMutations!`
- `rclone: RCloneMutations!`

### 其他
- `initiateFlashBackup(input: InitiateFlashBackupInput!): FlashBackupStatus!`
- `updateSettings(input: JSON!): UpdateSettingsResponse!`
- `configureUps(config: UPSConfigInput!): Boolean!`
- `addPlugin(input: PluginManagementInput!): Boolean!`
- `removePlugin(input: PluginManagementInput!): Boolean!`

---

## Subscription（11）
- `notificationAdded: Notification!`
- `notificationsOverview: NotificationOverview!`
- `ownerSubscription: Owner!`
- `serversSubscription: Server!`
- `parityHistorySubscription: ParityCheck!`
- `arraySubscription: UnraidArray!`
- `logFile(path: String!): LogFileContent!`
- `systemMetricsCpu: CpuUtilization!`
- `systemMetricsCpuTelemetry: CpuPackages!`
- `systemMetricsMemory: MemoryUtilization!`
- `upsUpdates: UPSDevice!`

---

## 关键模块展开（按 root 入口）

### Docker（`Query.docker` / `Mutation.docker`）

#### `Docker`
- `id: PrefixedID!`
- `containers(skipCache: Boolean! = false): [DockerContainer!]!`
- `networks(skipCache: Boolean! = false): [DockerNetwork!]!`

#### `DockerMutations`
- `start(id: PrefixedID!): DockerContainer!`
- `stop(id: PrefixedID!): DockerContainer!`

> 结论：该实例的 `DockerMutations` **不包含** `updateContainer/update` 等“更新容器/拉取镜像”能力，因此 GraphQL 无法直接更新容器。

### VM（`Query.vms` / `Mutation.vm`）

#### `Vms`
- `id: PrefixedID!`
- `domains: [VmDomain!]`
- `domain: [VmDomain!]`

#### `VmMutations`
- `start(id: PrefixedID!): Boolean!`
- `stop(id: PrefixedID!): Boolean!`
- `pause(id: PrefixedID!): Boolean!`
- `resume(id: PrefixedID!): Boolean!`
- `forceStop(id: PrefixedID!): Boolean!`
- `reboot(id: PrefixedID!): Boolean!`
- `reset(id: PrefixedID!): Boolean!`

### Array / Parity Check（`Query.array` / `Mutation.array` / `Mutation.parityCheck`）

#### `UnraidArray`
- `id: PrefixedID!`
- `state: ArrayState!`
- `capacity: ArrayCapacity!`
- `boot: ArrayDisk`
- `parities: [ArrayDisk!]!`
- `disks: [ArrayDisk!]!`
- `caches: [ArrayDisk!]!`

#### `ArrayMutations`
- `setState(input: ArrayStateInput!): UnraidArray!`
- `addDiskToArray(input: ArrayDiskInput!): UnraidArray!`
- `removeDiskFromArray(input: ArrayDiskInput!): UnraidArray!`
- `mountArrayDisk(id: PrefixedID!): ArrayDisk!`
- `unmountArrayDisk(id: PrefixedID!): ArrayDisk!`
- `clearArrayDiskStatistics(id: PrefixedID!): Boolean!`

#### `ParityCheckMutations`
- `start(correct: Boolean!): JSON!`
- `pause: JSON!`
- `resume: JSON!`
- `cancel: JSON!`

### API Key（`Query.apiKeys` / `Mutation.apiKey`）

#### `ApiKeyMutations`
- `create(input: CreateApiKeyInput!): ApiKey!`
- `addRole(input: AddRoleForApiKeyInput!): Boolean!`
- `removeRole(input: RemoveRoleFromApiKeyInput!): Boolean!`
- `delete(input: DeleteApiKeyInput!): Boolean!`
- `update(input: UpdateApiKeyInput!): ApiKey!`

