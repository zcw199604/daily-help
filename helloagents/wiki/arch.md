# 架构设计

## 总体架构

```mermaid
flowchart TD
    U[用户（企业微信应用会话）] -->|消息/事件回调| W[企业微信服务端]
    W -->|回调URL| S[daily-help 服务]
    S --> C[连接器层（适配器）]
    C -->|MVP: Unraid| R[Unraid]
    R -->|容器管理| D[Docker]
```

## 技术栈
- **后端:** Go（`net/http`）
- **数据:** MVP 默认内存会话状态 + 文件日志（可选扩展 SQLite）
- **部署:** 本地服务器/NAS/Docker（需公网 HTTPS 或可被企业微信访问的回调地址）
 - **Unraid:** Unraid Connect 插件 GraphQL API（`/graphql` + `x-api-key`）

## 核心流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant WeCom as 企业微信
    participant Svc as daily-help
    participant Unraid as Unraid

    User->>WeCom: 发送“容器”/点击入口
    WeCom->>Svc: 回调消息（加密）
    Svc->>WeCom: 回复交互卡片（重启/停止/强制更新）
    User->>WeCom: 点击按钮并按提示输入容器名
    WeCom->>Svc: 回调按钮事件/文本消息
    Svc->>Unraid: 执行容器操作
    Svc->>WeCom: 回显结果与耗时
```

## 重大架构决策
完整的 ADR 存储在各变更的 how.md 中，本章节提供索引。

| adr_id | title | date | status | affected_modules | details |
|--------|-------|------|--------|------------------|---------|
| ADR-001 | 单体服务 + 适配器插件化 | 2026-01-12 | ✅已采纳 | core,wecom,unraid | [how.md](../history/2026-01/202601120816_wecom_unraid/how.md) |
| ADR-002 | MVP 默认“内存会话状态 + 日志审计” | 2026-01-12 | ✅已采纳 | core | [how.md](../history/2026-01/202601120816_wecom_unraid/how.md) |
