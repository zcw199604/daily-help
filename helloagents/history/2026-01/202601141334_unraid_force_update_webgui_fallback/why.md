# 变更提案: Unraid 容器强制更新 WebGUI 兜底

## 需求背景
当前 wecom-home-ops 的 Unraid “强制更新”通过 GraphQL 调用 `docker { updateContainer(...) }`（并回退尝试 `update(...)`）实现。在部分 Unraid 版本/实现中，`DockerMutations` 并不暴露容器更新相关 mutation，导致出现 GraphQL 400 校验错误，用户侧表现为“执行容器更新失败”。

同时，抓包可见 Unraid WebGUI 在页面更新容器时会调用 `webGui/include/StartCommand.php` 并执行 `update_container <name>` 命令（依赖 csrf_token/可能依赖登录 Cookie）。

## 变更内容
1. 保持 GraphQL 作为首选路径：继续优先使用可配置的强制更新 mutation（含回退候选）。
2. 当 GraphQL 明确提示“不支持字段/参数/类型”时，为强制更新新增 WebGUI StartCommand.php 兜底能力。
3. 增加配置项（可选启用）：`unraid.webgui_csrf_token` / `unraid.webgui_cookie` / `unraid.webgui_command_url`。

## 影响范围
- **模块:** unraid / config / app
- **文件:**
  - `internal/unraid/client.go`
  - `internal/config/config.go`
  - `internal/app/server.go`
  - `config.example.yaml`
  - `internal/unraid/client_test.go`
- **API:** 无对外 HTTP API 变更（仅内部能力增强）
- **数据:** 无

## 核心场景

### 需求: 强制更新容器
**模块:** unraid
在用户触发“强制更新”动作时，应尽可能完成容器更新：
- GraphQL 支持更新 mutation → 使用 GraphQL 更新成功
- GraphQL 不支持更新 mutation（Cannot query field/Unknown argument/Unknown type）→ 使用 WebGUI StartCommand.php 兜底更新（需要配置 csrf_token/可选 Cookie）

#### 场景: GraphQL 不支持 updateContainer
GraphQL 返回类似 `Cannot query field "updateContainer" on type "DockerMutations"`。
- 预期结果：自动尝试 WebGUI 兜底；若未配置兜底参数，则错误信息提示需要配置的 key

## 风险评估
- **风险:** WebGUI StartCommand.php 属于内部接口，可能受版本变化影响；且在启用 WebGUI 登录保护时需要 Cookie/CSRF，配置不当会导致兜底失败。
- **缓解:**
  - 仅在 GraphQL 明确“不支持”时才启用兜底，避免重复触发真实更新操作
  - 配置项全部可选，不影响现有 GraphQL 正常路径
  - 对 WebGUI 返回做基本错误识别（HTTP 非 2xx / CSRF 无效 / 登录页面）
