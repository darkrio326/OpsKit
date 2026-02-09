# OpsKit（中文版）

OpsKit 是一款**面向离线/信创环境的服务器全生命周期运维与验收工具**，覆盖 A-F 阶段：Preflight / Baseline / Deploy / Operate / Recover / Accept-Handover。

当前版本：`v0.4.2-preview.1`（M4 自部署模板扩展预览版）

## 当前能力（Milestone 3）

- 通用巡检能力：A / D 阶段可独立执行，支持状态汇总与问题分级
- 可复核证据包：`accept` 生成 manifest + hashes + reports + snapshots
- UI 状态页：读取 `state/*.json` 展示 overall、阶段状态与证据入口
- 模板驱动：支持模板与变量渲染（内置 + 外部）
- 模板变量校验：必填/类型/枚举/默认值/示例值/变量分组
- 统一执行器：`executil` 作为唯一外部命令执行与审计入口
- 并发安全：全局锁防并发执行（冲突返回退出码 `4`）

## 当前不包含/不承诺

- 生产级中间件一键部署模板（当前仅提供 demo 模板）
- 客户定制模板、客户环境适配与驻场脚本
- 登录/权限系统（RBAC、账号体系）
- 多节点编排与分布式协调

## 三种使用模式

- 无模板：临时接管 / 排障 / 补验收；不承诺按模板交付
- 标准模板（推荐）：按服务器用途选择内置/标准模板，执行 `A -> D -> Accept` 输出可复核结果
- 定制模板（高级）：面向特殊项目，必须通过 `docs/product-design/09-模板设计指南.md` 与 `docs/product-design/DELIVERY_GATE.md` 门禁

## 快速开始（最短跑通）

1. 构建发布二进制（Linux）：

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/opskit-linux-arm64 ./cmd/opskit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/opskit-linux-amd64 ./cmd/opskit
```

2. 本地最小链路（A / D / Accept）：

```bash
go build -o opskit ./cmd/opskit
./opskit run A --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit run D --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit accept --template templates/builtin/default-manage.json --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo --json
```

模板接入前建议做机器可读校验：

```bash
./opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
```

3. 启动 UI 查看状态：

```bash
./opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080 --status-interval 15s
```

浏览器访问：`http://127.0.0.1:18080`

参考文档：

- `docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- `docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md`
- `docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`
- `scripts/kylin-offline-validate.sh`
- `scripts/generic-readiness-check.sh`

## 可选验证（银河麒麟 V10 Docker）

```bash
make -C examples/generic-manage docker-kylin-e2e
```

## 目录结构（简述）

```text
assets/templates/          # Demo 模板（去生产化）
docs/                      # 设计、规范、发布文档
internal/                  # 核心实现（engine/state/plugins/...）
cmd/opskit/                # CLI 主程序入口
templates/builtin/         # 内置模板
web/ui/                    # 开发态 UI 资源
```

## 路线图（Milestone 4-6）

- Milestone 4：模板库扩展（ELK 等示例模板、模板验收规范）
- Milestone 5：Recover/Operate 深化（策略化恢复、更多通用检查）
- Milestone 6：交付能力增强（多格式 handover、模板仓库化、多实例预研）

详见：`ROADMAP.md` 与 `docs/architecture/ROADMAP.md`。

## 文档与发布入口

- 规格与安全：`docs/specs/README.md`
- 产品设计（整理版）：`docs/product-design/README.md`
- 模板交付门禁：`docs/product-design/DELIVERY_GATE.md`
- GitHub 发布说明：`docs/GITHUB_RELEASE.md`
- 版本规划规范：`docs/RELEASE_PLANNING_GUIDE.md`
- 当前稳定版发布说明：`docs/releases/notes/RELEASE_NOTES_v0.3.7.md`
- 当前预览版发布说明：`docs/RELEASE_NOTES_v0.4.2-preview.1.md`
- 当前预览版发布计划：`docs/RELEASE_PLAN_v0.4.2-preview.1.md`
- 版本变更记录：`CHANGELOG.md`
- 安全边界：`SECURITY.md`
- 开源许可证：`LICENSE`（Apache-2.0）
- 英文 README：`README.md`

## 发布前门禁（v0.4.2-preview.1）

常规门禁：

```bash
scripts/template-validate-check.sh --clean
scripts/release-check.sh --with-offline-validate
scripts/template-delivery-check.sh --clean
```

真实服务器验证前建议：

```bash
scripts/generic-readiness-check.sh --clean
```

附加 `release-check summary.json` 契约检查：

```bash
scripts/generic-readiness-check.sh --with-release-json-contract --clean
```

严格模式（通用链路 + 离线回归均要求全 0 退出码）：

```bash
scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
```

离线严格门禁（离线回归阶段退出码必须全为 `0`）：

```bash
scripts/release-check.sh --with-offline-validate --offline-strict-exit
```

构建发布资产：

```bash
scripts/release.sh --version v0.4.2-preview.1 --clean
```

发布资产包含：Linux 双架构二进制、`checksums.txt`、`release-metadata.json`。

## 免责声明

当前版本（`v0.4.2-preview.1`）用于**内网/离线场景的通用能力验证与验收演练**。用于生产前，请完成安全性、稳定性与合规性评估。
