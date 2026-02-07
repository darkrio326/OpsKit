# OpsKit v0.3.4-preview.1

> 预览版本（草稿）。主线聚焦“可视化补齐 + 跨环境健壮性”。

## 计划重点

- UI 展示阶段 summary 与一致性结果
- acceptance consistency 纳入 artifacts 索引
- checks 跨平台兼容性增强
- release-check 输出增强

## 计划变更

### Added

- UI 新增阶段 summary 展示（`total/pass/warn/fail/skip`）
- artifacts 高亮区新增 latest acceptance consistency 入口
- release-check 新增步骤耗时与总耗时汇总输出

### Changed

- accept/report 结构补充 consistency 摘要，并将 consistency 报告纳入 artifacts 索引
- check 降级原因指标标准化：新增 `check_degraded` / `check_degraded_reason` 统一口径
- `dns_resolve` 新增 `skip_network_query` 参数，用于受限网络场景跳过外部查询
- `load_average` 新增跨平台回退探测链路（`/proc/loadavg` -> `uptime` -> `sysctl vm.loadavg`）

### Fixed

- 修复 `load_average` 在非 Linux 环境下容易降级告警的问题
- 修复 `dns_resolve` 在受限网络场景下的误报/误降级问题

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
scripts/release-check.sh
scripts/release.sh --version v0.3.4-preview.1 --clean
```
