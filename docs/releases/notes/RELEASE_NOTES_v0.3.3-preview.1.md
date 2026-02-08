# OpsKit v0.3.3-preview.1

> 预览版本（草稿）。主线聚焦“通用能力做厚 + 发布门禁收敛”。

## 计划重点

- Recover 结果结构标准化
- Operate 通用检查扩展
- state/accept 一致性增强
- 发布门禁脚本化（release-check）

## 计划变更

### Added

- 新增 `scripts/release-check.sh`，固化发布前最小回归（测试/模板校验/A-D-accept 干跑）
- Recover 结果新增统一 `recover_reason_code` 指标
- Recover collect 产物新增来源标识（`source: command/journal/file`）
- Recover collect 产物新增截断元数据（`originalLength`/`truncatedLength`）
- 新增通用检查：`ntp_sync`、`dns_resolve`、`systemd_restart_count`
- Accept 阶段新增一致性校验记录：`acceptance-consistency-*.json`

### Changed

- Recover circuit 状态新增 `lastErrorCode`，用于状态汇总与追踪
- Recover summary 新增 `lastReasonCode`
- demo 审计模板 D 阶段新增 `dns_resolve` 检查
- demo 审计模板 D 阶段新增 `ntp_sync` 检查
- `lifecycle.json` 阶段新增 `summary(total/pass/warn/fail/skip)` 计数结构
- accept 新增 `manifest <-> hashes <-> state` 一致性校验流程与指标
- accept 一致性策略明确：生成 `acceptance-consistency-*.json`，并以 `accept_consistency` / `accept_consistency_missing` 指标输出结果

### Fixed

- 修复 recover 失败原因仅文本、不易聚合的问题（改为 code + message 双轨）
- 修复 accept 证据包一致性缺少显式记录的问题

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit run A --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
scripts/release-check.sh
scripts/release.sh --version v0.3.3-preview.1 --clean
```
