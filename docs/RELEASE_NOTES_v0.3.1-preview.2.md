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

- 证据包新增 `manifest.json`（与 `hashes.txt` 一致，可复核）
- Accept/Handover/Recover bundle 写入基础元信息（stage/template/bundle）
- 新增回归测试：CLI 退出码、模板校验、证据包一致性、脱敏

### Changed

- 模板校验增强：未解析变量与 severity 枚举直接报错
- 模板变量 schema 校验：必填/类型/枚举/默认值处理
- 模板 JSON 严格解析：未知字段直接失败
- 麒麟 V10 Docker e2e 文档与示例说明统一

### Fixed

- 暂无

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
