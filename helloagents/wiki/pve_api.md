# Proxmox VE（PVE）API 要点整理

## 来源与检索时间
本页为“要点整理”，权威信息以官方文档为准（检索时间：2026-01-17）：
- Proxmox Wiki：`Proxmox VE API`
  - https://pve.proxmox.com/wiki/Proxmox_VE_API
- Proxmox VE 文档索引（含 API Viewer）
  - https://pve.proxmox.com/pve-docs/
- API Viewer（可交互浏览所有接口与参数）
  - https://pve.proxmox.com/pve-docs/api-viewer/index.html

## API 概述
- PVE 提供 REST-like API，主数据格式为 JSON。
- API 使用 JSON Schema 形式化描述，可用于自动生成文档与参数校验。
- 具体接口清单与参数以 API Viewer 为准（适合查接口路径、权限、参数类型、返回结构）。

## API 稳定性（兼容性预期）
- 通常会在**同一主版本**内尽量保持兼容（例如 6.x 内部相对稳定，但升级到 7.x 可能存在不兼容）。
- 典型破坏性变更：移除接口、迁移路径、移除参数、将返回类型从非空改为其他类型等。

## Base URL 与返回格式
- API 基于 HTTPS，默认端口 `8006`。
- 常用 Base URL（JSON）：
  - `https://<PVE_HOST>:8006/api2/json/`
- 也可通过 URL 中的格式段切换返回格式（调试用途更常见）：
  - `json` / `extjs` / `html` / `text`

## 鉴权方式
PVE API 常用两类鉴权方式：Ticket Cookie（浏览器/会话）与 API Token（服务端/无状态）。

### 1) Ticket Cookie + CSRF（常用于浏览器会话）
- 通过 `POST /access/ticket` 获取：
  - `ticket`（用于 Cookie：`PVEAuthCookie`）
  - `CSRFPreventionToken`（写操作必须带）
- Ticket 默认有效期约 2 小时；可在过期前用旧 ticket 续期（细节以官方说明为准）。

示例（仅演示结构，使用占位符）：
```bash
# 获取 ticket + csrf（注意：生产/多人系统避免在命令行明文传密码）
curl -k -d 'username=root@pam' --data-urlencode 'password=<PASSWORD>' \
  https://<PVE_HOST>:8006/api2/json/access/ticket

# 读请求：带 Cookie
curl -k -b \"PVEAuthCookie=<TICKET>\" \
  https://<PVE_HOST>:8006/api2/json/

# 写请求：带 Cookie + CSRFPreventionToken Header
curl -k -X POST -b \"PVEAuthCookie=<TICKET>\" -H \"CSRFPreventionToken: <CSRF>\" \
  https://<PVE_HOST>:8006/api2/json/<PATH>
```

### 2) API Token（推荐用于服务端集成）
- API Token 适合“无状态访问”，可单独授予权限并设置过期时间，泄露后可单独吊销。
- 使用 `Authorization` Header，格式如下（以官方为准）：
  - `Authorization: PVEAPIToken=USER@REALM!TOKENID=UUID`
- 使用 API Token 进行 POST/PUT/DELETE 通常不需要 CSRF（不在浏览器上下文，CSRF 攻击面不成立）。

示例：
```bash
curl -k -H \"Authorization: PVEAPIToken=root@pam!monitoring=<UUID>\" \
  https://<PVE_HOST>:8006/api2/json/
```

### 安全提示（强烈建议）
- 避免把密码/token 直接写在命令行参数里（会被其他用户通过进程列表读取）；可将 Header 写入仅自己可读的文件再用 `curl -H @file.headers` 引用。

## API Viewer（查接口的首选入口）
- 线上文档的 API Viewer：
  - https://pve.proxmox.com/pve-docs/api-viewer/index.html
- 在 PVE 节点上通常也能访问本地 API Viewer（用于对照具体版本的接口差异）：
  - `https://<PVE_HOST>:8006/pve-docs/api-viewer/index.html`
- API Viewer 的 schema 数据可用于自动化（例如从 `apidoc.js` 提取路径/参数/权限信息），但不建议直接把大体量 schema 文件快照提交进仓库。

## 命令行工具：pvesh
- `pvesh` 是 PVE 自带的“REST API 命令行入口”，在节点上以 root 执行时可代理到集群其他节点（底层机制以官方说明为准）。
- 常用示例（路径以实际 API Viewer 为准）：
```bash
pvesh get /version
pvesh get /access/users
pvesh create /access/users --userid testuser@pve
pvesh delete /access/users/testuser@pve
```

## 客户端库（部分摘录）
- 官方维护（Perl）：https://git.proxmox.com/?p=pve-apiclient.git;a=summary
- 社区生态（示例）：
  - Python：proxmoxer（https://pypi.python.org/pypi/proxmoxer）
  - Go：Telmate/proxmox-api-go、luthermonson/go-proxmox
  - Terraform：Telmate/terraform-provider-proxmox

## 对 wecom-home-ops 的落地建议
- 如后续要新增 PVE Provider，建议优先使用 **API Token**（最小权限、可吊销、避免保存账号密码与处理 CSRF）。
- 具体接口路径、权限要求与参数结构，请以目标 PVE 版本的 **API Viewer** 为准，并在知识库中补充“目标环境差异快照”（参考 `unraid_schema_*.md` 的做法）。

