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
- （待补）release-check 统计输出

### Changed

- （待补）accept/report 结构补充 consistency 摘要
- （待补）check 降级原因指标标准化

### Fixed

- （待补）load_average 非 Linux 解析降级问题
- （待补）dns_resolve 受限环境误报问题

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
scripts/release-check.sh
scripts/release.sh --version v0.3.4-preview.1 --clean
```
