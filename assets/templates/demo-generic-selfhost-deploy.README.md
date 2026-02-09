# demo-generic-selfhost-deploy

`demo-generic-selfhost-deploy.json` 是“自行部署类服务”的通用基线模板（`mode=deploy`）。
它用于抽象通用链路：离线包校验、解压、systemd 安装启动、运行态检查与证据输出。

## 用途

- 作为新模板起点：MinIO/ELK/PowerJob 这类模板都可从此基线扩展
- 作为单服务通用模板：仅替换变量即可跑通 A/B/C/D/F

## 关键变量

- `PACKAGE_FILE` / `PACKAGE_SHA256`：离线包和校验值
- `SERVICE_EXEC` / `SERVICE_ARGS`：启动命令
- `SERVICE_UNIT` / `SERVICE_PORT`：systemd 与端口
- `INSTALL_ROOT` / `CONF_DIR` / `EVIDENCE_DIR`：状态与证据路径

## 示例变量文件

```bash
examples/vars/demo-generic-selfhost-deploy.json
examples/vars/demo-generic-selfhost-deploy.env
```

## 校验

```bash
./opskit template validate --vars-file examples/vars/demo-generic-selfhost-deploy.json assets/templates/demo-generic-selfhost-deploy.json
./opskit template validate --json --vars-file examples/vars/demo-generic-selfhost-deploy.json assets/templates/demo-generic-selfhost-deploy.json
```

## 执行示例

```bash
./opskit run A --template assets/templates/demo-generic-selfhost-deploy.json --vars-file examples/vars/demo-generic-selfhost-deploy.json --output ./.tmp/opskit-generic-selfhost
./opskit run C --template assets/templates/demo-generic-selfhost-deploy.json --vars-file examples/vars/demo-generic-selfhost-deploy.json --output ./.tmp/opskit-generic-selfhost
./opskit run D --template assets/templates/demo-generic-selfhost-deploy.json --vars-file examples/vars/demo-generic-selfhost-deploy.json --output ./.tmp/opskit-generic-selfhost
./opskit accept --template assets/templates/demo-generic-selfhost-deploy.json --vars-file examples/vars/demo-generic-selfhost-deploy.json --output ./.tmp/opskit-generic-selfhost
```

## 说明

- `E` 阶段默认 `enabled=false`，避免非预期自动恢复动作
- 这是通用基线，不包含产品级配置细节
