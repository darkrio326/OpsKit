# OpsKit v0.4.0-preview.2 发布任务清单

> 目标：在 `v0.4.0-preview.1` 基础上完成 M4 收口，按“每个 preview 固定 5 个功能点”一次做完，并接入通用 `manage` 模板门禁。

## 1. 版本目标

- 完成 M4 五个功能点闭环并固化门禁
- 把 `generic-manage-v1` 接入模板校验与发布门禁默认链路
- 冻结 `v0.4.0` 正式版前的模板契约口径

## 2. 范围边界

### In Scope

- `template.vars` 元数据增强（`example/group`）与校验
- demo 模板扩展（含 runtime baseline）
- 模板 JSON 契约门禁增强（含负例断言）
- 发布门禁与通用 manage 模板门禁接入
- 文档口径统一到 `preview.2`

### Out of Scope

- 生产级中间件部署模板
- 客户环境定制脚本
- 多节点编排与远程 agent

## 3. 功能点拆分（固定 5 项）

### 功能点 1（完成）：模板变量 `example` 元数据与校验

- `template.vars` 支持 `example`
- `template validate` 校验 `example` 类型/枚举一致性
- demo 模板补齐关键示例值

验收：

```bash
go test ./internal/schema
go run ./cmd/opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
```

### 功能点 2（完成）：模板变量分组 `group` 与兼容策略

- `template.vars` 支持 `group`
- `group` 格式约束：`^[a-z][a-z0-9_]{0,31}$`
- 规格文档补充分组命名建议

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-hello-service.json --json assets/templates/demo-hello-service.json
rg -n "\"group\"|template.vars|变量分组" assets/templates docs/specs
```

### 功能点 3（完成）：新增去生产化 demo 模板

- 新增 `demo-runtime-baseline`（A/D/F）
- 新增 `demo-blackbox-middleware-manage`（A~F，黑箱中间件 Manage）
- 补齐模板说明与 vars 示例
- 维持“无生产中间件脚本”边界

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
```

### 功能点 4（完成）：模板门禁输出增强

- `template-validate-check` 增加变量类型错误负例
- 错误码清单补齐到规格文档
- `template validate` 参数顺序兼容（模板路径前后都支持 flags）

验收：

```bash
go test ./cmd/opskit ./internal/schema
scripts/template-validate-check.sh --clean
```

### 功能点 5（完成）：M4 门禁收敛 + 通用 manage 模板接入

- `release-check` 默认接入 `generic-manage-v1` 模板校验
- `release-check` dry-run 增加 `generic-manage-v1` 的 A/D 校验步骤
- `template-validate-check` 默认接入 `generic-manage-v1` 正向校验
- 版本文档口径统一到 `v0.4.0-preview.2`

验收：

```bash
scripts/release-check.sh --skip-tests --skip-run
scripts/release-check-json-contract.sh --clean
scripts/template-validate-check.sh --clean
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `template-validate-check` 通过（含正负例）
- `release-check` 通过（含 generic-manage 模板步骤）
- `release-check-json-contract` 通过
- 文档版本入口一致（README/ROADMAP/GITHUB_RELEASE/RELEASE_NOTES）

## 5. 交付物

- 代码：schema/CLI/门禁脚本/模板与 vars 示例
- 文档：`RELEASE_NOTES_v0.4.0-preview.2.md`、spec 更新
- 发布资产：二进制、`checksums.txt`、`release-metadata.json`

## 6. 风险与回滚

- 风险：模板规则增强导致旧模板误报
- 缓解：保持 v1 “只追加字段，不改语义”，门禁覆盖正负例
- 回滚：优先回滚模板/门禁层改动，不回滚核心状态链路

## 7. 容量与时间盒

- 固定 5 个功能点，每个功能点 3~5 项可验收变更
- Day 1：功能点 1/2
- Day 2：功能点 3/4
- Day 3：功能点 5 + 回归发布
