# 银河麒麟 V10 离线回归验证清单（OpsKit v0.3.x）

本文用于发布后用户侧验收，目标是确认二进制在离线麒麟 V10 服务器可运行，并产出可追踪状态与证据。

## 1. 前置条件

- 已将 release 包拷贝到离线服务器并安装 `opskit`
- 可写目录：例如 `/data/opskit-regression-v036`
- 可执行 `grep`（系统默认通常具备）

## 2. 一次性回归命令

优先使用一键脚本：

```bash
scripts/kylin-offline-validate.sh \
  --bin /usr/local/bin/opskit \
  --output /data/opskit-regression-v036 \
  --json-status-file /data/opskit-regression-v036/status.json \
  --summary-json-file /data/opskit-regression-v036/summary.json \
  --clean
```

若你希望“所有阶段必须 exit=0”再算通过，可开启严格模式：

```bash
scripts/kylin-offline-validate.sh \
  --bin /usr/local/bin/opskit \
  --output /data/opskit-regression-v036 \
  --json-status-file /data/opskit-regression-v036/status.json \
  --summary-json-file /data/opskit-regression-v036/summary.json \
  --strict-exit \
  --clean
```

若需手工执行，可使用：

```bash
set -e
export OPSKIT_OUT=/data/opskit-regression-v036
mkdir -p "$OPSKIT_OUT"

opskit run A --template generic-manage-v1 --output "$OPSKIT_OUT"
opskit run D --template generic-manage-v1 --output "$OPSKIT_OUT"
opskit accept --template generic-manage-v1 --output "$OPSKIT_OUT"
opskit status --output "$OPSKIT_OUT"
```

## 3. 验收点（必须满足）

1. 命令执行层

- `run A`、`run D`、`accept` 返回码属于 `0/1/3`
- `status` 返回码属于 `0/1/3`

2. 状态文件层

- 存在 `state/overall.json`
- 存在 `state/lifecycle.json` 且各阶段有 `summary(total/pass/warn/fail/skip)`
- 存在 `state/artifacts.json`
- 存在 `status.json`，且含 `command/schemaVersion/exitCode/health`
- 存在 `summary.json`，且含 `result/reasonCode/stageResults`

3. 报告与证据层

- 存在 `reports/accept-*.html`
- `state/artifacts.json` 中包含 `acceptance-consistency-*.json` 路径
- `reports/accept-*.html` 包含 `consistency` 区块

## 4. 快速检查命令

```bash
grep -R "\"summary\"" "$OPSKIT_OUT/state/lifecycle.json"
grep -R "acceptance-consistency-" "$OPSKIT_OUT/state/artifacts.json"
grep -R "\"consistency\"" "$OPSKIT_OUT/reports"/accept-*.html
grep -R "\"schemaVersion\"" "$OPSKIT_OUT/status.json"
grep -R "\"exitCode\"" "$OPSKIT_OUT/status.json"
grep -R "\"health\"" "$OPSKIT_OUT/status.json"
grep -R "\"result\"" "$OPSKIT_OUT/summary.json"
grep -R "\"reasonCode\"" "$OPSKIT_OUT/summary.json"
```

## 5. UI 验证

```bash
opskit web --output "$OPSKIT_OUT" --listen 127.0.0.1:18080
```

本机访问 `http://127.0.0.1:18080`，应能看到：

- A~F 阶段卡片与状态
- 阶段 summary 计数
- artifacts 区域可见 acceptance consistency 入口

## 6. 失败判定与处理

- `exit=4`：同一输出目录有并发任务；等待或切换 `--output`
- `status.json health` 与退出码对照：`ok=0`、`warn=3`、`fail=1`
- `template not found`：离线机仅有二进制时用内置模板 ID（`generic-manage-v1`）
- 无 `acceptance-consistency`：确认执行过 `accept`，并使用相同输出目录执行 `status`/`web`
- UI 空白：确认 `--output` 指向已产生 `state/*.json` 的目录
- `status=1`：通常表示存在 FAIL 检查项，不代表程序不可用；优先核对 state/reports/artifacts 是否已生成
