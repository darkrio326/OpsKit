# OpsKit v0.4.0-preview.4（草稿）

> 预览版本（进行中）。主线聚焦 Milestone 4：模板可发现性与模板契约门禁。

## 当前已完成

### Added

- 新增命令：`opskit template list`
- 新增命令：`opskit template list --json`
- 新增模板目录能力：内置模板 `ref/source/id/name/mode/aliases` 自动输出
- 新增规范：`docs/specs/SPEC_TEMPLATE_LIST_JSON.md`
- 新增计划：`docs/RELEASE_PLAN_v0.4.0-preview.4.md`
- 新增模板：`assets/templates/demo-generic-selfhost-deploy.json`
- 新增模板：`assets/templates/demo-minio-deploy.json`
- 新增模板：`assets/templates/demo-elk-deploy.json`
- 新增模板：`assets/templates/demo-powerjob-deploy.json`
- 新增示例变量：`examples/vars/demo-generic-selfhost-deploy.json` / `examples/vars/demo-generic-selfhost-deploy.env`
- 新增示例变量：`examples/vars/demo-minio-deploy.json` / `examples/vars/demo-minio-deploy.env`
- 新增示例变量：`examples/vars/demo-elk-deploy.json` / `examples/vars/demo-elk-deploy.env`
- 新增示例变量：`examples/vars/demo-powerjob-deploy.json` / `examples/vars/demo-powerjob-deploy.env`

### Changed

- `cmd template` 子命令入口调整为：`validate|list`
- `scripts/release-check.sh` 新增 `template list --json` 冒烟步骤
- `printUsage` 增加 `opskit template list [--json]`
- `demo-blackbox-middleware-manage` README 新增 default/FCS/Kingdee 对照矩阵与最小变量集
- 明确模板双模式边界：黑箱中间件走 `manage`，自行部署服务走 `deploy`
- `scripts/template-validate-check.sh` 新增 `demo-generic-selfhost-deploy`/`demo-minio-deploy`/`demo-elk-deploy`/`demo-powerjob-deploy` 校验步骤
- `demo-elk-deploy` 增加高级变量：JVM（ES/Logstash）、pipeline 路径、Kibana TLS 配置
- `opskit template list` 增加模板分类字段：`serviceScope` 与 `tags`
- `opskit template list` 支持输出本地 demo 模板（若存在 `assets/templates/*.json`）
- `template validate` 支持直接使用 demo 模板引用名（如 `demo-elk-deploy`）
- UI 新增 Templates 区块，按 `mode/serviceScope/tags` 分组与筛选展示
- `opskit status` 额外写出 `state/templates.json` 供 UI 直接读取

### Tests

```bash
go test ./cmd/opskit ./internal/templates
scripts/release-check.sh --skip-tests --skip-run
```

## 进行中

- 模板接入辅助脚本细化（命名规范/目录结构/模板说明完整性）
- 基于通用 deploy 基线继续扩展业务模板库

## Out of Scope

- 生产级中间件部署模板
- 客户环境定制逻辑
- 多节点编排
