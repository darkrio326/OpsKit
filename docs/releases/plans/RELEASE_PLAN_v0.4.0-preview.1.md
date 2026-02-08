# OpsKit v0.4.0-preview.1 发布任务清单（已并入 preview.2）

> 该计划已并入 `docs/RELEASE_PLAN_v0.4.0-preview.2.md` 统一收口；本文件保留用于追溯。

## 1. 版本目标

- 把模板能力从“可用”推进到“可规模化维护”
- 增强模板变量模型、校验与错误定位的一致性
- 为后续 ELK/OA/MinIO 等模板化扩展建立统一规范

## 2. 范围边界

### In Scope

- 模板 schema/vars 能力增强（分组、默认值、枚举、描述）
- 增加去生产化 demo 模板（不含客户脚本）
- 模板验收文档与脚本门禁收敛
- 模板输出与状态视图映射优化（不改核心状态语义）

### Out of Scope

- 生产级中间件一键部署脚本
- 客户环境适配逻辑
- 多节点编排与远程 agent

## 3. 功能点拆分（每个 preview 固定 5 项）

### 功能点 1（已完成）：模板变量 `example` 元数据与校验

- 为 `template.vars` 增加 `example` 字段（字符串）
- `template validate` 时校验 `example` 与 `type/enum` 一致性
- demo 模板补齐关键变量示例

验收：

```bash
go test ./internal/schema
go run ./cmd/opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
```

### 功能点 2（已完成）：模板变量分组（group）与展示约定

- 为变量增加 `group`（如 `service`/`paths`/`runtime`）
- 补充分组命名规范与兼容策略
- 文档与示例同步

验收：

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-hello-service.json --json assets/templates/demo-hello-service.json
go run ./cmd/opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
rg -n "\"group\"|变量分组|group" assets/templates docs/examples docs/specs
```

### 功能点 3（已完成）：新增 1 个去生产化 demo 模板

- 新增纯通用场景模板（不含生产中间件）
- 补齐最小变量集、预期产物与常见失败说明

验收：

```bash
./opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
./opskit template validate --vars-file examples/vars/demo-hello-service.json --json assets/templates/demo-hello-service.json
./opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
```

### 功能点 4（已完成）：模板门禁输出增强

- 模板校验错误码清单补齐
- 示例中增加失败断言字段
- 契约文档更新

验收：

```bash
scripts/template-validate-check.sh --clean
rg -n "template_var_|template_unknown_|template_file_|template_json_" docs/specs docs/getting-started
```

### 功能点 5（待开始）：M4 版本化门禁收敛

- `release-check` / `generic-readiness` 维持模板契约门禁
- 版本文档与脚本命令口径一致
- 发布前门禁清单按 M4 固化

验收：

```bash
scripts/release-check-json-contract.sh --clean
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --with-release-json-contract --clean
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- 模板 JSON 契约门禁通过
- 发布门禁 summary 契约门禁通过
- 文档版本指针一致（README/ROADMAP/GITHUB_RELEASE/RELEASE_NOTES）

## 5. 交付物清单

- 代码：
  - 模板模型与校验增强实现
  - 新增 demo 模板与 vars 示例
  - 模板门禁脚本与回归脚本更新
- 文档：
  - `RELEASE_NOTES_v0.4.0-preview.1.md`
  - 模板 README 与 `docs/examples` 对应说明
  - 规格文档增补（如新增错误码/字段）
- 发布资产：
  - 双架构二进制
  - `checksums.txt`
  - `release-metadata.json`

## 6. 风险与回滚

- 风险：
  - 模板校验规则增强导致旧模板兼容性问题
  - 模板文档与实际行为不一致导致门禁误判
  - 单次引入过多模板导致回归面过大
- 缓解：
  - 维持 v1 契约“仅追加字段”
  - 每个模板变更都附最小可运行示例
  - preview 阶段按主题包拆分提交并独立验收
- 回滚：
  - 若出现兼容性问题，优先回滚模板层变更，不回滚通用门禁主链
  - 保持 `v0.3.7` 作为稳定回退基线

## 7. 版本容量控制

- 本版固定 5 个功能点（每个 preview 一致）
- 每个功能点最多 3~5 项可验收变更
- 单次发布建议 PR 数量控制在 5~10 个以内（可回滚、可审阅）

## 8. 建议时间盒

- Day 1：功能点 1 + 2（已完成）
- Day 2：功能点 3 + 4（已完成）
- Day 3：功能点 5 + 回归发布（进行中）

## 9. 决策点（需确认）

- `v0.4.0-preview.1` 是否仅做“模板模型+门禁”，不引入新 action 插件
- demo 模板数量是否固定为 3 个（避免范围膨胀）
- preview 完成后是否直接进入 `v0.4.0` 正式版，还是先 `v0.4.0-preview.2`
