# OpsKit v0.3.1-preview.2

> 预览版本（Preview）。聚焦“通用能力稳定性与可复核性”。

## 本次发布重点

- 退出码与错误分类一致性增强（`run/status/accept`）
- 模板校验增强（变量必填/类型/枚举/路径化报错）
- Accept 证据包一致性与脱敏回归加强
- 银河麒麟 v10 Docker 验证链路固化

## 兼容说明

- 不引入生产级部署能力变更
- demo 模板可继续用于离线验证与演示
- 仍不承诺生产 SLA

## 主要变更

### Added

- （待补充）

### Changed

- （待补充）

### Fixed

- （待补充）

## 验证命令（发布前）

```bash
GOCACHE=$PWD/.gocache go test ./...
go run ./cmd/opskit template validate assets/templates/demo-server-audit.json
go run ./cmd/opskit template validate assets/templates/demo-hello-service.json
scripts/release.sh --version v0.3.1-preview.2 --clean
```

## Release 资产

- `opskit-v0.3.1-preview.2-linux-arm64`
- `opskit-v0.3.1-preview.2-linux-amd64`
- `checksums.txt`

## 安全提示

- 仅发布框架、通用能力与 demo 模板
- 不发布生产 deploy 模板与客户环境信息
- UI 默认建议监听 `127.0.0.1`
