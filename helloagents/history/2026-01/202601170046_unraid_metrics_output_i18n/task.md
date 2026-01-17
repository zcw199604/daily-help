# 轻量迭代任务清单：unraid_metrics_output_i18n

目标：将 Unraid “系统资源概览/详情”输出改为更易理解的中文模板，便于快速判断资源情况。

## 任务

- [√] 调整系统资源概览输出：中文标签 + 增加关键字段（可用/实际占用/网络等）
- [√] 调整系统资源详情输出：中文标签 + 先给出关键结论，再给细节（CPU 核心/内存 raw vs effective）
- [√] 更新知识库/Changelog：记录输出格式变更（用户可见）
- [√] 验证：Docker 编译通过
- [√] 验证：单元测试通过（go test ./...）
- [√] 迁移方案包至 history/ 并更新 history/index.md
- [√] 交付：git commit（中文）并 push
