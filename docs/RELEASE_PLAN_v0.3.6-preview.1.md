# OpsKit v0.3.6-preview.1 发布任务清单

> 主线：在保持“离线可跑通”的基础上，提升“结果可判读、失败可定位、发布可复核”的工程完成度。

> 进度：A/B/C/D 已完成并合入主干，E 收敛中（新增统一门禁脚本，待最终回归与发布）。

## 1. 版本目标（增加交付密度）

- 把回归结果从“脚本日志”提升为“结构化摘要 + 失败定位”
- 把 A/D/F 关键检查扩展到更通用的离线场景（磁盘/文件系统/时间）
- 把 release 门禁从“可跑”提升为“可比较、可追溯”
- 把文档入口统一为“离线机器拿包即可执行”

## 2. 范围边界

### In Scope

- `scripts/kylin-offline-validate.sh` 和 `scripts/release-check.sh` 的结果增强
- `internal/plugins/checks` 通用检查补充（仅通用、去生产化）
- `docs/getting-started` 与 `README*` 的离线使用链路统一
- `docs/specs` 契约补齐（状态与退出码判读）

### Out of Scope

- 生产级 deploy 模板和真实中间件编排
- 多节点调度/远程 agent
- 账号系统与权限模型扩展

## 3. 任务拆分（每项可验收）

### A. 离线回归摘要化（结果更可读）

- `kylin-offline-validate` 新增 machine-readable summary（总耗时、阶段状态、失败点）
- 失败时输出统一 reason code（便于 CI/脚本做分支判断）
- 支持 `--summary-json-file`（默认写到输出目录）

验收：

```bash
scripts/kylin-offline-validate.sh --bin ./.tmp/opskit-local --output ./.tmp/offline-validate --clean --summary-json-file ./.tmp/offline-validate/summary.json
```

### B. 通用检查补齐（D 阶段厚一层）

- 新增 `fs_readonly` 检查（关键挂载点是否只读异常）
- 新增 `disk_inodes` 检查（inode 使用率阈值）
- 新增 `clock_skew` 检查（本地时间偏移粗检，离线可运行）

验收：

```bash
go run ./cmd/opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
```

### C. 状态判读与退出码对齐

- `status --json` 增加 `health` 汇总字段（ok/warn/fail）
- 文档补齐 `exit=0/1/3/4` 的使用建议与自动化判读规则
- `release-check` 在 summary 中输出建议动作（继续发布/阻断发布）

验收：

```bash
go run ./cmd/opskit status --output ./.tmp/opskit-demo --json
scripts/release-check.sh --with-offline-validate
```

### D. 发布资产可复核（产物闭环）

- `release.sh` 增加 release metadata（git commit、build time、go version）
- 自动校验 `checksums.txt` 与二进制资产一致性
- 文档补齐“离线下载包 -> 校验 -> 执行 -> 验收”完整路径

验收：

```bash
scripts/release.sh --version v0.3.6-preview.1 --clean
scripts/release-check.sh --with-offline-validate --offline-strict-exit
```

### E. 文档与示例收敛（降低上手门槛）

- `GETSTART`、`KYLIN_V10_OFFLINE_RELEASE`、`GITHUB_RELEASE` 链接与版本号同步
- demo 模板 README 补齐“最低可运行变量集”与“常见失败”
- `CHANGELOG` 按 `Added/Changed/Fixed` 保持可追溯
- 新增统一门禁脚本：`scripts/generic-readiness-check.sh`（用于真实服务器前 Go/No-Go）

验收：

```bash
rg -n "v0.3.6-preview.1|offline-validate|release-check|generic-readiness-check|status --json" README.md README.zh-CN.md docs/**/*.md
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `scripts/release-check.sh --with-offline-validate` 通过
- 严格模式可选通过：`--offline-strict-exit`
- demo 模板 `run A/D/accept` 与 `status --json` 跑通
- 发布文档、release notes、changelog 无版本漂移

### 4.1 strict 使用策略（本版决策）

- 默认放行口径：`0/1/3`（non-strict），用于确认通用链路可运行、产物可复核。
- `strict` 作为可选门禁，不作为离线首轮回归阻断条件。
- 在标准服务器完成基线治理（挂载/服务/时间同步等）后，补跑 strict 并以“全绿”为目标。

## 5. 建议时间盒

- Day 1：A + C
- Day 2：B
- Day 3：D
- Day 4：E + 回归 + 发布
