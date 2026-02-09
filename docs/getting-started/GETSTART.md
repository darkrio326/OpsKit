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

机器可读校验（推荐用于脚本/CI）：

```bash
./opskit template validate generic-manage-v1
./opskit template list --json
./opskit template validate --vars-file examples/vars/demo-server-audit.json --json assets/templates/demo-server-audit.json
./opskit template validate --vars-file examples/vars/demo-runtime-baseline.json --json assets/templates/demo-runtime-baseline.json
./opskit template validate --vars-file examples/vars/demo-blackbox-middleware-manage.json --json assets/templates/demo-blackbox-middleware-manage.json
```

失败样例（用于断言）：

```bash
./opskit template validate --json /no/such/template.json
```

建议校验字段：

- `valid=false`
- `errorCount>0`
- `issues[0].code=template_file_not_found`

## 3) 本地 dry-run 跑通 A~F

```bash
./opskit install --template generic-manage-v1 --dry-run --no-systemd --output ./.tmp/opskit-demo
./opskit run AF --template generic-manage-v1 --dry-run --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo --json
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

## 6.1) Web UI 自动刷新服务器状态（推荐）

`web` 进程可后台定时刷新 `state/*.json`，UI 再读取这些最新状态：

```bash
./opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080 --status-interval 15s
```

说明：

- `--status-interval` 默认 `15s`
- 设为 `0` 或负值可关闭后台状态刷新
- UI 前端自动刷新只负责重读 JSON，不执行巡检动作

## 7) 麒麟离线机一键回归（v0.3.x）

```bash
scripts/kylin-offline-validate.sh \
  --bin /usr/local/bin/opskit \
  --output /data/opskit-regression \
  --json-status-file /data/opskit-regression/status.json \
  --summary-json-file /data/opskit-regression/summary.json \
  --clean
```

模板 JSON 契约门禁（建议一起执行）：

```bash
scripts/template-validate-check.sh --clean
```

## 8) 进入真实服务器前统一门禁（建议）

```bash
scripts/generic-readiness-check.sh --clean
```

可选：附加 `release-check summary.json` 契约门禁

```bash
scripts/generic-readiness-check.sh --with-release-json-contract --clean
```

严格模式（通用链路与离线回归都要求 exit=0）：

```bash
scripts/generic-readiness-check.sh --generic-strict --offline-strict --clean
```

更多说明：

- 项目总览：`README.md` / `README.zh-CN.md`
- UI 模板视图说明：`docs/getting-started/UI_TEMPLATE_USAGE.md`
- UI 截图版本同步：`scripts/screenshot-sync.sh --version <version>` + `scripts/screenshot-check.sh --version <version>`
- 麒麟离线部署（从 Release 包开始）：`docs/getting-started/KYLIN_V10_OFFLINE_RELEASE.md`
- 麒麟离线回归清单（v0.3.x）：`docs/getting-started/KYLIN_V10_OFFLINE_VALIDATION.md`
- 麒麟容器部署：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- 通用示例：`docs/examples/generic-manage/README.md` / `docs/examples/generic-manage/README.zh-CN.md`
