# Demo 模板目录

该目录仅包含可公开的 **demo 模板**，用于演示 OpsKit 通用能力与模板机制。

## 服务器两种接入模式

- `manage`（黑箱存量）：适用于已部署、不可改安装过程的中间件（如金蝶/FCS/WO）
- `deploy`（自行部署）：适用于由 OpsKit 驱动部署与运维的服务（如 ELK/MinIO/PowerJob）

约束：

- 黑箱场景优先使用 `manage`，不承诺复现厂商安装流程
- 自行部署场景使用 `deploy`，可逐步补齐安装/配置/启动动作

当前 demo 覆盖：

- 黑箱 `manage`：`demo-blackbox-middleware-manage.json`
- 通用审计 `manage`：`demo-server-audit.json`、`demo-runtime-baseline.json`
- 轻量 `deploy` 演示：`demo-hello-service.json`
- 通用 `deploy` 基线：`demo-generic-selfhost-deploy.json`
- 自行部署 `deploy` 示例：`demo-minio-deploy.json`、`demo-elk-deploy.json`、`demo-powerjob-deploy.json`

- `demo-server-audit.json`：仅 A/D/F 审计与证据链
- `demo-hello-service.json`：轻量 hello-service 模板演示
- `demo-runtime-baseline.json`：运行时基线检查模板（A/D/F）
- `demo-blackbox-middleware-manage.json`：黑箱中间件服务器 Manage 模板（vars 可选 default/FCS/Kingdee 预设）
- `demo-generic-selfhost-deploy.json`：自行部署通用基线（单服务）
- `demo-minio-deploy.json`：MinIO 自行部署模板（离线包校验 + systemd 部署 + 运维检查）
- `demo-elk-deploy.json`：ELK 一体模板（Elasticsearch + Logstash + Kibana）
- `demo-powerjob-deploy.json`：PowerJob 一体模板（server + worker）

配套说明：

- `demo-server-audit.README.md`
- `demo-hello-service.README.md`
- `demo-runtime-baseline.README.md`
- `demo-blackbox-middleware-manage.README.md`
- `demo-generic-selfhost-deploy.README.md`
- `demo-minio-deploy.README.md`
- `demo-elk-deploy.README.md`
- `demo-powerjob-deploy.README.md`

注意：此目录不包含任何生产级部署模板或客户环境信息。
