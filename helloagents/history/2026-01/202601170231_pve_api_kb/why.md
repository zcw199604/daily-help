# 变更提案: PVE API 文档入库

## 需求背景
wecom-home-ops 当前已沉淀 Unraid/青龙等服务的接入资料，但 PVE（Proxmox VE）作为常见虚拟化平台，后续可能纳入“多服务 Provider 框架”的扩展范围。为了降低后续接入成本并统一文档入口，需要先把 PVE 的官方 API 资料要点整理进本地知识库。

## 变更内容
1. 新增 PVE REST API 文档要点（鉴权、Base URL、常用工具与文档入口）
2. 更新知识库索引，补充快速链接
3. 更新变更记录，便于追溯

## 影响范围
- **模块:** 知识库（wiki）
- **文件:** `helloagents/wiki/*`、`helloagents/CHANGELOG.md`
- **API:** 无（仅文档整理）
- **数据:** 无

## 核心场景

### 需求: PVE API 文档入库
**模块:** 知识库（wiki）
将 PVE 官方 API 文档的关键点整理为中文要点，包含来源链接与检索时间，后续接入 PVE Provider 时可直接引用。

#### 场景: 查阅鉴权与调用示例
在编写集成或排障时，能快速确认：
- Base URL 与返回格式（json/extjs 等）
- Ticket Cookie + CSRF 与 API Token 的差异
- 通过 API Viewer / pvesh 定位与验证具体接口

## 风险评估
- **风险:** 直接复制上游大量原文导致版权/维护成本问题
- **缓解:** 采用“要点整理 + 引用官方来源链接”的方式，并记录检索时间；避免存入大体量 schema 原始文件

