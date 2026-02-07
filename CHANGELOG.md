# Changelog

本项目变更记录遵循“按版本可追溯”原则，发布节奏以 `v0.x Preview` 为主。

## [Unreleased]

### Added

- 新增 `docs/RELEASE_PLAN_v0.3.6-preview.1.md`（下一版扩展任务清单）
- 新增 `docs/RELEASE_NOTES_v0.3.6-preview.1.md`（下一版发布文案草稿）
- `scripts/kylin-offline-validate.sh` 新增 `--summary-json-file`（结构化回归摘要）
- 新增检查插件：`fs_readonly`、`disk_inodes`、`clock_skew`
- `opskit status --json` 新增 `health` 字段（`ok|warn|fail`）
- `scripts/release.sh` 新增 `release-metadata.json`（`gitCommit/buildTime/goVersion`）
- `examples/generic-manage/run-af.sh` 新增可门禁模式与 `status.json` 固定输出
- 通用自检 `summary.json` 新增 `result/reasonCode/recommendedAction`
- 新增 `scripts/generic-readiness-check.sh`（真实服务器前统一 Go/No-Go 门禁）
- 新增 `docs/RELEASE_PLAN_v0.3.4-preview.1.md`（下一版可执行任务清单）
- 新增 `docs/RELEASE_NOTES_v0.3.4-preview.1.md`（下一版发布文案草稿）
- 新增 `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`（麒麟离线用户侧回归清单）
- 新增 `scripts/kylin-offline-validate.sh`（麒麟离线机一键回归脚本）
- 新增 `opskit status --json`（机器可读状态输出）
- 新增 `docs/specs/SPEC_STATUS_JSON.md`（status JSON 输出契约）
- 离线回归脚本支持 `--json-status-file`（落盘 status JSON）
- 离线回归脚本支持 `--strict-exit`（可选严格通过模式）
- `scripts/release-check.sh` 支持 `--with-offline-validate` 可选门禁
- `scripts/release-check.sh` 支持 `--offline-strict-exit` 透传
- UI 新增阶段 summary 展示（`total/pass/warn/fail/skip`）
- UI artifacts 高亮新增 latest acceptance consistency 入口
- Accept 阶段新增 `acceptance-consistency` 报告索引写入 `artifacts.json`
- 新增回归测试：accept consistency 索引、跨文件一致性、`load_average` 跨平台回退、`dns_resolve` 跳网参数

### Changed

- `README.md` / `README.zh-CN.md` 下一版计划入口滚动到 `v0.3.6-preview.1`
- `ROADMAP.md` 最新发布/进行中版本滚动到 `v0.3.5` / `v0.3.6-preview.1`
- `docs/GITHUB_RELEASE.md` 发布文案模板与命令滚动到 `v0.3.6-preview.1`
- `scripts/release-check.sh` 新增 `--offline-summary-json-file` 透传
- `scripts/release-check.sh` 汇总新增 `recommended action`（continue_release/block_release）
- 离线回归文档补充 `summary.json` 产物与校验口径
- 离线部署/验证文档补充 `status.json health` 与退出码映射口径
- 离线回归与发布文档明确 strict 策略：默认 non-strict 放行，基线治理后再做 strict 全绿校验
- `assets/templates/demo-server-audit.json` D 阶段接入 inode/只读挂载/时钟偏移检查
- `scripts/release.sh` 自动生成并校验 `checksums.txt`
- `scripts/kylin-offline-validate.sh` 默认输出目录改为版本无关路径（`/data/opskit-regression`）
- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md` 示例版本与路径口径同步到 `v0.3.6`/`v036`
- demo 模板 README 补充“最低可运行变量集”与“常见失败”
- `docs/examples/generic-manage/README*.md` 补充 strict 模式与 `status.json` 产物说明
- `README.md` / `README.zh-CN.md` / `docs/getting-started/GETSTART.md` / `docs/GITHUB_RELEASE.md` 补充统一门禁脚本入口
- `README.md` / `README.zh-CN.md` 当前版本标记同步到 `v0.3.5-preview.1`
- `README.md` / `README.zh-CN.md` / `ROADMAP.md` / `docs/GITHUB_RELEASE.md` 滚动到 `v0.3.3-preview.1`
- `README.md` / `README.zh-CN.md` / `docs/getting-started/GETSTART.md` 增加麒麟离线部署与回归入口
- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md` 回归口径从 dry-run 调整为真实执行，并补充退出码判读
- `README.md` / `README.zh-CN.md` / `docs/GITHUB_RELEASE.md` 补充 `v0.3.5` 常规与严格门禁发版命令
- `scripts/release-check.sh` 新增步骤耗时与总耗时汇总输出
- `dns_resolve` 新增 `skip_network_query` 参数（受限网络可跳过外部查询）
- check 降级结果统一输出指标：`check_degraded` / `check_degraded_reason`
- `load_average` 解析链路增强（`/proc/loadavg` -> `uptime` -> `sysctl vm.loadavg`）

