# OpsKit 测试与验收规范（Test & Acceptance Spec）

## 1. 目标

定义 v1 最小可交付测试路径，确保 A~F 链路可验证、可复现、可审计。

## 2. 最小验收路径

建议按以下顺序执行：

1. `template validate`
2. `install`
3. `run AF`
4. `status`
5. `accept`
6. `handover`

## 3. 阶段级测试用例（最小集）

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

## 4. 结果判定

- 状态文件完整：`overall/lifecycle/services/artifacts.json`
- 报告可读、产物可追踪（artifacts 中可索引）
- 退出码符合规范（0/1/2/3/4）

## 5. 回归建议

- 本地通用回归：`examples/generic-manage/run-af.sh`
- 容器纯净回归：`make -C examples/generic-manage docker-kylin-e2e`
- 每次模板变更前后各跑一次通用回归
