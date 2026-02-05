# OpsKit（中文版）

OpsKit 是一个基于 Go 的运维生命周期工具，围绕 `A -> F` 阶段执行：

- A Preflight（前置检查）
- B Baseline（基线）
- C Deploy（部署）
- D Operate（巡检）
- E Recover（恢复）
- F Accept/Handover（验收与移交）

目标是用一个二进制和统一 JSON 状态输出，把“可执行、可观测、可留证据”的链路跑通。

快速上手：`docs/getting-started/GETSTART.md`
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
./opskit install [--template id|path] [--vars k=v] [--dry-run] [--output dir]
./opskit run <A|B|C|D|E|F|AF> [--template id|path] [--vars k=v] [--dry-run] [--output dir]
./opskit status [--output dir]
./opskit accept [--template id|path] [--vars k=v] [--dry-run] [--output dir]
./opskit handover [--output dir]
./opskit web [--output dir] [--listen :18080]
./opskit template validate <file>
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
- 插件机制：checks/actions/evidence 统一注册入口
- 最小 UI：静态页面读取 JSON，展示 A~F 状态与产物入口

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
