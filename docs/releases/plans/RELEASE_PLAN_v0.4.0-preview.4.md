# OpsKit v0.4.0-preview.4 发布任务清单

> 目标：在 M4 阶段继续增强模板可发现性与模板契约门禁，降低模板接入成本。

## 1. 版本目标

- 提供模板目录能力（CLI 可直接列出内置模板）
- 补齐 `template list --json` 契约与门禁接入
- 固化模板双模式边界：黑箱 `manage` 与自行部署 `deploy`
- 继续沉淀黑箱 Manage 模板的接入规范，为后续自行部署模板铺路

## 2. 范围边界

### In Scope

- 新命令：`opskit template list`（文本 + JSON）
- 模板目录数据从内置模板自动提取（ref/source/id/name/mode/aliases）
- 发布门禁增加 `template list --json` 冒烟步骤
- 规格文档补齐 `SPEC_TEMPLATE_LIST_JSON.md`
- 版本入口滚动到 preview.4 草稿

### Out of Scope

- 生产级中间件部署模板
- 客户环境脚本与定制参数
- 多节点管理能力

## 3. 功能点拆分（固定 5 项）

### 功能点 1（完成）：模板目录命令

- 新增 `opskit template list`
- 新增 `opskit template list --json`
- `template` 子命令入口支持 `validate|list`

验收：

```bash
go run ./cmd/opskit template list
go run ./cmd/opskit template list --json
```

### 功能点 2（完成）：模板目录契约

- 新增 `docs/specs/SPEC_TEMPLATE_LIST_JSON.md`
- 约束 `command/schemaVersion/count/templates[]`

验收：

```bash
rg -n "SPEC_TEMPLATE_LIST_JSON|template list --json" docs/specs
```

### 功能点 3（完成）：门禁接入模板目录冒烟

- `scripts/release-check.sh` 增加 `template list --json` 步骤
- 失败 reason code：`step_failed_template_list_json`

验收：

```bash
scripts/release-check.sh --skip-tests --skip-run
```

### 功能点 4（完成）：黑箱模板接入说明矩阵

- 在模板 README 中增加 default/FCS/Kingdee 对照矩阵
- 明确变量命名与最小必填集

验收：

```bash
rg -n "对照矩阵|最小变量集|default|FCS|Kingdee" assets/templates/demo-blackbox-middleware-manage.README.md
```

### 功能点 5（完成）：模板接入辅助脚本 + 自行部署模板扩展

- 增加模板自检脚本（检查模板命名、vars 示例齐全性）
- 输出 machine-readable summary 供 CI 使用
- 输出模式分类提示（`manage`/`deploy`）用于模板接入门禁
- 新增通用 deploy 基线：`demo-generic-selfhost-deploy`
- 新增自行部署模板：`demo-minio-deploy`、`demo-elk-deploy`、`demo-powerjob-deploy`
- `demo-elk-deploy` 增加高级变量（JVM/pipeline/TLS）以覆盖多服务模板细化场景
- `template list` 输出增加分类标签（mode + serviceScope + tags）
- `template list` 合并展示 builtin + 本地 demo 模板目录
- demo 模板 `ref` 可直接用于 `template validate`
- UI 增加模板分组视图（读取 `state/templates.json`）

验收：

```bash
scripts/template-validate-check.sh --clean
go run ./cmd/opskit template validate --vars-file examples/vars/demo-generic-selfhost-deploy.json assets/templates/demo-generic-selfhost-deploy.json
go run ./cmd/opskit template validate --vars-file examples/vars/demo-minio-deploy.json assets/templates/demo-minio-deploy.json
go run ./cmd/opskit template validate --vars-file examples/vars/demo-elk-deploy.json assets/templates/demo-elk-deploy.json
go run ./cmd/opskit template validate --vars-file examples/vars/demo-powerjob-deploy.json assets/templates/demo-powerjob-deploy.json
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `scripts/template-validate-check.sh --clean` 通过
- `scripts/release-check.sh --skip-tests --skip-run` 通过
- `scripts/release-check-json-contract.sh --clean` 通过
- 文档版本入口一致（README/README.zh-CN/GITHUB_RELEASE）

## 5. 交付物

- 代码：`template list` 命令、templates catalog、门禁步骤
- 测试：`cmd/opskit` 与 `internal/templates` 新增用例
- 文档：本计划 + `SPEC_TEMPLATE_LIST_JSON`
