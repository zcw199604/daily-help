# 任务清单: Unraid 10.10.10.100 GraphQL Schema 入库

目录: `helloagents/plan/202601141432_unraid_schema_snapshot/`

---

## 1. 知识库整理
- [√] 1.1 请求 `http://10.10.10.100/graphql` 并通过 introspection 获取 schema，整理 Query/Mutation/Subscription 与关键模块字段清单
- [√] 1.2 写入知识库：新增 schema 摘要文档，并在 unraid 模块/官方 API 文档中补充目标实例差异

## 2. 安全检查
- [√] 2.1 确认文档不包含敏感信息（API Key/Cookie/CSRF token 不落库）
