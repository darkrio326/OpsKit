# OpsKit v0.4.0-preview.1（归档）

> 本版本内容已并入 `docs/RELEASE_NOTES_v0.4.0-preview.2.md`，本文件仅保留历史上下文。

## 计划重点

- 模板变量模型增强（约束、描述、错误定位）
- 去生产化 demo 模板扩展
- 模板门禁与发布门禁链路收敛

## 计划变更

### Added

- 新增 `docs/RELEASE_PLAN_v0.4.0-preview.1.md`
- 规划固定“每个 preview 5 个功能点”的推进节奏
- 新增去生产化模板：`assets/templates/demo-runtime-baseline.json`
- 新增模板说明：`assets/templates/demo-runtime-baseline.README.md`
- 新增 vars 示例：`examples/vars/demo-runtime-baseline.json`、`examples/vars/demo-runtime-baseline.env`

### Changed

- `template.vars` 新增 `group` 元数据，并纳入 `template validate` 格式校验
- demo 模板变量补充 `group` 标注（`service/paths/runtime/network`）
- `scripts/template-validate-check.sh` 新增变量类型错误负例断言
- `scripts/release-check.sh` 默认新增 `demo-runtime-baseline` 模板校验步骤
- `docs/specs/SPEC_TEMPLATE_VALIDATE_JSON.md` 补齐错误码清单
- `docs/specs/SPEC_CORE.md` 补充模板变量字段与 `group` 命名约束
- 功能点完成度：1/2/3/4 已完成，功能点 5 进行中

### Out of Scope

- 生产级中间件部署模板
- 客户环境适配
- 多节点编排能力

## 计划验证命令

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/template-validate-check.sh --clean
scripts/release-check-json-contract.sh --clean
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --with-release-json-contract --clean
```
