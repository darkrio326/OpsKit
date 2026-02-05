# 08 CLI、安装与容器验证

## 8.1 CLI 命令集（v1）

- `opskit install`
- `opskit run <A|B|C|D|E|F|AF>`
- `opskit status`
- `opskit accept`
- `opskit handover`
- `opskit template validate <file>`
- `opskit web`

常见参数：

- `--template`
- `--vars`
- `--dry-run`
- `--fix`
- `--output`

## 8.2 install 行为基线

`install` 至少完成：

- 目录初始化（state/reports/evidence/bundles/cache/ui）
- 静态页面落盘
- 可选 systemd 单元安装与激活
- 首次状态引导（如 A + D）

## 8.3 通用自检流程

建议每次模板开发前先跑：

`template validate -> install -> run AF -> status -> accept -> handover`

脚本入口：

- `examples/generic-manage/run-af.sh`

## 8.4 麒麟 V10 容器验证

标准入口：

- `make -C examples/generic-manage docker-kylin-e2e`
- 或 `examples/generic-manage/run-af-kylin-v10-docker.sh`

验证产物：

- `state/*.json`
- `reports/*.html`
- `bundles/*.tar.gz`
- `summary.json`
- `ui/index.html`

说明：

- 容器内部分 systemd 检查可能为 WARN，属预期现象
