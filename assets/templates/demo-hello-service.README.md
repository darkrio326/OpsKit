# demo-hello-service.json 使用说明

用途：演示“轻量模板 + 变量渲染 + 证据输出”流程。  
该模板为去生产化 demo，不包含真实中间件部署逻辑。

## 推荐执行顺序

```bash
./opskit run B --template assets/templates/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit run C --template assets/templates/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit run D --template assets/templates/demo-hello-service.json --output ./.tmp/opskit-hello
./opskit accept --template assets/templates/demo-hello-service.json --output ./.tmp/opskit-hello
```

## 常用变量

模板内已声明 `vars` 校验，缺少必填变量会直接报错。

- `SERVICE_NAME`（默认 `hello-service`）
- `SERVICE_PORT`（默认 `18080`）
- `SERVICE_UNIT`（默认 `hello-service.service`，仅作占位）
- `PROCESS_MATCH`（默认 `hello-service`）

示例：

```bash
./opskit run C \
  --template assets/templates/demo-hello-service.json \
  --output ./.tmp/opskit-hello \
  --vars "SERVICE_NAME=demo-hello,SERVICE_PORT=19090,PROCESS_MATCH=demo-hello"
```

使用 `--vars-file`（推荐）：

```bash
./opskit run C \
  --template assets/templates/demo-hello-service.json \
  --vars-file examples/vars/demo-hello-service.json \
  --output ./.tmp/opskit-hello
```

## 预期输出

- `conf/<service>/hello-demo.env`（模板渲染结果）
- `state/stack.json`
- `state/artifacts.json`
- `evidence/<service>-env-hash.json`
- `evidence/<service>-conf-dir-hash.json`

说明：模板中的 systemd/端口健康检查为占位并默认禁用，避免误操作真实服务。
