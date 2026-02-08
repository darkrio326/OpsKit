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
- 新增通用检查：`fs_readonly`、`disk_inodes`、`clock_skew`
- `assets/templates/demo-server-audit.json` D 阶段接入上述 3 个检查
- `opskit status --json` 增加 `health` 字段（`ok|warn|fail`）
- `scripts/release.sh` 增加 `release-metadata.json`（`gitCommit/buildTime/goVersion`）
- `scripts/release.sh` 自动校验 `checksums.txt` 与二进制一致性
- `examples/generic-manage/run-af.sh` 增强为可门禁模式（期望退出码判定、`status.json` 固定落盘）
- 通用自检 `summary.json` 增加 `result/reasonCode/recommendedAction`
- 新增 `scripts/generic-readiness-check.sh`，统一发布门禁与通用链路门禁输出

### Changed

- `README.md` / `README.zh-CN.md` 下一版计划入口更新到 `v0.3.6-preview.1`
- `ROADMAP.md` 当前进行中版本滚动到 `v0.3.6-preview.1`
- `docs/GITHUB_RELEASE.md` 下一版文案与发版命令滚动到 `v0.3.6-preview.1`
- `scripts/release-check.sh` 增加 `--offline-summary-json-file` 透传
- `scripts/release-check.sh` summary 增加 `recommended action`
- `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md` / `docs/getting-started/GETSTART.md` 增加 `summary.json` 用法
- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md` 示例版本统一为 `v0.3.6` 口径
- `assets/templates/demo-server-audit.README.md` / `assets/templates/demo-hello-service.README.md` 补充“最低可运行变量集”和“常见失败”
- `docs/examples/generic-manage/README*.md` 补充 strict 模式与 `status.json` 产物说明
- `README*` / `GETSTART` / `GITHUB_RELEASE` 新增真实服务器前统一门禁入口说明

### Fixed

- 修复多处发布文档版本号不同步的问题（计划内持续收敛）

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
scripts/kylin-offline-validate.sh --bin ./.tmp/opskit-local --output ./.tmp/offline-validate --clean
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --clean --skip-tests
scripts/release.sh --version v0.3.6-preview.1 --clean
```
