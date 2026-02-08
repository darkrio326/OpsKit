# demo-runtime-baseline

`demo-runtime-baseline.json` 是一个去生产化的运行时基线检查模板。
该模板聚焦通用服务器就绪性检查与证据输出，不包含任何中间件部署逻辑。

## 适用范围

- 生命周期阶段：`A` / `D` / `F`
- 检查类别：系统信息、挂载、时间同步、DNS、磁盘/内存/负载
- 证据类别：文件哈希、目录哈希、命令输出

## 变量说明

- `INSTALL_ROOT`（`group=paths`，必填）：输出根目录（state/reports）
- `EVIDENCE_DIR`（`group=paths`，必填）：证据输出目录
- `ROOT_MOUNT`（`group=runtime`，默认 `/`）：磁盘检查目标挂载点
- `DNS_HOST`（`group=network`，默认 `localhost`）：`dns_resolve` 检查使用的主机名
- `PROFILE`（`group=runtime`，默认 `baseline`）：证据文件名前缀标签

## 模板校验

```bash
go run ./cmd/opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
```

## 执行示例（可选）

```bash
go run ./cmd/opskit run A --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
go run ./cmd/opskit run D --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
go run ./cmd/opskit accept --template assets/templates/demo-runtime-baseline.json --vars-file examples/vars/demo-runtime-baseline.json --output /tmp/opskit-demo-runtime
```

预期输出：

- `/tmp/opskit-demo-runtime/state/*.json`
- `/tmp/opskit-demo-runtime/reports/*.html`
- `/tmp/opskit-demo-runtime/evidence/*-hash.json`
