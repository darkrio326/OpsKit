# OpsKit v0.4.0-preview.3

> 预览版本（补齐发布）。主线仍聚焦 Milestone 4：模板能力与发布门禁收敛。

## 重点

- 补齐 `v0.4.0-preview.2` 之后未进入已发布 tag 的改动
- 黑箱中间件模板采用“模板名通用 + vars 预设区分”策略
- `docs/` 根目录仅保留当前版本发布文档，历史版统一归档

## 主要变更

### Added

- 新增发布计划：`docs/RELEASE_PLAN_v0.4.0-preview.3.md`
- 新增模板：`assets/templates/demo-blackbox-middleware-manage.json`
- 新增模板说明：`assets/templates/demo-blackbox-middleware-manage.README.md`
- 新增 vars 示例：`examples/vars/demo-blackbox-middleware-manage.json`
- 新增 vars 示例：`examples/vars/demo-blackbox-middleware-manage.env`
- 新增 vars 预设：`examples/vars/demo-blackbox-middleware-manage-fcs.json`
- 新增 vars 预设：`examples/vars/demo-blackbox-middleware-manage-fcs.env`
- 新增 vars 预设：`examples/vars/demo-blackbox-middleware-manage-kingdee.json`
- 新增 vars 预设：`examples/vars/demo-blackbox-middleware-manage-kingdee.env`
- 新增发布文档归档索引：`docs/releases/README.md`

### Changed

- `scripts/template-validate-check.sh` 默认校验黑箱模板三套 vars（default/FCS/Kingdee）
- `scripts/release-check.sh` 默认校验与 dry-run 覆盖黑箱模板三套 vars
- 黑箱模板命名策略统一：模板文件不带厂商后缀，厂商差异只放 vars 文件
- `docs/` 根目录仅保留当前版本发布文档：
  - `docs/RELEASE_NOTES_v0.4.0-preview.3.md`
  - `docs/RELEASE_PLAN_v0.4.0-preview.3.md`
- 历史发布文档归档到：
  - `docs/releases/notes/`
  - `docs/releases/plans/`

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
