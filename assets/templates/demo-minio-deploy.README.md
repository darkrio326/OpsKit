# demo-minio-deploy

`demo-minio-deploy.json` 是一个去生产化的 MinIO 自行部署模板（`mode=deploy`）。
它用于演示“离线包校验 + 解压部署 + systemd 拉起 + 运行态检查 + 证据输出”完整链路。
该模板基于 `demo-generic-selfhost-deploy` 的通用 deploy 链路，增加了 MinIO 专用变量与启动参数。

- Delivery Level: `Demo`

## 适用范围

- 生命周期阶段：`A` / `B` / `C` / `D` / `F`（`E` 默认禁用）
- 目标场景：由 OpsKit 统一部署和监管的服务（如 MinIO/ELK/PowerJob）
- 非目标场景：已经是黑箱交付的厂商中间件（这类应使用 `manage` 模板）

## 部署前准备（离线）

至少准备以下文件/路径：

- MinIO 离线包（tar.gz）：`PACKAGE_FILE`
- 离线包 SHA256：`PACKAGE_SHA256`
- 可执行文件最终路径：`SERVICE_EXEC`（通常在解压目录内）
- 数据目录：`MINIO_DATA_DIR`
- 可写输出目录：`INSTALL_ROOT` / `CONF_DIR` / `EVIDENCE_DIR`

说明：

- 模板会在 `${CONF_DIR}/${SERVICE_NAME}/minio.env` 写入账号变量
- 模板会安装 `${SERVICE_UNIT}` 到 `${SYSTEMD_UNIT_DIR}`
- 模板不会下载任何在线依赖

## 变量文件

推荐直接用示例变量文件再改值：

```bash
examples/vars/demo-minio-deploy.json
examples/vars/demo-minio-deploy.env
```

最少需要你重点确认的变量：

- `PACKAGE_FILE`
- `PACKAGE_SHA256`
- `SERVICE_EXEC`
- `MINIO_ROOT_USER`
- `MINIO_ROOT_PASSWORD`

## 模板校验

文本模式：

```bash
./opskit template validate --vars-file examples/vars/demo-minio-deploy.json assets/templates/demo-minio-deploy.json
```

机器可读模式：

```bash
./opskit template validate --json --vars-file examples/vars/demo-minio-deploy.json assets/templates/demo-minio-deploy.json
```

## 执行示例

只做基础检查与运维态检查：

```bash
./opskit run A --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
./opskit run D --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
./opskit accept --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
```

执行部署链路（会做 systemd enable/start）：

```bash
./opskit run C --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
```

## 预期输出

- `state/overall.json`
- `state/lifecycle.json`
- `state/services.json`
- `state/artifacts.json`
- `evidence/<service>-env-hash.json`
- `evidence/<service>-unit-hash.json`
- `evidence/<service>-process-args.json`
- `evidence/<service>-ss-ltn.json`

## 常见失败

- `sha256_verify` 失败：`PACKAGE_SHA256` 与离线包不一致
- `untar` 失败：包路径错误或解压目录权限不足
- `systemd_enable_start` 失败：目标主机未启用 systemd 或 unit 配置有误
- `port_listening` WARN/FAIL：服务未实际监听 `SERVICE_PORT` / `CONSOLE_PORT`

## 交付门禁信息

### 接管职责（一句话）

该模板接管“MinIO 自部署单机节点的 deploy+operate+accept 闭环”职责。

### 模板不做什么

- 不负责 MinIO 集群化/分布式部署编排
- 不负责对象数据迁移与业务级初始化
- 不依赖公网拉取制品

### 单机自洽前提

- 离线包及 SHA256 在本机可访问
- 目标目录可写，systemd 可用（不可用则失败留痕）
- 不依赖外部控制平面

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
./opskit run D --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
./opskit accept --template assets/templates/demo-minio-deploy.json --vars-file examples/vars/demo-minio-deploy.json --output ./.tmp/opskit-minio
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-minio-deploy/vars.example.yaml`
- 变量只表达 package/path/unit/port/account 差异，不表达流程逻辑

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-minio-deploy.json --vars "STACK_ID=" assets/templates/demo-minio-deploy.json
```

### 失败可交付说明

即便部署失败或服务未起来，也应可执行 `accept`，输出统一状态与证据索引用于交付复核。
