# OpsKit v0.3.4-preview.1 发布任务清单

> 主线：把 0.3.3 新增的状态与证据能力真正“看得见、用得上”，并继续加固跨环境可运行性。

## 1. 版本目标

- UI 展示补齐：阶段 summary 与 acceptance consistency 可视化
- 状态/产物索引增强：一致性报告纳入 artifacts 可追踪
- 通用检查跨平台健壮性增强（优先修复已知降级误报）
- 发布脚本可观测性增强（输出关键统计与耗时）

## 2. 范围边界

### In Scope

- `web/ui` 与 `internal/installer/assets` 的状态展示增强
- `internal/stages` / `internal/state` 的 artifacts 索引补强
- `internal/plugins/checks` 的兼容性改进与测试补齐
- `docs/getting-started` 与 release 文档同步

### Out of Scope

- 生产级 deploy 模板
- 多节点编排
- 权限系统与账号体系

## 3. 任务拆分（每项可验收）

### A. UI 展示增强

- 阶段卡片展示 `summary(total/pass/warn/fail/skip)`
- artifacts 区域增加 latest acceptance consistency 入口
- Recover/Accept 关键指标显示（reason code / consistency）
- 当前进度：已完成（web/ui + installer assets 同步）

验收：

```bash
go run ./cmd/opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080
```

### B. Acceptance 一致性索引增强

- 将 `acceptance-consistency-*.json` 纳入 artifacts 索引
- accept 报告中增加 consistency 摘要块
- 增加跨文件一致性回归测试（state/artifacts/report）

验收：

```bash
GOCACHE=$PWD/.gocache go test ./internal/stages ./internal/state ./internal/reporting
```

### C. 通用检查兼容性加固

- 修复/优化 `load_average` 在非 Linux 环境解析策略
- 为 `dns_resolve` 增加参数开关（可选跳过受限网络调用）
- 增加 check 级降级原因标准化指标

验收：

```bash
GOCACHE=$PWD/.gocache go test ./internal/plugins/checks
```

### D. 发布门禁增强

- `release-check` 输出阶段统计与耗时
- 文档增加“快速判定发布可用”的读法

验收：

```bash
scripts/release-check.sh
scripts/release.sh --version v0.3.4-preview.1 --clean
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- demo 模板 `run A/D/accept` 与 `status` 可运行
- UI 可看到 summary 与 acceptance consistency
- 文档与 changelog 同步，无敏感信息

## 5. 建议时间盒

- Day 1：A + B
- Day 2：C
- Day 3：D + 回归 + 发布
