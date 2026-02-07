# OpsKit v0.3.6-preview.1

> 预览版本（草稿）。主线聚焦“离线回归摘要化 + 通用检查扩展 + 发布资产可复核”。

## 计划重点

- 离线回归输出结构化 summary，失败定位更直接
- D 阶段增加更通用的离线检查项（文件系统/inode/时间偏移）
- 发布脚本补齐 metadata 与校验闭环

## 计划变更

### Added

- 新增 `docs/RELEASE_PLAN_v0.3.6-preview.1.md`
- `scripts/kylin-offline-validate.sh` 支持 `--summary-json-file`
- 离线回归 `summary.json` 增加 `result/reasonCode/stageResults/durationSeconds`
- 新增通用检查：`fs_readonly`、`disk_inodes`、`clock_skew`（计划）
- 发布资产 metadata 与 checksum 一致性自检（计划）

### Changed

- `README.md` / `README.zh-CN.md` 下一版计划入口更新到 `v0.3.6-preview.1`
- `ROADMAP.md` 当前进行中版本滚动到 `v0.3.6-preview.1`
- `docs/GITHUB_RELEASE.md` 下一版文案与发版命令滚动到 `v0.3.6-preview.1`
- `scripts/release-check.sh` 增加 `--offline-summary-json-file` 透传
- `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md` / `docs/getting-started/GETSTART.md` 增加 `summary.json` 用法

### Fixed

- 修复多处发布文档版本号不同步的问题（计划内持续收敛）

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
scripts/kylin-offline-validate.sh --bin ./.tmp/opskit-local --output ./.tmp/offline-validate --clean
scripts/release-check.sh --with-offline-validate
scripts/release.sh --version v0.3.6-preview.1 --clean
```
