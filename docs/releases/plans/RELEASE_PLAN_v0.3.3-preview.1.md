# OpsKit v0.3.3-preview.1 发布任务清单

> 主线：在不引入生产模板前提下，把通用能力做厚一层，每个版本交付可见增量。

## 1. 版本目标（本版增加工作量）

- Recover 与 Operate 再增强：补齐可观测与可追溯字段
- 状态模型增强：阶段摘要与失败原因规范化
- 证据链增强：accept 包结构一致性与可校验性再提升
- 发布质量门禁：固定回归套件 + docker 场景验证

## 2. 范围边界

### In Scope

- `internal/recover`、`internal/plugins/actions/recover_sequence.go` 的行为与输出规范
- `internal/state` 的摘要字段、状态聚合与 JSON schema 一致性
- 文档与 demo 模板的“可复制命令”完整性
- 发布脚本与检查清单强化

### Out of Scope

- 生产级中间件 deploy 模板
- 多节点编排与远程 agent
- 登录与权限系统

## 3. 任务拆分（每项可验收）

### A. Recover 结果标准化

- 为 recover 结果增加统一 `reason/code` 字段
- collect 产物增加来源标识（command/journal/file）
- 超限截断保留 `original_length` 与 `truncated_length`
- 当前进度：已完成（代码与测试已落地）

验收：

```bash
GOCACHE=$PWD/.gocache go test ./internal/recover ./internal/plugins/actions
```

### B. Operate 检查扩展

- 新增 2~3 个通用检查（优先：`ntp_sync`、`dns_resolve`、`service_restart_count`）
- 检查结果统一 severity 映射（fail/warn/info）
- 新增对应模板示例与 README 说明
- 当前进度：已完成（已落地 `ntp_sync`、`dns_resolve`、`systemd_restart_count`）

验收：

```bash
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
```

### C. State 与 Accept 一致性

- `lifecycle.json` 为每个阶段增加 `summary`（total/pass/warn/fail/skip）
- `accept` 产物新增一致性校验记录（manifest <-> hashes <-> state）
- 引入回归测试覆盖跨文件一致性
- 当前进度：已完成（stage summary + accept consistency record + 测试）

验收：

```bash
GOCACHE=$PWD/.gocache go test ./internal/state ./internal/engine
```

### D. 发布门禁升级

- 增加 `scripts/release-check.sh`：一键跑模板校验、关键单测、产物完整性
- 文档中统一“发布最小命令集”
- 增加一次 docker 回归作为建议门禁
- 当前进度：已完成（`release-check` 与发布脚本已可执行）

验收：

```bash
scripts/release-check.sh
scripts/release.sh --version v0.3.3-preview.1 --clean
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- demo 模板校验与最小链路可运行（A/D/accept）
- `scripts/release-check.sh` 通过
- 发布文档与 changelog 同步，无敏感信息

## 5. 建议时间盒

- Day 1-2：Recover + Operate 检查扩展
- Day 3：State/Accept 一致性
- Day 4：发布门禁脚本 + 文档回归
