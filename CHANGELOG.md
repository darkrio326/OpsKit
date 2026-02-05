# Changelog

本项目变更记录遵循“按版本可追溯”原则，发布节奏以 `v0.x Preview` 为主。

## [Unreleased]

### Added

- 新增 `docs/RELEASE_PLAN_v0.3.1-preview.2.md`（下一版可执行任务清单）
- 新增 `docs/RELEASE_NOTES_v0.3.1-preview.2.md`（下一版发布文案草稿）

### Changed

- `README.md` 增加“下一版发布计划”入口
- `docs/GITHUB_RELEASE.md` 增加 preview.2 任务单与 release notes 链接

### Fixed

- 待补充

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
