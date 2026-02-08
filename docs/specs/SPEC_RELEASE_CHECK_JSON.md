# OpsKit `release-check` Summary JSON 输出契约规范

## 1. 目标与范围

本规范定义 `scripts/release-check.sh` 的机器可读汇总输出契约，供：

- 发布前 CI 门禁
- 本地自动化回归脚本
- 上层门禁脚本（如 `generic-readiness-check.sh`）

使用范围仅限 `summary.json` 文件，不包含脚本的普通终端输出。

## 2. 调用约定

示例：

```bash
scripts/release-check.sh --with-offline-validate --summary-json-file ./.tmp/release-check/summary.json
```

行为约定：

- 脚本成功退出码为 `0`，失败为非 `0`
- 无论成功或失败，都会写出 `summary.json`
- 调用方应同时读取进程退出码与 `summary.json`

## 3. 顶层字段（`schemaVersion = v1`）

字段：

- `schemaVersion`：当前固定 `v1`
- `generatedAt`：生成时间（ISO8601）
- `result`：`pass|fail`
- `reasonCode`：本次执行主因（见第 5 节）
- `recommendedAction`：`continue_release|block_release`
- `output`：release-check 输出目录
- `goCacheDir`：本次使用的 Go 缓存目录
- `skipTests`：是否跳过 `go test`
- `skipTemplateJsonContract`：是否跳过模板 JSON 契约门禁
- `skipRun`：是否跳过 `run A/D/accept/status` dry-run
- `withOfflineValidate`：是否执行离线门禁
- `offlineStrictExit`：离线门禁是否严格模式
- `steps`：执行步骤总数
- `totalDurationSeconds`：总耗时秒数
- `stepResults[]`：步骤结果数组
- `offline`：离线门禁相关输出路径

`stepResults[]` 字段：

- `name`：步骤名称
- `elapsedSeconds`：该步骤耗时
- `exitCode`：该步骤退出码
- `reasonCode`：`ok` 或该步骤失败原因码

`offline` 字段：

- `output`：离线门禁输出目录
- `jsonStatusFile`：离线 `status --json` 文件路径
- `summaryJsonFile`：离线门禁 summary 文件路径

## 4. 示例

成功：

```json
{
  "schemaVersion": "v1",
  "generatedAt": "2026-02-08T08:24:18Z",
  "result": "pass",
  "reasonCode": "ok",
  "recommendedAction": "continue_release",
  "steps": 3,
  "stepResults": [
    {
      "name": "template validate demo-server-audit",
      "elapsedSeconds": 2,
      "exitCode": 0,
      "reasonCode": "ok"
    }
  ]
}
```

失败：

```json
{
  "schemaVersion": "v1",
  "result": "fail",
  "reasonCode": "step_failed_build_offline_binary",
  "recommendedAction": "block_release",
  "stepResults": [
    {
      "name": "build offline validation binary",
      "elapsedSeconds": 1,
      "exitCode": 1,
      "reasonCode": "step_failed_build_offline_binary"
    }
  ]
}
```

## 5. `reasonCode` 约定（v1）

主因 `reasonCode`：

- `ok`
- `step_failed_go_test`
- `step_failed_template_validate_demo_server_audit`
- `step_failed_template_validate_demo_hello_service`
- `step_failed_template_validate_json_contract`
- `step_failed_run_a_dry_run`
- `step_failed_run_d_dry_run`
- `step_failed_accept_dry_run`
- `step_failed_status_refresh`
- `step_failed_build_offline_binary`
- `step_failed_offline_validation_gate`

说明：

- 顶层 `reasonCode` 只表达本次失败主因（首个失败步骤）
- `stepResults[].reasonCode` 为该步骤自身结果：成功为 `ok`，失败为步骤失败码

## 6. 兼容性策略

- `v1` 内仅允许追加字段，不改变既有字段语义
- `reasonCode` 可新增，不允许改变已有代码语义
- 调用方应忽略未知字段，并允许 `stepResults[]` 增长

## 7. 调用方建议

- 先看进程退出码，再看 `summary.json`
- 失败时优先读取顶层 `reasonCode`，再定位对应 `stepResults[]`
- 在 CI 中可直接断言：
  - `result=pass`
  - `reasonCode=ok`
  - `recommendedAction=continue_release`
