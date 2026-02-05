OpsKit Docker 部署说明（银河麒麟 V10）

1. 目标
	•	使用“纯净”银河麒麟 V10 容器作为验证环境
	•	在容器里执行 OpsKit 通用能力 A~F 流程
	•	输出完整 state/reports/bundles/summary 产物，便于回归测试

⸻

2. 目录与文件
	•	基础镜像 Dockerfile：`docker/kylin-v10/Dockerfile`
	•	一键验证脚本：`examples/generic-manage/run-af-kylin-v10-docker.sh`
	•	通用流程脚本：`examples/generic-manage/run-af.sh`

⸻

3. 前置条件
	•	本机可用 Docker
	•	本机可用 Go（用于构建 Linux 版 `opskit`）
	•	仓库根目录执行命令

⸻

4. 直接运行（推荐）

```bash
OUTPUT=$PWD/.tmp/generic-e2e-kylin \
  ./examples/generic-manage/run-af-kylin-v10-docker.sh
```

或使用 Make：

```bash
make -C examples/generic-manage docker-kylin-e2e
```

脚本会自动：
	•	识别镜像架构并构建匹配架构的 Linux 二进制
	•	启动 `kylinv10/kylin:b09` 容器
	•	将主机目录挂载为容器内 `/opt` `/data` `/logs`（满足 mount_check）
	•	执行 `template validate -> install -> run AF -> status -> accept -> handover`

⸻

5. 可选参数

```bash
IMAGE=kylinv10/kylin:b09 \
DOCKER_PLATFORM=linux/amd64 \
OUTPUT=$PWD/.tmp/generic-e2e-kylin-amd64 \
./examples/generic-manage/run-af-kylin-v10-docker.sh
```

	•	`IMAGE`：容器镜像（默认 `kylinv10/kylin:b09`）
	•	`DOCKER_PLATFORM`：可强制平台（例如 `linux/amd64`）
	•	`OUTPUT`：验证产物目录
	•	`DRY_RUN`：是否只执行 dry-run（`0` 或 `1`）

⸻

6. 验证产物
	•	`OUTPUT/state/overall.json`
	•	`OUTPUT/state/lifecycle.json`
	•	`OUTPUT/state/services.json`
	•	`OUTPUT/state/artifacts.json`
	•	`OUTPUT/reports/*.html`
	•	`OUTPUT/bundles/*.tar.gz`
	•	`OUTPUT/summary.json`
	•	`OUTPUT/ui/index.html`

说明：在容器环境下，部分 systemd 相关项通常是 WARN，这属于预期行为。
