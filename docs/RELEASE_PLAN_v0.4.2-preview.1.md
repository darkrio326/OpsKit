# OpsKit v0.4.2-preview.1 发布任务清单

> 目标：继续推进 M4，在保持 v0.4.x 设计冻结约束下，完成“通用 deploy 模板 + 行业 deploy 模板”第一轮可交付扩展。

## 1. 版本目标

- 抽象通用 deploy-manage 模板骨架，减少新项目接入成本
- 完成 MinIO/ELK/PowerJob 三类自部署模板首轮收敛
- 固化“黑箱管理 vs 自部署”双模式接入边界
- 保持模板失败可交付（可验收、可复核）

## 2. 范围边界

### In Scope

- 通用 deploy 模板能力增强（变量、README、门禁）
- MinIO/ELK/PowerJob 模板结构收敛与示例变量完善
- 模板交付门禁与回归脚本补齐
- UI 模板视图只读增强（不引入执行逻辑）

### Out of Scope

- 新增生命周期阶段
- 改动 `executil` 执行语义
- 客户生产参数与客户环境脚本
- 多节点编排与控制面

## 3. 功能点拆分（固定 5 项）

### 功能点 1：通用 deploy-manage 模板骨架

- 收敛 `demo-generic-selfhost-deploy` 的最小变量集与阶段职责
- 明确“模板做什么/不做什么/失败可交付”的 README 契约
- 输出统一 vars 示例（只表达差异，不承载逻辑）

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-generic-selfhost-deploy.json assets/templates/demo-generic-selfhost-deploy.json
scripts/template-delivery-check.sh --clean
```

### 功能点 2：MinIO 模板收敛

- 统一 MinIO 模板变量分组（service/path/network/package）
- 增加最短命令链与失败场景说明
- 校验 A/D/Accept 失败可交付

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-minio-deploy.json assets/templates/demo-minio-deploy.json
scripts/template-delivery-check.sh --clean
```

### 功能点 3：ELK 多服务模板收敛

- 明确 ES/Logstash/Kibana 的多服务边界与责任
- 补齐高级变量说明（JVM/TLS/pipeline）的非逻辑化约束
- 增加多服务失败诊断证据入口

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-elk-deploy.json assets/templates/demo-elk-deploy.json
scripts/template-delivery-check.sh --clean
```

### 功能点 4：PowerJob 模板收敛

- 明确应用服务与依赖边界
- 补齐示例变量与单机前提说明
- 统一 Delivery Level 与最短命令链

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-powerjob-deploy.json assets/templates/demo-powerjob-deploy.json
scripts/template-delivery-check.sh --clean
```

### 功能点 5：门禁与发布流程增强

- `release-check` 增加 deploy 模板校验组合冒烟
- 对齐 README/ROADMAP/GITHUB_RELEASE 当前预览入口
- 输出 `v0.4.2-preview.1` 的发布说明草稿

验收：

```bash
scripts/release-check.sh --skip-tests --skip-run
rg -n "v0.4.2-preview.1" README.md README.zh-CN.md ROADMAP.md docs/GITHUB_RELEASE.md docs/README.md
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `scripts/release-check.sh --with-offline-validate` 通过
- `scripts/template-delivery-check.sh --clean` 通过
- 版本入口一致（README/ROADMAP/docs index/GITHUB_RELEASE）

## 5. 交付物

- 当前版本计划：`docs/RELEASE_PLAN_v0.4.2-preview.1.md`
- 当前版本说明：`docs/RELEASE_NOTES_v0.4.2-preview.1.md`
- 历史归档：`docs/releases/notes/RELEASE_NOTES_v0.4.1-preview.1.md`、`docs/releases/plans/RELEASE_PLAN_v0.4.1-preview.1.md`
