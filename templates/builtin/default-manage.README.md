# generic-manage-v1（builtin/default-manage.json）

`generic-manage-v1` 是 OpsKit 的内置标准 Manage 模板，面向“单机通用服务器接管”。

- Delivery Level: `Pilot`

## 接管职责（一句话）

该模板接管“单机服务器的基础巡检、运行态检查与验收证据输出”职责（A/D/F）。

## 模板不做什么

- 不做业务系统部署或升级
- 不做跨节点编排
- 不承载客户环境定制策略

## 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template generic-manage-v1 --output ./.tmp/opskit-generic-manage
./opskit run D --template generic-manage-v1 --output ./.tmp/opskit-generic-manage
./opskit accept --template generic-manage-v1 --output ./.tmp/opskit-generic-manage
```

## vars 与边界

- vars 示例：`examples/vars/generic-manage-v1/vars.example.yaml`
- 本模板默认无需必填 vars；如需覆盖，仅用于路径/端口等环境差异

校验失败示例（本模板无必填 vars）：

```bash
# generic-manage-v1 无“缺失必填变量”场景，本项不适用
./opskit template validate --json generic-manage-v1
```

## 单机自洽前提

- 本机可执行基础系统命令
- 输出目录可写
- 不依赖外部控制面和公网

## 失败可交付说明

即便 A/D 阶段失败，仍可执行 `accept` 产出标准 `state/report/bundle` 结构用于验收与移交。