### Fixed

- 修复 acceptance consistency 结果未纳入 artifacts 索引的问题
- 修复 `load_average` 在非 Linux 环境容易误降级的问题
- 修复 `dns_resolve` 在受限网络场景误报问题

## [v0.3.3-preview.1] - 2026-02-07

### Added

- 新增 `scripts/release-check.sh`（发布前最小回归脚本）
- Recover 结果新增 `recover_reason_code` 指标
- Recover collect 新增 `source` 字段（command/journal/file）
- Recover collect 新增截断元数据（`originalLength`/`truncatedLength`）
- 新增检查插件：`ntp_sync`、`dns_resolve`、`systemd_restart_count`
- `lifecycle.json` 阶段新增 `summary(total/pass/warn/fail/skip)` 结构
- accept 新增一致性校验记录（`acceptance-consistency-*.json`）

### Changed

- Recover circuit 状态新增 `lastErrorCode`
- recover summary 新增 `lastReasonCode`
- demo 审计模板 D 阶段新增 `ntp_sync`、`dns_resolve` 检查
- 引擎在阶段执行后统一写入 stage summary 计数
- accept 新增 `manifest <-> hashes <-> state` 一致性检查与指标输出
- 新增并完善 `v0.3.3-preview.1` 计划文档与发布说明

### Fixed

- 修复 recover 失败原因无法稳定聚合的问题（统一 reason code）
- 修复 accept 证据包一致性缺少显式记录的问题

## [v0.3.2] - 2026-02-06

### Added

- 新增 `--vars-file`（支持 JSON 或 key=value 文件）
- 变量类型扩展：`array/object`（JSON 形式）
- 新增 vars-file 解析与回归测试
- 新增 vars-file 示例（`examples/vars/demo-hello-service.*`）
- 新增 vars-file 示例（`examples/vars/demo-server-audit.*`）
- Recover collect 新增输出限流（`collect_output_limit`）与截断标记
- 新增 `docs/RELEASE_PLAN_v0.3.2-preview.1.md`
- 新增 `docs/RELEASE_NOTES_v0.3.2-preview.1.md`

### Changed

- demo 模板变量说明更新
- 模板校验错误提示路径化
- Recover collect 命令/日志输出统一结构（含 exit code）并默认脱敏

### Fixed

- 修复 `docs/RELEASE_PLAN_v0.3.2-preview.1.md` 中失效的 vars-file 示例路径
- 修复 Recover collect 在大输出场景下可能导致 JSON 过大的问题（默认限流 + 截断标记）

## [v0.3.1-preview.2] - 2026-02-06

### Added

- 证据包新增 `manifest.json`（与 `hashes.txt` 一致）
- 新增回归测试：CLI 退出码、模板校验、证据包一致性、脱敏
- 新增截图占位目录：`docs/assets/screenshots/`
- 新增 v0.3.1 preview 计划与 release notes 草稿

### Changed

- 模板 JSON 严格解析与校验增强（未知字段/未解析变量/非法 severity）
- 模板变量 schema 校验（必填/类型/枚举/默认值）
- 麒麟 V10 Docker e2e 文档与示例说明统一
- `README.zh-CN.md` 增加界面预览占位
- `docs/GITHUB_RELEASE.md` 更新 preview.2 任务单与 release notes 链接

## [v0.3.0-preview.1] - 2026-02-05

### Added

- A~F 生命周期框架可运行（通用路径）
- 统一状态输出：`overall/lifecycle/services/artifacts` JSON
- 静态 UI 状态页（读取 JSON 展示阶段与产物）
- 模板与变量体系（内置/外部模板加载与渲染）
- 统一执行器入口（executil）与全局锁并发控制
- Accept/Handover 证据链（manifest/hashes/reports/snapshots）
- GitHub 发布文档与 release notes 模板
- 去生产化 demo 模板：`assets/templates/demo-server-audit.json`、`assets/templates/demo-hello-service.json`

### Changed

- 文档体系重组到 `docs/`，新增 product-design 与 specs 合并版主线
- 根目录中文文档入口恢复为 `README.zh-CN.md`

### Security

- 补充开源发布边界：不发布客户信息、生产模板、真实日志/证据包
- 增加 `SECURITY.md` 说明默认安全边界与脱敏规则
