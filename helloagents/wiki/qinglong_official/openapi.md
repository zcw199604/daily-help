# 青龙(QL) 官方 API / OpenAPI 接口清单（基于源码提取）

- 上游仓库: https://github.com/whyour/qinglong
- 上游提交: `d53437d1695d22db266cea3b680d3d7663ce86a6`
- 生成时间: 2026-01-13 05:18:41

## 1. 路径前缀与鉴权说明

- 后端真实挂载前缀为 `/api`；同时提供 `/open` 入口（服务端会将 `/open/*` rewrite 到 `/api/*`，接口实现一致）。
- OpenAPI（应用 token）鉴权：通过 `/open/auth/token` 使用 `client_id` + `client_secret` 换取 token，随后以 `Authorization: Bearer <token>` 调用。
- OpenAPI scope：上游 `AppScope` 定义为 `envs | crons | configs | scripts | logs | system`；调用 `/open/<scope>/...` 时需确保应用已授予对应 scope。
- 说明：上游会忽略 query 参数 `t`（会被全局中间件删除），可用于客户端规避缓存但对服务端无影响。

## 2. 接口清单（按模块）

### configs（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| GET | `/configs/:file` | `/open/configs/:file` | `/api/configs/:file` |
| GET | `/configs/detail` | `/open/configs/detail` | `/api/configs/detail` |
| GET | `/configs/files` | `/open/configs/files` | `/api/configs/files` |
| GET | `/configs/sample` | `/open/configs/sample` | `/api/configs/sample` |
| POST | `/configs/save` | `/open/configs/save` | `/api/configs/save` |

来源文件: `back/api/config.ts`

### crons（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/crons` | `/open/crons` | `/api/crons` |
| GET | `/crons` | `/open/crons` | `/api/crons` |
| POST | `/crons` | `/open/crons` | `/api/crons` |
| PUT | `/crons` | `/open/crons` | `/api/crons` |
| GET | `/crons/:id` | `/open/crons/:id` | `/api/crons/:id` |
| GET | `/crons/:id/log` | `/open/crons/:id/log` | `/api/crons/:id/log` |
| GET | `/crons/:id/logs` | `/open/crons/:id/logs` | `/api/crons/:id/logs` |
| GET | `/crons/detail` | `/open/crons/detail` | `/api/crons/detail` |
| PUT | `/crons/disable` | `/open/crons/disable` | `/api/crons/disable` |
| PUT | `/crons/enable` | `/open/crons/enable` | `/api/crons/enable` |
| GET | `/crons/import` | `/open/crons/import` | `/api/crons/import` |
| DELETE | `/crons/labels` | `/open/crons/labels` | `/api/crons/labels` |
| POST | `/crons/labels` | `/open/crons/labels` | `/api/crons/labels` |
| PUT | `/crons/pin` | `/open/crons/pin` | `/api/crons/pin` |
| PUT | `/crons/run` | `/open/crons/run` | `/api/crons/run` |
| PUT | `/crons/status` | `/open/crons/status` | `/api/crons/status` |
| PUT | `/crons/stop` | `/open/crons/stop` | `/api/crons/stop` |
| PUT | `/crons/unpin` | `/open/crons/unpin` | `/api/crons/unpin` |
| DELETE | `/crons/views` | `/open/crons/views` | `/api/crons/views` |
| GET | `/crons/views` | `/open/crons/views` | `/api/crons/views` |
| POST | `/crons/views` | `/open/crons/views` | `/api/crons/views` |
| PUT | `/crons/views` | `/open/crons/views` | `/api/crons/views` |
| PUT | `/crons/views/disable` | `/open/crons/views/disable` | `/api/crons/views/disable` |
| PUT | `/crons/views/enable` | `/open/crons/views/enable` | `/api/crons/views/enable` |
| PUT | `/crons/views/move` | `/open/crons/views/move` | `/api/crons/views/move` |

来源文件: `back/api/cron.ts`

### envs（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/envs` | `/open/envs` | `/api/envs` |
| GET | `/envs` | `/open/envs` | `/api/envs` |
| POST | `/envs` | `/open/envs` | `/api/envs` |
| PUT | `/envs` | `/open/envs` | `/api/envs` |
| GET | `/envs/:id` | `/open/envs/:id` | `/api/envs/:id` |
| PUT | `/envs/:id/move` | `/open/envs/:id/move` | `/api/envs/:id/move` |
| PUT | `/envs/disable` | `/open/envs/disable` | `/api/envs/disable` |
| PUT | `/envs/enable` | `/open/envs/enable` | `/api/envs/enable` |
| PUT | `/envs/name` | `/open/envs/name` | `/api/envs/name` |
| PUT | `/envs/pin` | `/open/envs/pin` | `/api/envs/pin` |
| PUT | `/envs/unpin` | `/open/envs/unpin` | `/api/envs/unpin` |
| POST | `/envs/upload` | `/open/envs/upload` | `/api/envs/upload` |

