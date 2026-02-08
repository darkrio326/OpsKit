# OpsKit v0.4.0-preview.1 发布任务清单

> 主线：正式进入 Milestone 4（模板能力增强），在通用能力与门禁契约稳定基础上推进模板库与模板验收体系。

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

## 3. 任务拆分（可验收）

### A. 模板模型增强

- 扩展模板变量描述能力（描述、示例、约束提示）
- 强化模板未知字段与未解析变量定位
- 增补模板校验错误码文档

验收：

```bash
go run ./cmd/opskit template validate --json assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json
go run ./cmd/opskit template validate --json assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json
```

### B. demo 模板扩展

- 新增 1~2 个去生产化模板（仅通用检查/动作）
- 每个模板附最小变量集、预期产物、常见失败
- 模板示例数据补齐（json/env）

验收：

```bash
rg -n "最低可运行变量集|常见失败|expected outputs|vars-file" assets/templates/*.README.md docs/examples/**/*.md
```

### C. 门禁与发布流程收敛

- `release-check` / `generic-readiness` 对模板契约门禁持续可用
- 发布文档切换到 v0.4.0-preview.1 口径
- 版本文档与脚本命令保持一致

验收：

```bash
scripts/template-validate-check.sh --clean
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

- 本版建议最多 3 个主题包（A/B/C）
- 每个主题包最多 3~5 项可验收变更
- 单次发布建议 PR 数量控制在 5~10 个以内（可回滚、可审阅）

## 8. 建议时间盒

- Day 1：A（模板模型与校验）
- Day 2：B（demo 模板扩展）
- Day 3：C（门禁收敛 + 回归）

## 9. 决策点（需确认）

- `v0.4.0-preview.1` 是否仅做“模板模型+门禁”，不引入新 action 插件
- demo 模板数量是否固定为 2 个（避免范围膨胀）
- preview 完成后是否直接进入 `v0.4.0` 正式版，还是先 `v0.4.0-preview.2`
