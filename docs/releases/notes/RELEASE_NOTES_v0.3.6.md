# OpsKit v0.3.6

发布日期：2026-02-07  
版本定位：通用能力正式版（模板接入前基线版）

## 版本摘要

`v0.3.6` 聚焦“通用能力可验收闭环”，目标是让离线/信创环境在不依赖业务模板的情况下，完成可执行、可观测、可复核的运维链路。

## Added

- 新增通用检查：`fs_readonly`、`disk_inodes`、`clock_skew`
- 新增统一门禁脚本：`scripts/generic-readiness-check.sh`
- 离线回归脚本支持 `--summary-json-file`
- `status --json` 增加 `health` 字段（`ok|warn|fail`）
- 发布脚本新增 `release-metadata.json`（`gitCommit/buildTime/goVersion`）

## Changed

- `release-check` 支持离线回归摘要透传（`--offline-summary-json-file`）
- `release-check` 汇总输出建议动作（`continue_release` / `block_release`）
- 通用回归摘要统一：`result/reasonCode/recommendedAction`
- 离线脚本默认输出目录更新为版本无关路径：`/data/opskit-regression`
- 门禁策略明确为双轨：
  - 默认 non-strict：允许 `run A/D/accept/status` 返回 `0/1/3`
  - strict 可选：基线治理完成后要求 `run A/D/accept/status` 全 `0`

## Fixed

- 修复多处发布文档版本口径不一致的问题
- 修复离线回归与发布门禁在 strict 适用场景描述不清的问题

## 验证命令

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --clean
scripts/release.sh --version v0.3.6 --clean
```

## Release 资产

- `opskit-v0.3.6-linux-arm64`
- `opskit-v0.3.6-linux-amd64`
- `checksums.txt`
- `release-metadata.json`

## 升级说明

- 从 preview 版本升级时，建议优先按 non-strict 跑一轮通用回归，确认状态与证据产物完整。
- 若目标环境要求“基础检查全绿”，请在环境基线治理后开启 strict 门禁复验。
