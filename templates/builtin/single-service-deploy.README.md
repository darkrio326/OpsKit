# single-service-deploy-v1（builtin/single-service-deploy.json）

`single-service-deploy-v1` 是 OpsKit 的内置标准 Deploy 模板，面向“单机自部署服务节点”。

- Delivery Level: `Pilot`

## 接管职责（一句话）

该模板接管“单机单服务节点的 deploy+operate+accept 基础闭环”职责（A/C/D/F）。

## 模板不做什么

- 不做多服务拓扑编排
- 不做业务级发布流程编排
- 不承载跨机房/跨节点策略

## 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template single-service-deploy-v1 --output ./.tmp/opskit-single-service-deploy
./opskit run D --template single-service-deploy-v1 --output ./.tmp/opskit-single-service-deploy
./opskit accept --template single-service-deploy-v1 --output ./.tmp/opskit-single-service-deploy
```

## vars 与边界

- vars 示例：`examples/vars/single-service-deploy-v1/vars.example.yaml`
- 本模板默认无需必填 vars；如需覆盖，仅用于 unit/port/path/package 差异

校验失败示例（本模板无必填 vars）：

```bash
# single-service-deploy-v1 无“缺失必填变量”场景，本项不适用
./opskit template validate --json single-service-deploy-v1
```

## 单机自洽前提

- 本机具备部署所需基础命令
- 输出目录可写
- 不依赖外部控制面和公网

## 失败可交付说明

即便部署动作失败，也应通过 `accept` 输出统一状态与证据索引，满足失败可交付原则。
