# demo-hello-service

`demo-hello-service.json` 是一个去生产化的轻量服务模板。
该模板用于演示“变量渲染 + 部署占位动作 + 证据输出”流程，不包含真实中间件部署。

- Delivery Level: `Demo`

## 适用范围

- 生命周期阶段：`B` / `C` / `D` / `F`
- 核心能力：目录准备、模板渲染、清单声明、证据哈希
- 占位检查：`systemd` 与端口监听检查默认关闭，避免误操作真实服务

## 变量说明

- `SERVICE_NAME`（`group=service`，默认 `hello-service`）：服务名
- `SERVICE_PORT`（`group=service`，默认 `18080`）：服务监听端口
- `SERVICE_UNIT`（`group=runtime`，默认 `hello-service.service`）：systemd unit 名（占位）
- `PROCESS_MATCH`（`group=runtime`，默认 `hello-service`）：进程匹配关键字
- `INSTALL_ROOT`（`group=paths`，必填）：输出根目录
- `CONF_DIR`（`group=paths`，必填）：配置目录根路径

推荐使用示例变量文件：

```bash
examples/vars/demo-hello-service.json
```

## 模板校验

文本模式：

```bash
./opskit template validate --vars-file examples/vars/demo-hello-service.json assets/templates/demo-hello-service.json
```

机器可读模式（脚本/CI 推荐）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-hello-service.json assets/templates/demo-hello-service.json
```

## 执行示例（可选）

```bash
./opskit run B --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit run C --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit run D --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit accept --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello
```

覆盖变量示例：

```bash
./opskit run C \
  --template assets/templates/demo-hello-service.json \
  --output ./.tmp/opskit-hello \
  --vars "SERVICE_NAME=demo-hello,SERVICE_PORT=19090,PROCESS_MATCH=demo-hello"
```

## 预期输出

- `conf/<service>/hello-demo.env`
- `state/stack.json`
- `state/artifacts.json`
- `evidence/<service>-env-hash.json`
- `evidence/<service>-conf-dir-hash.json`

## 常见失败

- `template validate` 报缺少变量：补齐 `INSTALL_ROOT` / `CONF_DIR`，或使用 `--vars-file`
- `run C` 写文件失败：检查 `INSTALL_ROOT`、`CONF_DIR` 是否可写
- `run D` 无 systemd/端口结果：该模板默认禁用占位检查，属于预期行为
- `template validate --json` 返回 `template_var_type_mismatch`：变量类型与 schema 不一致（如 `array/object` 需合法 JSON）

## 交付门禁信息

### 接管职责（一句话）

该模板接管“单机轻量服务的模板链路演示与验收证据输出”职责（A/B/C/D/F）。

### 模板不做什么

- 不提供生产级中间件部署策略
- 不承诺高可用或多节点治理
- 不替代真实业务发布流程

### 单机自洽前提

- 本机可写 `INSTALL_ROOT` 与 `CONF_DIR`
- 不依赖外部系统、集群控制面或公网

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello-service
./opskit run D --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello-service
./opskit accept --template assets/templates/demo-hello-service.json --vars-file examples/vars/demo-hello-service.json --output ./.tmp/opskit-hello-service
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-hello-service/vars.example.yaml`
- 变量仅覆盖 service name/port/unit/path，不表达流程选择

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-hello-service.json --vars "INSTALL_ROOT=,CONF_DIR=" assets/templates/demo-hello-service.json
```

### 失败可交付说明

即便服务未启动或检查失败，仍应输出状态 JSON 与可复核证据入口，用于后续接管。
