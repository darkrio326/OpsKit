# OpsKit v0.4.0-preview.2

> 预览版本（收口中）。主线聚焦 Milestone 4：模板能力增强与门禁契约收敛。

## 重点

- 每个 preview 固定 5 个功能点推进，本版完成 5/5
- 通用 `generic-manage-v1` 模板已接入默认门禁链路
- 模板校验与脚本门禁口径统一（含正负例）

## 主要变更

### Added

- 新增发布计划：`docs/RELEASE_PLAN_v0.4.0-preview.2.md`
- 新增 demo 模板：`assets/templates/demo-runtime-baseline.json`
- 新增 vars 示例：`examples/vars/demo-runtime-baseline.json`
- 新增 vars 示例：`examples/vars/demo-runtime-baseline.env`

### Changed

- `template.vars` 增强：新增 `group` 元数据并校验格式
- `template validate` 增强：支持模板路径前后均可传 `--vars/--vars-file/--json`
- `scripts/template-validate-check.sh` 新增变量类型错误负例断言
- `scripts/template-validate-check.sh` 默认新增 `generic-manage-v1` 正向校验
- `scripts/release-check.sh` 默认新增 `generic-manage-v1` 模板校验步骤
- `scripts/release-check.sh` dry-run 增加 `generic-manage-v1` 的 A/D 计划校验
- `docs/specs/SPEC_TEMPLATE_VALIDATE_JSON.md` 补齐错误码清单
- `docs/specs/SPEC_CORE.md` 补齐模板变量字段约束

### Out of Scope

- 生产级中间件部署模板
- 客户环境定制逻辑
- 多节点编排

## 验证命令

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/template-validate-check.sh --clean
scripts/release-check.sh --skip-tests --skip-run
scripts/release-check-json-contract.sh --clean
```
