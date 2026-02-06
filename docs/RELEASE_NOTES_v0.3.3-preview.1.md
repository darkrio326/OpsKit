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
- （待补）Recover 统一 reason/code 与 collect 来源字段
- （待补）Operate 新增通用检查项

### Changed

- （待补）阶段摘要结构统一
- （待补）accept 产物一致性校验策略

### Fixed

- （待补）

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit run A --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
scripts/release-check.sh
scripts/release.sh --version v0.3.3-preview.1 --clean
```
