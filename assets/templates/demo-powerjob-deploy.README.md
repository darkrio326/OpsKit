# demo-powerjob-deploy

`demo-powerjob-deploy.json` 是 PowerJob 自行部署示例模板（`mode=deploy`）。
单模板覆盖两个服务：

- `powerjob-server`
- `powerjob-worker`

## 特点

- C 阶段统一处理 server/worker 两份离线包与 systemd unit
- D 阶段统一检查两个 unit、两个端口与重启计数
- F 阶段输出 server/worker 两类进程参数证据

## 示例变量文件

```bash
examples/vars/demo-powerjob-deploy.json
examples/vars/demo-powerjob-deploy.env
```

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