来源文件: `back/api/env.ts`

### logs（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/logs` | `/open/logs` | `/api/logs` |
| GET | `/logs` | `/open/logs` | `/api/logs` |
| GET | `/logs/:file` | `/open/logs/:file` | `/api/logs/:file` |
| GET | `/logs/detail` | `/open/logs/detail` | `/api/logs/detail` |
| POST | `/logs/download` | `/open/logs/download` | `/api/logs/download` |

来源文件: `back/api/log.ts`

### scripts（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/scripts` | `/open/scripts` | `/api/scripts` |
| GET | `/scripts` | `/open/scripts` | `/api/scripts` |
| POST | `/scripts` | `/open/scripts` | `/api/scripts` |
| PUT | `/scripts` | `/open/scripts` | `/api/scripts` |
| GET | `/scripts/:file` | `/open/scripts/:file` | `/api/scripts/:file` |
| GET | `/scripts/detail` | `/open/scripts/detail` | `/api/scripts/detail` |
| POST | `/scripts/download` | `/open/scripts/download` | `/api/scripts/download` |
| PUT | `/scripts/rename` | `/open/scripts/rename` | `/api/scripts/rename` |
| PUT | `/scripts/run` | `/open/scripts/run` | `/api/scripts/run` |
| PUT | `/scripts/stop` | `/open/scripts/stop` | `/api/scripts/stop` |

来源文件: `back/api/script.ts`

### system（✅ OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| GET | `/system` | `/open/system` | `/api/system` |
| PUT | `/system/auth/reset` | `/open/system/auth/reset` | `/api/system/auth/reset` |
| PUT | `/system/command-run` | `/open/system/command-run` | `/api/system/command-run` |
| PUT | `/system/command-stop` | `/open/system/command-stop` | `/api/system/command-stop` |
| GET | `/system/config` | `/open/system/config` | `/api/system/config` |
| PUT | `/system/config/cron-concurrency` | `/open/system/config/cron-concurrency` | `/api/system/config/cron-concurrency` |
| PUT | `/system/config/dependence-clean` | `/open/system/config/dependence-clean` | `/api/system/config/dependence-clean` |
| PUT | `/system/config/dependence-proxy` | `/open/system/config/dependence-proxy` | `/api/system/config/dependence-proxy` |
| PUT | `/system/config/global-ssh-key` | `/open/system/config/global-ssh-key` | `/api/system/config/global-ssh-key` |
| PUT | `/system/config/linux-mirror` | `/open/system/config/linux-mirror` | `/api/system/config/linux-mirror` |
| PUT | `/system/config/log-remove-frequency` | `/open/system/config/log-remove-frequency` | `/api/system/config/log-remove-frequency` |
| PUT | `/system/config/node-mirror` | `/open/system/config/node-mirror` | `/api/system/config/node-mirror` |
| PUT | `/system/config/python-mirror` | `/open/system/config/python-mirror` | `/api/system/config/python-mirror` |
| PUT | `/system/config/timezone` | `/open/system/config/timezone` | `/api/system/config/timezone` |
| PUT | `/system/data/export` | `/open/system/data/export` | `/api/system/data/export` |
| PUT | `/system/data/import` | `/open/system/data/import` | `/api/system/data/import` |
| DELETE | `/system/log` | `/open/system/log` | `/api/system/log` |
| GET | `/system/log` | `/open/system/log` | `/api/system/log` |
| PUT | `/system/notify` | `/open/system/notify` | `/api/system/notify` |
| PUT | `/system/reload` | `/open/system/reload` | `/api/system/reload` |
| PUT | `/system/update` | `/open/system/update` | `/api/system/update` |
| PUT | `/system/update-check` | `/open/system/update-check` | `/api/system/update-check` |

来源文件: `back/api/system.ts`

