# 数据模型

## 概述
MVP 以“个人使用 + 低依赖”为目标，优先采用内存会话状态与结构化日志；后续如需多设备/重启不丢状态，可引入 SQLite 做轻量持久化。

---

## 关键数据对象

### 配置（Config）
- **用途:** 存储企业微信与 Unraid 的连接信息、权限白名单等
- **来源:** 环境变量或 YAML 配置文件
- **敏感字段:** 企业微信 Secret、回调 Token、EncodingAESKey、Unraid 凭据（必须避免写入日志）

### 会话状态（Conversation State）
- **用途:** 支撑“按钮选择动作 → 提示用户输入参数 → 二次确认 → 执行”的多步交互
- **主键:** `wecom_userid`
- **建议字段:**
  - `pending_action`（restart/stop/force_update）
  - `pending_target`（container_name）
  - `expires_at`（TTL 超时）

### 审计事件（Audit Event）
- **用途:** 记录关键操作请求与执行结果，便于追溯
- **存储:** 结构化日志（可选落地为 SQLite）
- **关键字段:** `wecom_userid`, `action`, `target`, `request_id`, `result`, `error`, `duration_ms`

