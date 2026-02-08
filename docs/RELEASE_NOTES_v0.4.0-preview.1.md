# OpsKit v0.4.0-preview.1

> 预览版本（规划中）。主线聚焦 Milestone 4：模板能力增强与模板库推进。

## 计划重点

- 模板变量模型增强（约束、描述、错误定位）
- 去生产化 demo 模板扩展
- 模板门禁与发布门禁链路收敛

## 计划变更

### Added

- 新增 `docs/RELEASE_PLAN_v0.4.0-preview.1.md`
- 计划新增更多去生产化模板与模板示例

### Changed

- 计划增强模板校验错误码与可读性
- 计划统一模板文档与门禁验收口径

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
