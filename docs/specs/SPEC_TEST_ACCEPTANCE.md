# OpsKit 测试与验收规范（Test & Acceptance Spec）

## 1. 目标

定义 v1 最小可交付测试路径，确保 A~F 链路可验证、可复现、可审计。
本规范同时定义发布前和真实服务器前的门禁口径，避免“脚本通过但判定不一致”。

## 2. 门禁策略（冻结）

### 2.1 默认门禁（non-strict，当前默认）

- 适用阶段：离线首轮回归、通用能力回归、发布前基础门禁。
- 退出码口径：`run A/D/accept/status` 允许 `0/1/3`。
- 判定目标：链路可运行、状态可刷新、证据可复核。
- 推荐脚本：
  - `scripts/release-check.sh --with-offline-validate`
  - `scripts/generic-readiness-check.sh --clean`

### 2.2 严格门禁（strict，可选）

- 适用阶段：标准服务器已完成基线治理后（挂载、服务、时间同步等）。
- 退出码口径：`run A/D/accept/status` 必须为 `0`。
- 判定目标：基础检查全绿，进入模板接入或生产验证。
- 推荐脚本：
  - `scripts/release-check.sh --with-offline-validate --offline-strict-exit`
  - `scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean`

## 3. 最小验收路径

建议按以下顺序执行：

1. `template validate`
2. `install`
3. `run AF`
4. `status`
5. `accept`
6. `handover`

## 4. 阶段级测试用例（最小集）

### A Preflight

- 构造端口冲突 -> 预期 `FAILED` 或 `WARN`（按模板 severity）
- 清理冲突后重跑 -> 预期恢复为 `PASSED/WARN`

### C Deploy

- 缺少离线包 -> 预期前置失败（退出码 `2` 或阶段失败）
- 启动失败 -> 预期有报告和失败建议

### D Operate

- 停止关键服务 -> 预期状态降级/失败
- 启动服务 -> 预期状态回升

### E Recover

- 重启或模拟异常 -> 预期触发恢复
- 连续失败 -> 预期熔断生效，避免无限重试

### F Accept

- 生成验收包 -> 预期 tar.gz 存在
- 校验证据项 -> 预期 hash/快照可复核

## 5. 结果判定

- 状态文件完整：`overall/lifecycle/services/artifacts.json`
- 报告可读、产物可追踪（artifacts 中可索引）
- 退出码符合规范（0/1/2/3/4）
- `status --json` 满足契约（至少包含 `command/schemaVersion/exitCode/health`）
- `acceptance-consistency` 报告可在 `artifacts.json` 中被索引

## 6. 验收命令（最小集合）

通用回归（默认门禁）：

```bash
env GOCACHE=$PWD/.gocache go test ./...
scripts/release-check.sh --with-offline-validate
scripts/generic-readiness-check.sh --clean
```

标准服务器基线治理完成后（严格门禁）：

```bash
scripts/release-check.sh --with-offline-validate --offline-strict-exit
scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
```

## 7. 回归建议

- 本地通用回归：`examples/generic-manage/run-af.sh`
- 容器纯净回归：`make -C examples/generic-manage docker-kylin-e2e`
- 每次模板变更前后各跑一次通用回归
