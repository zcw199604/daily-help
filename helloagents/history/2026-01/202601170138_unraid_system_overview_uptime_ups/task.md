# 轻量迭代任务清单：unraid_system_overview_uptime_ups

目标：在“系统资源概览”中补充 Unraid 运行时长（服务启动时长）与 UPS 信息，便于一眼判断整体状态。

## 任务

- [√] unraid：GetSystemMetrics 增加 Unraid uptime 获取（`info { os { uptime } }`），并在概览输出展示
- [√] unraid：GetSystemMetrics 增加 UPS 信息获取（`upsDevices { ... }`），并在概览输出展示
- [√] 更新知识库与 Changelog：记录新增字段与展示口径
- [√] 验证：Docker 编译通过
- [√] 验证：单元测试通过（go test ./...）
- [√] 迁移方案包至 history/ 并更新 history/index.md
- [√] 交付：git commit（中文）并 push
