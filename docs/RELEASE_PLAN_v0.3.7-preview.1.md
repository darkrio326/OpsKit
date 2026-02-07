# OpsKit v0.3.7-preview.1 发布任务清单

> 主线：在 `v0.3.6` 通用能力正式版基础上，进入模板接入前置增强（不引入生产模板）。

## 1. 版本目标

- 把模板接入能力与通用能力进一步解耦，避免模板问题影响基础门禁
- 补齐模板参数约束、错误定位和验收口径
- 为 Milestone 4 模板库增强建立稳定输入条件

## 2. 范围边界

### In Scope

- 模板 schema/变量约束增强（必填、类型、枚举、默认值、错误路径）
- 模板验收文档补充（最小变量集、失败定位、验收命令）
- 回归门禁保持双轨：non-strict 默认、strict 可选

### Out of Scope

- 生产级中间件部署模板
- 客户环境适配逻辑
- 多节点编排与远程 agent

## 3. 任务拆分（可验收）

### A. 模板约束增强

- 增强模板变量校验错误可读性（路径化错误 + 建议修复）
- 统一模板校验输出结构，便于脚本/CI 判读

验收：

```bash
go run ./cmd/opskit template validate assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json
go run ./cmd/opskit template validate assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.env
```

### B. 模板文档收敛

- 每个 demo 模板补齐“最低可运行变量集/常见失败/产物预期”
- 中文与英文文档同步更新，避免口径漂移

验收：

```bash
rg -n "最低可运行变量集|常见失败|expected outputs|vars-file" assets/templates/*.README.md docs/examples/**/*.md
```

### C. 门禁与规范对齐

- `SPEC_TEST_ACCEPTANCE` 保持 non-strict/strict 双轨一致
- 发布说明同步模板接入前置策略

验收：

```bash
rg -n "non-strict|strict|0/1/3|offline-strict-exit|generic-readiness-check" docs/specs/*.md docs/GITHUB_RELEASE.md
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `scripts/release-check.sh --with-offline-validate` 通过
- `scripts/generic-readiness-check.sh --clean` 通过
- 模板校验错误提示与文档示例一致

## 5. 建议时间盒

- Day 1：A
- Day 2：B
- Day 3：C + 回归
