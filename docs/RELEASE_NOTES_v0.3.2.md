# OpsKit v0.3.2

> 版本主题：模板变量可用性与 Recover 输出稳健性增强。

## 本次发布重点

- 新增 `--vars-file`，支持文件式变量输入
- 模板变量类型扩展（`array/object`）
- 模板校验报错路径化，定位更直接
- Recover collect 输出限流与统一格式，避免证据 JSON 过大

## 主要变更

### Added

- CLI 新增 `--vars-file <path>`（JSON 或 key=value）
- 变量文件解析与回归测试
- 新增 vars-file 示例：
  - `examples/vars/demo-hello-service.json`
  - `examples/vars/demo-hello-service.env`
  - `examples/vars/demo-server-audit.json`
  - `examples/vars/demo-server-audit.env`
- Recover collect 支持 `collect_output_limit`

### Changed

- 模板校验错误提示统一为路径化输出（`template.<path>: <reason>`）
- demo 模板 README 统一为可直接复制执行的命令
- Recover collect 命令/日志输出统一格式：`[cmd exit=code]`

### Fixed

- 修复发布任务单中的失效 vars-file 示例路径
- 修复 Recover collect 大输出导致 JSON 过大风险（默认限流 + 截断标记）

## 发布前验证命令

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit template validate --vars-file ./examples/vars/demo-server-audit.json assets/templates/demo-server-audit.json
go run ./cmd/opskit template validate --vars-file ./examples/vars/demo-hello-service.env assets/templates/demo-hello-service.json
scripts/release.sh --version v0.3.2 --clean
```

## Release 资产

- `opskit-v0.3.2-linux-arm64`
- `opskit-v0.3.2-linux-amd64`
- `checksums.txt`
