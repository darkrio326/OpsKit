# OpsKit 版本规划规范（Release Planning Guide）

## 1. 每个版本计划建议包含的固定内容

每个 `docs/RELEASE_PLAN_*.md` 建议至少包含以下 10 项：

1. 版本目标（1~3 条，必须可衡量）
2. 范围边界（In Scope / Out of Scope）
3. 任务拆分（按 A/B/C 包组织）
4. 每包验收命令（脚本可直接执行）
5. 发布准入（Go/No-Go）
6. 交付物清单（代码、文档、资产）
7. 风险与回滚策略
8. 版本容量（本版最多承载的变更量）
9. 时间盒（按天或迭代）
10. 决策点（需要产品/技术负责人确认的项）

补充约定（当前采用）：

- 每个 `preview` 固定拆分为 **5 个功能点**
- 每个功能点建议控制在 3~5 条可验收变更
- 功能点必须给出可直接执行的验收命令

## 2. 版本容量建议

- 预览版（`preview`）：
  - 固定 5 个功能点
  - 每个功能点 3~5 个可验收事项
- 正式版（`v0.x.y`）：
  - 建议聚焦 1~3 个高价值主题
  - 优先“稳定性 + 可观测 + 可回滚”

## 3. 发版频率建议（OpsKit 当前阶段）

- `preview`：建议每 1~2 周一次（快速验证方向）
- 正式小版本：建议每 3~6 周一次（聚焦质量收敛）
- 若门禁未全绿：宁可延迟，不建议“按日期硬发”

## 4. 何时应该“多发版”

适合增加发版频次的场景：

- 接口/契约快速迭代期（例如当前 M4 模板能力）
- 用户反馈密集、修复需要快速下发
- 每次变更较小且可独立回滚

不适合增加发版频次的场景：

- 单次变更跨度大、门禁覆盖不足
- 跨模块耦合改动多，回滚成本高

## 5. 质量底线（发布前必须满足）

- `go test ./...` 通过
- `scripts/release-check.sh --with-offline-validate` 通过
- `scripts/release-check-json-contract.sh --clean` 通过
- `scripts/generic-readiness-check.sh --with-release-json-contract --clean` 通过
- 文档版本口径一致（README/ROADMAP/GITHUB_RELEASE/RELEASE_NOTES）
