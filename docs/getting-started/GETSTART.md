# OpsKit GetStart

这是一份根目录快速上手说明，帮助你在几分钟内完成本地验证。

## 1) 构建二进制

```bash
go build -o opskit ./cmd/opskit
```

## 2) 验证模板

```bash
./opskit template validate templates/builtin/default-manage.json
```

## 3) 本地 dry-run 跑通 A~F

```bash
./opskit install --template generic-manage-v1 --dry-run --no-systemd --output ./.tmp/opskit-demo
./opskit run AF --template generic-manage-v1 --dry-run --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
```

## 4) 用示例脚本跑通通用链路

```bash
BIN=./opskit OUTPUT=$PWD/.tmp/generic-e2e ./examples/generic-manage/run-af.sh
```

## 5) 在银河麒麟 V10 Docker 中做纯净验证

```bash
make -C examples/generic-manage docker-kylin-e2e
```

可选参数：

```bash
IMAGE=kylinv10/kylin:b09 DOCKER_PLATFORM=linux/amd64 DRY_RUN=1 make -C examples/generic-manage docker-kylin-e2e
```

## 6) 关键产物位置

- `OUTPUT/state/overall.json`
- `OUTPUT/state/lifecycle.json`
- `OUTPUT/state/services.json`
- `OUTPUT/state/artifacts.json`
- `OUTPUT/reports/*.html`
- `OUTPUT/bundles/*.tar.gz`
- `OUTPUT/ui/index.html`
- `OUTPUT/summary.json`

更多说明：

- 项目总览：`README.md` / `README.zh-CN.md`
- 麒麟容器部署：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- 通用示例：`docs/examples/generic-manage/README.md` / `docs/examples/generic-manage/README.zh-CN.md`
