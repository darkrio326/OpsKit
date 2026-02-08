# OpsKit v0.3.7-preview.1

> 预览版本（草稿）。主线聚焦“模板接入前置增强”，在 `v0.3.6` 通用能力正式版之上推进 Milestone 4。

## 计划重点

- 模板变量与 schema 约束增强
- 模板错误定位可读性提升（路径化 + 修复建议）
- 模板文档与验收口径收敛（中文/英文一致）

## 计划变更

### Added

- 新增 `docs/RELEASE_PLAN_v0.3.7-preview.1.md`
- 模板验收示例命令与最小变量集说明（按 demo 模板补齐）
- 新增 `opskit template validate --json`（机器可读模板校验输出）
- 模板校验错误新增统一结构：`path/code/message/advice`
- 新增 `docs/specs/SPEC_TEMPLATE_VALIDATE_JSON.md`（模板校验 JSON 契约）
- 新增 `scripts/template-validate-check.sh`（CI 脚本化验收）

### Changed

- `template validate` 错误提示结构化增强（路径化 + 修复建议）
- `template validate` 类型错误增强（array/object 提供 JSON 示例和解析原因）
- `docs/specs/SPEC_TEST_ACCEPTANCE.md` 门禁双轨策略用于模板接入前置验收
- demo 模板 README 增加 `template validate --json` 用法和常见错误码指引

### Out of Scope

- 生产级中间件模板发布
- 客户环境适配与多节点能力

## 计划验证命令

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --clean
go run ./cmd/opskit template validate assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json
go run ./cmd/opskit template validate assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.env
```
