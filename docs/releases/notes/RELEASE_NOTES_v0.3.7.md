# OpsKit v0.3.7

发布日期：2026-02-08  
版本定位：模板接入前置增强版（Milestone 4 预备）

## 版本摘要

`v0.3.7` 在 `v0.3.6` 通用能力正式版基础上，重点补齐“模板接入前的门禁与契约化输出”，确保模板扩展阶段具备稳定、可脚本判读的输入条件。

## Added

- 新增 `opskit template validate --json`，输出机器可读校验结果
- 新增模板校验契约规范：`docs/specs/SPEC_TEMPLATE_VALIDATE_JSON.md`
- 新增发布门禁 summary 契约规范：`docs/specs/SPEC_RELEASE_CHECK_JSON.md`
- 新增模板 JSON 契约门禁脚本：`scripts/template-validate-check.sh`
- 新增 `release-check summary.json` 契约门禁脚本：`scripts/release-check-json-contract.sh`
- `scripts/release-check.sh` 新增 `--summary-json-file`

## Changed

- `scripts/release-check.sh` 默认接入模板 JSON 契约门禁（可用 `--skip-template-json-contract` 跳过）
- `scripts/release-check.sh` 失败时先落盘 `summary.json` 再退出
- `scripts/release-check.sh` 步骤级 `reasonCode` 语义统一：成功 `ok`，失败为具体步骤失败码
- `scripts/generic-readiness-check.sh` 增加对 `release-check/summary.json` 的契约校验
- `scripts/generic-readiness-check.sh` 新增可选参数 `--with-release-json-contract`
- `scripts/generic-readiness-check.sh` summary 新增路径字段：
  - `releaseCheckOutput`
  - `genericOutput`
  - `releaseJsonContractOutput`
- 文档与示例补充模板 JSON 成功/失败样例及关键断言字段

## 验证命令

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/template-validate-check.sh --clean
scripts/release-check-json-contract.sh --clean
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --with-release-json-contract --clean
scripts/release.sh --version v0.3.7 --clean
```

## Release 资产

- `opskit-v0.3.7-linux-arm64`
- `opskit-v0.3.7-linux-amd64`
- `checksums.txt`
- `release-metadata.json`

## 升级说明

- 建议先执行：
  - `scripts/release-check.sh --with-offline-validate`
  - `scripts/generic-readiness-check.sh --with-release-json-contract --clean`
- 若目标环境要求“全绿”，请在基线治理完成后执行 strict 门禁：
  - `scripts/release-check.sh --with-offline-validate --offline-strict-exit`
  - `scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean`
