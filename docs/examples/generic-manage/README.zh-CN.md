# generic-manage 端到端示例

该示例用于在进入业务模板前，先验证 **通用能力（非模板特化）** 是否跑通。

执行流程：

1. template validate
2. install（本地示例默认不安装 systemd）
3. run AF
4. status
5. accept
6. handover

## 快速开始

```bash
BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

或直接从源码运行：

```bash
GOCACHE=/tmp/opskit-gocache BIN="go run ./cmd/opskit" OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

## Dry-run 模式

```bash
DRY_RUN=1 BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

严格退出码模式（所有阶段命令要求 `exit=0`）：

```bash
STRICT_EXIT=1 BIN=./opskit OUTPUT=/tmp/opskit-generic ./examples/generic-manage/run-af.sh
```

## 预期产物

- `OUTPUT/state/overall.json`
- `OUTPUT/state/lifecycle.json`
- `OUTPUT/state/artifacts.json`
- `OUTPUT/status.json`
- `OUTPUT/summary.json`
- `OUTPUT/reports/*.html`
- `OUTPUT/bundles/*.tar.gz`
- `OUTPUT/ui/index.html`

说明：

- 在非 Linux 或最小化环境中，部分检查可能是 WARN/FAILED。
- 这是通用能力验证的预期情况，脚本仍会输出产物路径供检查。
- `summary.json` 新增 `result/reasonCode/recommendedAction`，可直接用于门禁判读。

## 在银河麒麟 V10 Docker 中运行（纯净环境）

部署说明：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`

```bash
OUTPUT=$PWD/.tmp/generic-e2e-kylin \
  ./examples/generic-manage/run-af-kylin-v10-docker.sh
```

或使用示例内 Makefile：

```bash
make -C examples/generic-manage docker-kylin-e2e
```

这个封装脚本会：

- 构建 Linux 版本 `opskit`（按镜像架构自动选择 `amd64/arm64`）
- 启动 `kylinv10/kylin:b09`
- 把主机输出目录挂载到容器 `/out`
- 把主机目录挂载到容器 `/opt`、`/data`、`/logs`（满足 mount_check）
- 在容器内执行标准 `run-af.sh` 流程

可覆盖镜像/平台：

```bash
IMAGE=kylinv10/kylin:b09 \
DOCKER_PLATFORM=linux/amd64 \
OUTPUT=$PWD/.tmp/generic-e2e-kylin-amd64 \
./examples/generic-manage/run-af-kylin-v10-docker.sh
```

容器内 dry-run：

```bash
DRY_RUN=1 make -C examples/generic-manage docker-kylin-e2e
```
