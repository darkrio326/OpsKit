# OpsKit v0.3.5-preview.1

> 预览版本（草稿）。主线聚焦“离线回归自动化 + 判读口径统一”。

## 计划重点

- 麒麟离线一键回归脚本
- 回归文档执行口径统一（真实执行）
- 发布文档入口收敛

## 计划变更

### Added

- 新增 `scripts/kylin-offline-validate.sh`，支持离线机一键回归
- 新增 `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md` 离线回归清单
- 新增 `docs/RELEASE_PLAN_v0.3.5-preview.1.md`
- 新增 `opskit status --json` 机器可读状态输出
- 新增 `docs/specs/SPEC_STATUS_JSON.md`（`status --json` 契约规范）
- 离线回归脚本新增 `--json-status-file`，可保存 `status --json` 结果
- 离线回归脚本新增 `--strict-exit`，可选严格通过模式
- `scripts/release-check.sh` 新增 `--with-offline-validate` 可选门禁
- `scripts/release-check.sh` 新增 `--offline-strict-exit` 透传严格模式

### Changed

- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md` 回归步骤调整为真实执行
- `README.md` / `README.zh-CN.md` / `docs/getting-started/GETSTART.md` 增加离线一键回归入口
- `ROADMAP.md` 最新发布版本更新到 `v0.3.4-preview.1`

### Fixed

- 修复离线回归文档中 dry-run 与证据产物校验口径不一致的问题

## 计划验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
scripts/kylin-offline-validate.sh --bin ./.tmp/opskit-local --output ./.tmp/offline-validate --clean
GOCACHE=$PWD/.gocache go run ./cmd/opskit status --output ./.tmp/offline-validate --json
scripts/release-check.sh
```
