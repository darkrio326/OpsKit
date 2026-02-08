# OpsKit v0.4.0-preview.3 发布任务清单

> 目标：作为 `v0.4.0-preview.2` 的补齐版本，完成 M4 关键改动入版与发布文档口径收敛。

## 1. 版本目标

- 补齐 `preview.2` 之后的模板与门禁改动到已发布版本
- 固化黑箱模板命名策略（模板通用名 + vars 预设）
- 固化发布文档归档结构，避免根目录多版本混杂

## 2. 范围边界

### In Scope

- 黑箱模板与 vars 预设（default/FCS/Kingdee）
- `template-validate-check` 与 `release-check` 门禁覆盖三套 vars
- 发布文档归档结构调整与入口链接更新
- `preview.3` 发布文档与发版资产准备

### Out of Scope

- 生产级部署动作
- 客户环境脚本
- 多节点 agent 与编排

## 3. 功能点拆分（固定 5 项）

### 功能点 1（完成）：补齐黑箱模板入版

- 引入 `demo-blackbox-middleware-manage` 模板与 README
- 模板保持通用命名，不绑定厂商后缀

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage.json --json assets/templates/demo-blackbox-middleware-manage.json
```

### 功能点 2（完成）：黑箱 vars 三套预设

- 新增 default/FCS/Kingdee 三套 vars（json + env）
- README 示例覆盖三套预设

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json --json assets/templates/demo-blackbox-middleware-manage.json
go run ./cmd/opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json --json assets/templates/demo-blackbox-middleware-manage.json
```

### 功能点 3（完成）：门禁脚本覆盖收敛

- `template-validate-check` 增加 blackbox default/FCS/Kingdee 正向校验
- `release-check` 增加 blackbox default/FCS/Kingdee 校验与 dry-run

验收：

```bash
scripts/template-validate-check.sh --clean
scripts/release-check.sh --skip-tests --skip-run
```

### 功能点 4（完成）：发布文档归档与入口统一

- 历史 `RELEASE_NOTES_*` 移入 `docs/releases/notes/`
- 历史 `RELEASE_PLAN_*` 移入 `docs/releases/plans/`
- `docs/` 根目录仅保留当前版本 release 文档

验收：

```bash
ls -1 docs | rg "RELEASE_(NOTES|PLAN)_v0.4.0-preview.3"
ls -1 docs/releases/notes | head
ls -1 docs/releases/plans | head
```

### 功能点 5（完成）：preview.3 发版就绪

- 生成 `preview.3` notes/plan
- 更新 README/GITHUB_RELEASE/docs 索引到 `preview.3`
- 生成 `dist` 产物用于 GitHub Release

验收：

```bash
scripts/release.sh --version v0.4.0-preview.3 --clean
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `template-validate-check` 通过
- `release-check` 通过
- `release-check-json-contract` 通过
- 版本入口一致（README/README.zh-CN/docs/README/GITHUB_RELEASE）

## 5. 交付物

- 代码：模板、vars、门禁脚本
- 文档：`RELEASE_NOTES_v0.4.0-preview.3.md`、`RELEASE_PLAN_v0.4.0-preview.3.md`
- 资产：`dist/opskit-v0.4.0-preview.3-linux-{arm64,amd64}`、`checksums.txt`、`release-metadata.json`
