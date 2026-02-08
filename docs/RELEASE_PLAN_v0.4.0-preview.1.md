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

## 5. 建议时间盒

- Day 1：A（模板模型与校验）
- Day 2：B（demo 模板扩展）
- Day 3：C（门禁收敛 + 回归）
