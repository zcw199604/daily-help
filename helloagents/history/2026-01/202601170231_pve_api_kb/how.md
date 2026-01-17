# 技术设计: PVE API 文档入库

## 技术方案

### 核心技术
- 知识库文档采用 Markdown（`helloagents/wiki/`）
- 来源以 Proxmox 官方站点为准（wiki + pve-docs API Viewer）

### 实现要点
- 汇总 PVE API 的“调用必备信息”：Base URL、数据格式、鉴权方式、常见示例与工具链
- 统一引用入口：在 `helloagents/wiki/overview.md` 增加快速链接
- 控制内容粒度：不在仓库内保存大体量的 `apidoc.js`/schema 快照，避免仓库膨胀；需要全量接口时引导到官方 API Viewer

## 安全与性能
- **安全:** 文档中不写入任何真实地址、用户名、密码、token；示例使用占位符；提醒避免在命令行明文传递凭据
- **性能:** 仅文档变更，无运行时性能影响

## 测试与部署
- **测试:** 文档变更不涉及运行逻辑；执行 `go test ./...` 作为回归检查（如可用）
- **部署:** 无需部署变更；提交并 push 即可

