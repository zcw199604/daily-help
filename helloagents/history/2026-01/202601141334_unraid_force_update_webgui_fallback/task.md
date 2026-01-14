# 任务清单: Unraid 容器强制更新 WebGUI 兜底

目录: `helloagents/plan/202601141334_unraid_force_update_webgui_fallback/`

---

## 1. unraid 模块
- [√] 1.1 在 `internal/unraid/client.go` 增加 WebGUI StartCommand 兜底（配置项/派生 URL/POST form），验证 why.md#场景-graphql-不支持-updatecontainer
- [√] 1.2 在 `internal/config/config.go` / `internal/app/server.go` / `config.example.yaml` 增加 webgui 配置项并接入 ClientConfig，验证 why.md#需求-强制更新容器
- [√] 1.3 在 `internal/unraid/client_test.go` 增加单测覆盖“GraphQL 不支持 → WebGUI 兜底成功”，验证 why.md#场景-graphql-不支持-updatecontainer

## 2. 安全检查
- [√] 2.1 执行安全检查（敏感信息不落库/不打印；兜底触发条件避免重复更新；仅在校验错误时回退）

## 3. 文档更新
- [√] 3.1 更新 `helloagents/wiki/modules/unraid.md`：补充 WebGUI 兜底与配置项说明
- [√] 3.2 更新 `helloagents/wiki/unraid_mobile_ui.md`：明确其不包含容器更新能力，StartCommand.php 属于 WebGUI 内部接口
- [√] 3.3 更新 `helloagents/wiki/unraid_official_api.md`：补充“官方文档未列出容器更新 mutation，建议以 live schema 为准/或使用 WebGUI 兜底”
- [√] 3.4 更新 `helloagents/CHANGELOG.md`

## 4. 测试
- [√] 4.1 通过 Docker 运行 `go test ./...` 验证单测通过
