# 任务清单: 企业微信回调解密 PKCS7 对齐修复（blockSize=32）

目录: `helloagents/plan/202601131307_wecom_pkcs7_padding_fix/`

---

## 1. 修复加解密实现与官方一致
- [√] 1.1 将 PKCS7 padding 的 blockSize 调整为 32（修复 invalid pkcs7 padding）
- [√] 1.2 确保 Encrypt/Decrypt 均使用同一 padding 规则

## 2. 补齐测试覆盖
- [√] 2.1 增加单测覆盖 padding > 16 的场景，防止回归

## 3. 同步知识库与验证
- [√] 3.1 更新 `helloagents/wiki/modules/wecom.md` 与 `helloagents/CHANGELOG.md`
- [√] 3.2 `go test ./...` 验证通过
