# OpsKit v0.3.2-preview.1

> 预览版本（Preview）。主线聚焦“模板变量系统与校验可用性”。

## 本次发布重点

- 新增 `--vars-file` 支持（文件式变量输入）
- 变量类型与枚举校验增强（包含数组/对象最小支持）
- 模板校验报错更友好（路径化提示）
- demo 模板与文档示例统一

## 兼容说明

- 仍为 Preview 版本，不承诺生产 SLA
- 不包含生产级 deploy 模板

## 主要变更

### Added

- `--vars-file` 支持（JSON 或 key=value 文件）
- 变量类型扩展：`array/object`（JSON 形式）
- 变量文件解析与回归测试
- 新增 vars-file 示例（JSON/ENV）
- Recover collect 输出限流（`collect_output_limit`）与截断标记

### Changed

- 模板校验错误提示更友好（变量未解析/类型不匹配）
- demo 模板变量说明更新
- Recover collect 命令/日志输出统一为 `[cmd exit=code]` 格式并默认脱敏

### Fixed

- （待补充）

## 验证命令（发布前）

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit template validate assets/templates/demo-server-audit.json
go run ./cmd/opskit template validate assets/templates/demo-hello-service.json
scripts/release.sh --version v0.3.2-preview.1 --clean
```

## Release 资产

- `opskit-v0.3.2-preview.1-linux-arm64`
- `opskit-v0.3.2-preview.1-linux-amd64`
- `checksums.txt`

## 安全提示

- 仅发布框架、通用能力与 demo 模板
- 不发布生产 deploy 模板与客户环境信息
- UI 默认建议监听 `127.0.0.1`
