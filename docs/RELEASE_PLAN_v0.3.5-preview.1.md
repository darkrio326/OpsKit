# OpsKit v0.3.5-preview.1 发布任务清单

> 主线：把离线可用性与自动化验证再推进一层，降低“会跑但不易判读”的使用门槛。

## 1. 版本目标

- 提供麒麟离线一键回归脚本（A/D/F + 关键产物校验）
- 增强状态可读性（CLI 文本 + `status --json` 机器可读）
- 补齐最小回归资产（命令、退出码、产物路径）

## 2. 范围边界

### In Scope

- `scripts/kylin-offline-validate.sh` 脚本与文档联动
- `docs/getting-started` 离线回归路径固化
- release 文档与 changelog 同步

### Out of Scope

- 生产级 deploy 模板
- 多节点编排
- 账号/权限系统

## 3. 任务拆分（每项可验收）

### A. 离线回归脚本化

- 新增离线一键回归脚本：模板校验、`run A`、`run D`、`accept`、`status`
- 输出校验：`state/*.json`、`acceptance-consistency`、accept 报告 consistency 摘要
- 退出码策略：允许 `0/1/3`，拒绝其他码

验收：

```bash
scripts/kylin-offline-validate.sh --bin ./.tmp/opskit-local --output ./.tmp/offline-validate --clean
```

### B. 文档判读口径统一

- 离线部署文档统一为“真实执行回归”（非 dry-run）
- 明确 `status=1` 的含义（存在 FAIL，不等于程序不可用）
- `GETSTART` 与根 README 增加离线回归入口
- 新增 `opskit status --json` 机器可读输出

验收：

```bash
rg -n "kylin-offline-validate|0/1/3|status=1" README.md README.zh-CN.md docs/getting-started/*.md
GOCACHE=$PWD/.gocache go run ./cmd/opskit status --output ./.tmp/offline-validate --json
```

### C. 发布流程衔接

- release notes 与 changelog 对齐到 `v0.3.5-preview.1`
- GitHub release 指引包含离线回归脚本入口

验收：

```bash
scripts/release-check.sh
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- 一键离线回归脚本可跑通并生成验证通过总结
- 文档入口一致，不出现过时版本号

## 5. 建议时间盒

- Day 1：A
- Day 2：B
- Day 3：C + 回归 + 发布
