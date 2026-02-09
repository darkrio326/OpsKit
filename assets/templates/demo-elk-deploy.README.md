# demo-elk-deploy

`demo-elk-deploy.json` 是一个“单模板维护多服务”的 deploy 示例，覆盖：

- Elasticsearch
- Logstash
- Kibana

该模板演示同一模板下的多服务离线部署与统一运维检查。

- Delivery Level: `Demo`

## 特点

- A 阶段一次性检查 3 个端口冲突（`9200/5044/5601`）
- C 阶段依次执行 3 份离线包校验、解压、unit 安装和拉起
- D 阶段统一检查 3 个 unit 状态、3 个监听端口、3 个重启计数
- F 阶段统一输出多服务证据（env hash + process args + 端口快照）
- 支持 JVM 调优变量：`ES_JAVA_OPTS`、`LOG_JAVA_OPTS`
- 支持 Logstash pipeline 路径变量：`LOG_PIPELINE_CONFIG`
- 支持 Kibana TLS 变量：`KB_TLS_ENABLED`、`KB_TLS_CERT_FILE`、`KB_TLS_KEY_FILE`

## 示例变量文件

```bash
examples/vars/demo-elk-deploy.json
examples/vars/demo-elk-deploy.env
```

## 校验

```bash
./opskit template validate --vars-file examples/vars/demo-elk-deploy.json assets/templates/demo-elk-deploy.json
./opskit template validate --json --vars-file examples/vars/demo-elk-deploy.json assets/templates/demo-elk-deploy.json
```

## 执行示例

```bash
./opskit run A --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
./opskit run C --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
./opskit run D --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
./opskit accept --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
```

## 注意

- 该模板是去生产化演示，不包含集群拓扑与生产参数治理
- 三个服务使用一个模板统一编排，适合“同一服务器托管多组件”场景
- 开启 Kibana TLS 时请同时提供证书与私钥文件，并保证路径可读

## 交付门禁信息

### 接管职责（一句话）

该模板接管“ELK（ES/Logstash/Kibana）自部署单机节点的统一交付与监管”职责。

### 模板不做什么

- 不做多节点 ELK 集群编排
- 不做外部存储/消息系统依赖管理
- 不在 vars 中承载分支策略

### 单机自洽前提

- 三个服务离线包与路径参数可本机访问
- systemd 与端口检查可运行（失败会留痕）
- 不依赖外部控制面与公网

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
./opskit run D --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
./opskit accept --template assets/templates/demo-elk-deploy.json --vars-file examples/vars/demo-elk-deploy.json --output ./.tmp/opskit-elk
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-elk-deploy/vars.example.yaml`
- 变量只表达 package/unit/port/path/JVM 参数等差异

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-elk-deploy.json --vars "STACK_ID=" assets/templates/demo-elk-deploy.json
```

### 失败可交付说明

即便单个组件部署或巡检失败，仍应保留统一 `state/report/bundle` 结构，支持故障交付与复核。
