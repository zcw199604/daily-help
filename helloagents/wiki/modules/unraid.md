# unraid

## 目的
封装 Unraid 的容器管理能力，对外提供统一动作接口（重启/停止/强制更新）。

## 模块概述
- **职责:** 连接管理（如 SSH/HTTP）；执行容器操作；错误归类与回显信息格式化
- **状态:** 🚧开发中
- **最后更新:** 2026-01-12

## 规范

### 需求: 容器操作（MVP）
**模块:** unraid
对指定容器执行：
- 重启（restart）
- 停止（stop）
- 强制更新（force update）
  - 说明：强制更新依赖 Unraid GraphQL 提供相应 mutation；如名称不在默认探测列表，可通过配置 `unraid.force_update_mutation` 指定

### 需求: 连接方式可替换
**模块:** unraid
MVP 使用 Unraid Connect 插件提供的 GraphQL API（`/graphql` + `x-api-key`），并在实现层抽象“客户端/执行器”接口，允许后续在不改业务的情况下切换：
- 其他 API 形态（不同插件/版本差异）
- CLI 模块（如 `unraid-api` 提供可用命令能力）
- SSH + Docker CLI（仅作为备选）

## API接口
本模块不直接对外提供 HTTP API，通过内部接口供 core 调用。

## 数据模型
无；仅使用 Config 中的连接参数。

## 依赖
- core（接口约定）

## 变更历史
- 2026-01-12: 基于 GraphQL API 实现容器 stop/start/restart，并通过 introspection 探测“强制更新”能力
