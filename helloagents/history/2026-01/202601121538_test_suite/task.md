# 任务清单: 全功能测试用例补齐 + Claude 复审

目录: `helloagents/history/2026-01/202601121538_test_suite/`

---

## 1. 覆盖缺口梳理
- [√] 1.1 识别缺失测试：config 校验、core validation、Unraid/Qinglong Provider 交互流、回调边界

## 2. 单元测试补齐（详细）
- [√] 2.1 `internal/config/config_test.go`：Duration 解析、默认值、GraphQL identifier/type 校验、Unraid 配置覆盖行为
- [√] 2.2 `internal/core/validation_test.go`：容器名校验边界
- [√] 2.3 `internal/unraid/provider_test.go`：菜单/动作选择→输入→回显（含日志行数默认/输入/上限）
- [√] 2.4 `internal/qinglong/provider_test.go`：多实例选择、动作菜单、搜索/按ID、任务选择、确认执行
- [√] 2.5 `internal/wecom/callback_test.go`：补充更多边界（签名错误、无 MsgId/TaskId 的去重、core 返回 error 行为）

## 3. 验证
- [√] 3.1 gofmt
- [√] 3.2 `go test ./...`

## 4. Claude 复审
- [√] 4.1 使用 `claude -p` 审查测试覆盖盲区
- [√] 4.2 根据建议补齐/调整测试场景

## 5. 知识库与迁移
- [√] 5.1 更新 `helloagents/CHANGELOG.md`
- [√] 5.2 迁移方案包至 `helloagents/history/`