### dependencies（⚠️ 通常不在 OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/dependencies` | `/open/dependencies` | `/api/dependencies` |
| GET | `/dependencies` | `/open/dependencies` | `/api/dependencies` |
| POST | `/dependencies` | `/open/dependencies` | `/api/dependencies` |
| PUT | `/dependencies` | `/open/dependencies` | `/api/dependencies` |
| GET | `/dependencies/:id` | `/open/dependencies/:id` | `/api/dependencies/:id` |
| PUT | `/dependencies/cancel` | `/open/dependencies/cancel` | `/api/dependencies/cancel` |
| DELETE | `/dependencies/force` | `/open/dependencies/force` | `/api/dependencies/force` |
| PUT | `/dependencies/reinstall` | `/open/dependencies/reinstall` | `/api/dependencies/reinstall` |

来源文件: `back/api/dependence.ts`

### root（⚠️ 通常不在 OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/apps` | `/open/apps` | `/api/apps` |
| GET | `/apps` | `/open/apps` | `/api/apps` |
| POST | `/apps` | `/open/apps` | `/api/apps` |
| PUT | `/apps` | `/open/apps` | `/api/apps` |
| PUT | `/apps/:id/reset-secret` | `/open/apps/:id/reset-secret` | `/api/apps/:id/reset-secret` |
| GET | `/auth/token` | `/open/auth/token` | `/api/auth/token` |
| GET | `/health` | `/open/health` | `/api/health` |

来源文件: `back/api/health.ts`

### subscriptions（⚠️ 通常不在 OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| DELETE | `/subscriptions` | `/open/subscriptions` | `/api/subscriptions` |
| GET | `/subscriptions` | `/open/subscriptions` | `/api/subscriptions` |
| POST | `/subscriptions` | `/open/subscriptions` | `/api/subscriptions` |
| PUT | `/subscriptions` | `/open/subscriptions` | `/api/subscriptions` |
| GET | `/subscriptions/:id` | `/open/subscriptions/:id` | `/api/subscriptions/:id` |
| GET | `/subscriptions/:id/log` | `/open/subscriptions/:id/log` | `/api/subscriptions/:id/log` |
| GET | `/subscriptions/:id/logs` | `/open/subscriptions/:id/logs` | `/api/subscriptions/:id/logs` |
| PUT | `/subscriptions/disable` | `/open/subscriptions/disable` | `/api/subscriptions/disable` |
| PUT | `/subscriptions/enable` | `/open/subscriptions/enable` | `/api/subscriptions/enable` |
| PUT | `/subscriptions/run` | `/open/subscriptions/run` | `/api/subscriptions/run` |
| PUT | `/subscriptions/status` | `/open/subscriptions/status` | `/api/subscriptions/status` |
| PUT | `/subscriptions/stop` | `/open/subscriptions/stop` | `/api/subscriptions/stop` |

来源文件: `back/api/subscription.ts`

### update（⚠️ 通常不在 OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| PUT | `/update/data` | `/open/update/data` | `/api/update/data` |
| PUT | `/update/reload` | `/open/update/reload` | `/api/update/reload` |
| PUT | `/update/system` | `/open/update/system` | `/api/update/system` |

来源文件: `back/api/update.ts`

### user（⚠️ 通常不在 OpenAPI scope）

| 方法 | 资源路径 | `/open` 访问示例 | `/api` 访问示例 |
|---|---|---|---|
| GET | `/user` | `/open/user` | `/api/user` |
| PUT | `/user` | `/open/user` | `/api/user` |
| PUT | `/user/avatar` | `/open/user/avatar` | `/api/user/avatar` |
| PUT | `/user/init` | `/open/user/init` | `/api/user/init` |
| POST | `/user/login` | `/open/user/login` | `/api/user/login` |
| GET | `/user/login-log` | `/open/user/login-log` | `/api/user/login-log` |
| POST | `/user/logout` | `/open/user/logout` | `/api/user/logout` |
| GET | `/user/notification` | `/open/user/notification` | `/api/user/notification` |
| PUT | `/user/notification` | `/open/user/notification` | `/api/user/notification` |
| PUT | `/user/notification/init` | `/open/user/notification/init` | `/api/user/notification/init` |
| PUT | `/user/two-factor/active` | `/open/user/two-factor/active` | `/api/user/two-factor/active` |
| PUT | `/user/two-factor/deactive` | `/open/user/two-factor/deactive` | `/api/user/two-factor/deactive` |
| GET | `/user/two-factor/init` | `/open/user/two-factor/init` | `/api/user/two-factor/init` |
| PUT | `/user/two-factor/login` | `/open/user/two-factor/login` | `/api/user/two-factor/login` |

来源文件: `back/api/user.ts`
