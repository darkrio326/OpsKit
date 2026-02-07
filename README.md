# OpsKit

一句话定位：**面向离线/信创环境的服务器全生命周期运维与验收工具**（A~F 阶段：Preflight/Baseline/Deploy/Operate/Recover/Accept-Handover）。

当前版本：`v0.3.5-preview.1`

## 当前能力（Milestone 3）

- 通用巡检能力：A/D 阶段可独立运行，支持状态汇总与问题分级
- 可复核证据包：`accept` 生成 manifest + hashes + reports + snapshots
- UI 状态页：读取 `state/*.json` 展示 overall、阶段状态与产物入口
- 模板驱动：支持模板 + 变量渲染（内置与外部模板）
- 模板变量校验：必填/类型/枚举/默认值
- 统一执行器：`executil` 统一外部命令执行与审计入口
- 并发安全：全局锁防并发执行冲突（冲突返回退出码 `4`）

## 当前不包含/不承诺

- 生产级中间件一键部署模板（仅提供 demo 模板）
- 客户定制模板、客户环境适配与驻场脚本
- 登录/权限系统（账号体系、RBAC）
- 多节点集群编排与分布式协调

## Quick Start（最短跑通）

1) 构建二进制（Linux）

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/opskit-linux-arm64 ./cmd/opskit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/opskit-linux-amd64 ./cmd/opskit
```

2) 本地最小执行链路（A / D / Accept）

```bash
go build -o opskit ./cmd/opskit
./opskit run A --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit run D --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit accept --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo --json
```

3) 启动 UI 查看状态

```bash
./opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080
```

浏览器访问：`http://127.0.0.1:18080`

4) Docker 回归（银河麒麟 V10）

```bash
make -C examples/generic-manage docker-kylin-e2e
```

详见：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
麒麟离线部署：`docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md`
麒麟离线回归：`docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`
离线一键回归脚本：`scripts/kylin-offline-validate.sh`
真实服务器前统一门禁：`scripts/generic-readiness-check.sh`

## 目录结构（简述）

```text
assets/templates/          # demo 模板（去生产化）
docs/                      # 设计、规范、发布文档
internal/                  # 核心实现（engine/state/plugins/...）
cmd/opskit/                # CLI 主程序入口
templates/builtin/         # 内置模板
web/ui/                    # 开发态 UI 资源
```

## Roadmap（Milestone 4~6）

- Milestone 4：模板库增强（ELK 等示例模板、模板验收规范）
- Milestone 5：Recover/Operate 能力深化（策略化恢复、更多通用检查）
- Milestone 6：交付增强（handover 多格式、模板仓库化、多实例预研）

详见：`ROADMAP.md` 与 `docs/architecture/ROADMAP.md`

## 发布与文档入口

- 规格与安全：`docs/specs/README.md`
- 产品设计（整理版）：`docs/product-design/README.md`
- GitHub 发布说明：`docs/GITHUB_RELEASE.md`
- 下一版发布计划：`docs/RELEASE_PLAN_v0.3.6-preview.1.md`
- 版本变更记录：`CHANGELOG.md`
- 安全边界：`SECURITY.md`
- 开源许可证：`LICENSE`（Apache-2.0）
- 中文扩展说明：`README.zh-CN.md`

## 发布前门禁（v0.3.6 预演）

常规门禁（建议默认）：

```bash
scripts/release-check.sh --with-offline-validate
```

进入真实服务器前（建议）：

```bash
scripts/generic-readiness-check.sh --clean
```

严格模式（通用链路 + 离线回归都要求 exit=0）：

```bash
scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
```

严格门禁（要求离线回归阶段 exit 全为 0）：

```bash
scripts/release-check.sh --with-offline-validate --offline-strict-exit
```

构建发布资产：

```bash
scripts/release.sh --version v0.3.6-preview.1 --clean
```

发布资产会包含：二进制（amd64/arm64）、`checksums.txt`、`release-metadata.json`。

## 免责声明

本项目当前版本（`v0.3.x Preview`）主要用于**内网/离线场景验证**与能力演示。生产环境使用前请完成安全、稳定性与合规评估。
