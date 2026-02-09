# demo-blackbox-middleware-manage

`demo-blackbox-middleware-manage.json` 是一个面向黑箱型中间件服务器的 Manage 模板。
适用于“只能做运行态巡检与证据采集，不做部署变更”的场景，例如永中 FCS、金蝶中间件等。

- Delivery Level: `Demo`

## 适用范围

- 生命周期阶段：`A` / `B` / `C` / `D` / `E` / `F`
- 核心目标：系统就绪性检查、服务运行态检查、证据链输出
- 黑箱定位：不包含应用部署脚本，不依赖应用内部 API

## 变量说明

- `STACK_ID`：状态与证据命名前缀（例：`fcs-blackbox`、`kingdee-blackbox`）
- `SERVICE_NAME`：逻辑服务名
- `SERVICE_UNIT`：systemd unit 名
- `PROCESS_MATCH`：进程匹配关键字（用于 `process_args` 证据）
- `SERVICE_PORT`：服务监听端口
- `HEALTH_HOST`：DNS 解析检查主机（默认 `localhost`）
- `APP_DIR` / `DATA_DIR` / `LOG_DIR`：黑箱服务目录
- `INSTALL_ROOT` / `EVIDENCE_DIR`：OpsKit 状态与证据输出目录
- `MAX_RESTARTS`：允许的最大重启次数阈值
- `ROOT_MOUNT`：根挂载检查目标（默认 `/`）
- `RECOVER_TRIGGER`：E 阶段恢复触发来源标签

推荐直接使用变量文件（模板名固定，变量文件可分预设）：

```bash
examples/vars/demo-blackbox-middleware-manage.json              # default 预设
examples/vars/demo-blackbox-middleware-manage-fcs.json          # FCS 预设
examples/vars/demo-blackbox-middleware-manage-kingdee.json      # 金蝶预设
```

## 预设对照矩阵（default / FCS / Kingdee）

| 变量 | default | FCS | Kingdee |
| --- | --- | --- | --- |
| `STACK_ID` | `blackbox-middleware` | `fcs-blackbox` | `kingdee-blackbox` |
| `SERVICE_NAME` | `blackbox-service` | `fcs-server` | `kingdee-middleware` |
| `SERVICE_UNIT` | `blackbox.service` | `fcs.service` | `kingdee.service` |
| `PROCESS_MATCH` | `blackbox-service` | `fcs` | `kingdee` |
| `SERVICE_PORT` | `18080` | `18080` | `18081` |
| `APP_DIR` | `/opt/blackbox` | `/opt/fcs` | `/opt/kingdee` |
| `DATA_DIR` | `/data/blackbox` | `/data/fcs` | `/data/kingdee` |
| `LOG_DIR` | `/logs/blackbox` | `/logs/fcs` | `/logs/kingdee` |
| `INSTALL_ROOT` | `/data/opskit-blackbox` | `/data/opskit-fcs` | `/data/opskit-kingdee` |
| `EVIDENCE_DIR` | `/data/opskit-blackbox/evidence` | `/data/opskit-fcs/evidence` | `/data/opskit-kingdee/evidence` |
| `RECOVER_TRIGGER` | `blackbox_monitor` | `fcs_blackbox_monitor` | `kingdee_blackbox_monitor` |

说明：

- 三套预设都使用同一模板：`assets/templates/demo-blackbox-middleware-manage.json`
- 若迁移到新黑箱服务，优先复制 `default` 预设再调整路径/端口/unit

## 最小变量集（手工接入）

至少提供以下变量即可跑通基础链路（A/D/F）：

- `STACK_ID`
- `SERVICE_NAME`
- `SERVICE_UNIT`
- `PROCESS_MATCH`
- `SERVICE_PORT`
- `APP_DIR`
- `DATA_DIR`
- `LOG_DIR`
- `INSTALL_ROOT`
- `EVIDENCE_DIR`

可选但建议保留：

- `HEALTH_HOST`（默认 `localhost`）
- `MAX_RESTARTS`（默认 `3`）
- `ROOT_MOUNT`（默认 `/`）
- `RECOVER_TRIGGER`（默认 `blackbox_monitor`）

## 模板校验

文本模式：

```bash
./opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage.json assets/templates/demo-blackbox-middleware-manage.json
./opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json assets/templates/demo-blackbox-middleware-manage.json
./opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json assets/templates/demo-blackbox-middleware-manage.json
```

机器可读模式（脚本/CI 推荐）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-blackbox-middleware-manage.json assets/templates/demo-blackbox-middleware-manage.json
./opskit template validate --json --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json assets/templates/demo-blackbox-middleware-manage.json
./opskit template validate --json --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json assets/templates/demo-blackbox-middleware-manage.json
```

## 执行示例（可选）

```bash
# default 示例
./opskit run A --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox-default
./opskit run D --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox-default
./opskit accept --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox-default
./opskit status --output ./.tmp/opskit-blackbox-default

# FCS 示例
./opskit run A --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json --output ./.tmp/opskit-blackbox
./opskit run D --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json --output ./.tmp/opskit-blackbox
./opskit accept --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-fcs.json --output ./.tmp/opskit-blackbox
./opskit status --output ./.tmp/opskit-blackbox

# 金蝶示例
./opskit run A --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json --output ./.tmp/opskit-kingdee-blackbox
./opskit run D --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json --output ./.tmp/opskit-kingdee-blackbox
./opskit accept --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage-kingdee.json --output ./.tmp/opskit-kingdee-blackbox
./opskit status --output ./.tmp/opskit-kingdee-blackbox
```

说明：

- E 阶段 `recover_sequence` 默认 `enabled=false`，避免误触发 `systemctl start`。
- 需要演练恢复时，可使用 `--fix` 包含禁用步骤。

## 交付门禁信息

### 接管职责（一句话）

该模板接管“黑箱业务系统（FCS/金蝶等）的单机运行态监管与验收输出”职责（A/D/F）。

### 模板不做什么

- 不负责厂商安装流程与制品部署
- 不修改业务系统配置语义
- 不假设可访问外部控制平面

### 单机自洽前提

- 目标服务已由外部流程部署完成
- 本机可执行基础检查命令，`--output` 可写
- 不依赖其他节点状态

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox
./opskit run D --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox
./opskit accept --template assets/templates/demo-blackbox-middleware-manage.json --vars-file examples/vars/demo-blackbox-middleware-manage.json --output ./.tmp/opskit-blackbox
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-blackbox-middleware-manage/vars.example.yaml`
- FCS/金蝶差异通过 `unit/port/path/process_match` 表达，不通过 vars 决定执行流程

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-blackbox-middleware-manage.json --vars "STACK_ID=" assets/templates/demo-blackbox-middleware-manage.json
```

### 失败可交付说明

即便服务异常或检查失败，仍应产出 `state/*.json` 与报告/证据索引，支持黑箱系统问题定位和交付复核。

## 预期输出

- `state/overall.json`
- `state/lifecycle.json`
- `state/artifacts.json`
- `evidence/<stack>-state-dir-hash.json`
- `evidence/<stack>-overall-json-hash.json`
- `evidence/<stack>-process-args.json`
- `evidence/<stack>-systemd-status.json`

## 常见失败

- `systemd_unit_active` 失败：服务未启动或 unit 名错误
- `port_listening` 失败：端口未监听或端口配置不一致
- `systemd_restart_count` WARN：重启频次超过阈值
- `fs_readonly` 失败：目标挂载为只读
- `template validate --json` 返回 `template_var_type_mismatch`：变量类型与 schema 不一致
