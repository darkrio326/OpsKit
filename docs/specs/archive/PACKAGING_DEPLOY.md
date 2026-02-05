单文件交付与 install 行为

1. 交付目标
	•	单二进制：opskit
	•	离线可运行

⸻

2. install 行为

opskit install 将执行：
	•	创建目录：
	•	/var/lib/opskit/{state,reports,evidence,cache}
	•	/usr/share/opskit/ui
	•	安装 systemd：
	•	opskit-web.service
	•	opskit-patrol.timer
	•	opskit-recover.service
	•	初始化状态：
	•	自动执行一次 A + D

⸻

3. 卸载
	•	停止并 disable 所有 opskit service/timer
	•	删除 opskit 目录（可选保留 evidence）

⸻

4. Docker（银河麒麟 V10）验证部署
	•	基础镜像定义：`docker/kylin-v10/Dockerfile`
	•	一键验证脚本：`examples/generic-manage/run-af-kylin-v10-docker.sh`
	•	详细说明：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
