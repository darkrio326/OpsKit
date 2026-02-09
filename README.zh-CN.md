# OpsKit（中文版）

OpsKit 是一个基于 Go 的运维生命周期工具，围绕 `A -> F` 阶段执行：

- A Preflight（前置检查）
- B Baseline（基线）
- C Deploy（部署）
- D Operate（巡检）
- E Recover（恢复）
- F Accept/Handover（验收与移交）

目标是用一个二进制和统一 JSON 状态输出，把“可执行、可观测、可留证据”的链路跑通。

当前版本：`v0.3.7`（模板接入前置增强版，Milestone 4 预备）

快速上手：`docs/getting-started/GETSTART.md`
麒麟离线部署：`docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md`
麒麟离线回归：`docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`
离线一键回归脚本：`scripts/kylin-offline-validate.sh`
真实服务器前统一门禁：`scripts/generic-readiness-check.sh`
产品设计总览：`docs/product-design/README.md`

## 快速开始

```bash
go build -o opskit ./cmd/opskit
./opskit template validate templates/builtin/default-manage.json
./opskit install --template generic-manage-v1 --dry-run --no-systemd --output ./.tmp/opskit-demo
./opskit run AF --template generic-manage-v1 --dry-run --output ./.tmp/opskit-demo
```

## 常用命令

```bash
./opskit install [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--output dir]
./opskit run <A|B|C|D|E|F|AF> [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--output dir]
./opskit status [--output dir]
./opskit status [--output dir] [--json]
./opskit accept [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--output dir]
./opskit handover [--output dir]
./opskit web [--output dir] [--listen :18080] [--status-interval 15s]
./opskit template validate <file>
./opskit template validate --json <file>
```

说明：`web --status-interval` 会在后台定时刷新 `state/*.json`，UI 前端自动刷新读取的是这些更新后的状态。

## vars-file 示例

```bash
./opskit run C \
  --template assets/templates/demo-hello-service.json \
  --vars-file examples/vars/demo-hello-service.json \
  --output ./.tmp/opskit-hello
```

```bash
./opskit run A \
  --template assets/templates/demo-server-audit.json \
  --vars-file examples/vars/demo-server-audit.env \
  --output ./.tmp/opskit-demo
```

## 退出码

- `0`：成功
- `1`：失败
- `2`：前置条件不满足
- `3`：部分成功（存在 WARN）
- `4`：需人工介入（例如全局锁冲突）

## 关键能力（当前）

- 全局锁：并发执行冲突时返回退出码 `4`
- 统一 state JSON：`overall/lifecycle/services/artifacts`
- 原子写：所有状态文件以原子方式落盘
- 模板机制：支持模板加载、变量渲染、执行计划构建
- 模板变量校验：必填/类型/枚举/默认值/示例值/变量分组
- 插件机制：checks/actions/evidence 统一注册入口
- 最小 UI：静态页面读取 JSON，展示 A~F 状态与产物入口

## 界面预览（建议补充截图）

为便于开源读者快速理解，可以补充以下截图（建议放在 `docs/assets/screenshots/`）：

- 状态总览页（overall + A~F 概览）
- 阶段详情页（某一阶段的 checks/actions/evidence）
- 证据包列表页（reports/bundles 入口）

截图占位（稍后补图即可自动生效）：

![OpsKit UI 总览](docs/assets/screenshots/overview.png)
![OpsKit 阶段详情](docs/assets/screenshots/stage-detail.png)
![OpsKit 证据包列表](docs/assets/screenshots/artifacts.png)

## 目录说明（核心）

- `cmd/opskit`：CLI 入口
- `internal/core`：time/fs/exec/log/errors/exitcode/lock 等通用能力
- `internal/schema`：模板与 state 结构、枚举与校验
- `internal/engine`：Template -> Plan -> Stage 执行编排
- `internal/plugins`：checks/actions/evidence 插件实现与注册
- `internal/state`：state 读写、汇总与保留策略
- `web/ui`：开发态静态 UI
- `internal/installer/assets`：安装落盘 UI 资源
- `templates/builtin`：内置模板

## 在银河麒麟 V10 Docker 中验证（推荐回归入口）

```bash
make -C examples/generic-manage docker-kylin-e2e
```

可选参数示例：

```bash
IMAGE=kylinv10/kylin:b09 DOCKER_PLATFORM=linux/amd64 DRY_RUN=1 make -C examples/generic-manage docker-kylin-e2e
```

详细部署与验证说明见：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`

版本变更记录见：`CHANGELOG.md`
版本规划规范：`docs/RELEASE_PLANNING_GUIDE.md`
当前稳定版发布说明：`docs/releases/notes/RELEASE_NOTES_v0.3.7.md`
下一版发布说明：`docs/RELEASE_NOTES_v0.4.0-preview.4.md`
下一版发布计划：`docs/RELEASE_PLAN_v0.4.0-preview.4.md`

## 发布前门禁（v0.3.7）

常规门禁（推荐）：

```bash
scripts/template-validate-check.sh --clean
scripts/release-check.sh --with-offline-validate
```

进入真实服务器前（建议）：

```bash
scripts/generic-readiness-check.sh --clean
```

如需附加 `release-check summary.json` 契约门禁：

```bash
scripts/generic-readiness-check.sh --with-release-json-contract --clean
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
scripts/release.sh --version v0.3.7 --clean
```

发布资产包含：双架构二进制、`checksums.txt`、`release-metadata.json`。
