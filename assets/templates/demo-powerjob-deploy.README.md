# demo-powerjob-deploy

`demo-powerjob-deploy.json` 是 PowerJob 自行部署示例模板（`mode=deploy`）。
单模板覆盖两个服务：

- `powerjob-server`
- `powerjob-worker`

- Delivery Level: `Demo`

## 阶段职责（A/B/C/D/F）

- A：环境预检（系统、挂载、端口冲突）
- B：server/worker 目录与前置准备
- C：server/worker 离线部署（包校验、解压、unit 安装与启动）
- D：双服务运行态检查（unit、端口、重启次数）
- F：双服务证据采集（配置哈希、进程参数、端口快照）

## 特点

- C 阶段统一处理 server/worker 两份离线包与 systemd unit
- D 阶段统一检查两个 unit、两个端口与重启计数
- F 阶段输出 server/worker 两类进程参数证据

## 示例变量文件

```bash
examples/vars/demo-powerjob-deploy.json
examples/vars/demo-powerjob-deploy.env
examples/vars/demo-powerjob-deploy/vars.example.yaml
examples/vars/demo-powerjob-deploy/vars.invalid.empty_server_package_file.json
```

最小变量集（必填且无模板默认值）：

- `INSTALL_ROOT`
- `CONF_DIR`
- `EVIDENCE_DIR`
- `SERVER_PACKAGE_FILE` / `SERVER_PACKAGE_SHA256` / `SERVER_EXEC` / `SERVER_DATA_DIR`
- `WORKER_PACKAGE_FILE` / `WORKER_PACKAGE_SHA256` / `WORKER_EXEC` / `SERVER_ENDPOINT`

## 校验

```bash
./opskit template validate --vars-file examples/vars/demo-powerjob-deploy.json assets/templates/demo-powerjob-deploy.json
./opskit template validate --json --vars-file examples/vars/demo-powerjob-deploy.json assets/templates/demo-powerjob-deploy.json
```

## 执行示例

```bash
./opskit run A --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
./opskit run C --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
./opskit run D --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
./opskit accept --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
```

## 注意

- 该模板为去生产化示例，不包含数据库初始化和生产级参数优化
- 如需更复杂 topology，建议在此模板基础上扩展变量和 action

## 交付门禁信息

### 接管职责（一句话）

该模板接管“PowerJob（server/worker）自部署单机节点的统一交付与运维检查”职责。

### 模板不做什么

- 不做数据库初始化与集群规划
- 不做跨节点调度一致性治理
- 不依赖外部控制面

### 单机自洽前提

- server/worker 离线包与 unit 参数可本机访问
- 输出目录可写，systemd 可用或可失败留痕
- 不依赖公网

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
./opskit run D --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
./opskit accept --template assets/templates/demo-powerjob-deploy.json --vars-file examples/vars/demo-powerjob-deploy.json --output ./.tmp/opskit-powerjob
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-powerjob-deploy/vars.example.yaml`
- 变量仅表达包、端口、unit、目录等差异

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-powerjob-deploy/vars.invalid.empty_server_package_file.json assets/templates/demo-powerjob-deploy.json
```

### 失败可交付说明

即便 server/worker 任一组件失败，仍可输出标准状态与证据入口用于交付复核。
