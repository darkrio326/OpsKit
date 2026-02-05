OpsKit 路线图（从通用服务器到模板化交付）

0. 产品总目标
	•	一个二进制（opskit）
	•	一个命令部署：opskit install
	•	一个静态页面 + 若干 JSON（状态/阶段/报告入口）
	•	覆盖 A~F 全生命周期：Preflight → Baseline → Deploy → Operate → Recover → Accept/Handover
	•	先通用、后模板：先把“接管-巡检-恢复-验收证据”跑通，再做 ELK/OA 等模板化扩展

⸻

阶段路线图（按里程碑推进）

Milestone 1：通用服务器“可看、可跑、可留证据”（底座完成）

目标：拿到一台全新服务器，你能一键接管并在页面看到完整 A~F 架子（哪怕很多阶段是空壳），并且有真实数据可展示。

交付物
	•	opskit 单二进制
	•	opskit install 一键安装：
	•	静态页面（UI）
	•	state/reports 目录
	•	systemd 服务/定时器（至少 operate）
	•	立即生成首屏 JSON
	•	页面出现 A~F 六张卡片（骨架），全局状态条可用

核心能力（通用版）
	•	A Preflight（只做采集+判定，不做修复）
	•	OS/架构/内核/时区
	•	网络/路由/DNS
	•	磁盘/挂载（/data 等）
	•	防火墙/SELinux 状态
	•	端口占用（通用端口清单 + 可配置）
	•	D Operate（只做巡检）
	•	systemd unit 状态（通用列表可为空）
	•	资源阈值（磁盘/内存/负载）

完成标志：不管服务器用途是什么，你能“接管+出报告+页面可视化”。

⸻

Milestone 2：通用服务器“可自愈”（停电恢复闭环）

目标：停电/重启后，服务器起来就能自动检查并尝试恢复关键服务；失败能留下证据包，方便人工介入。

交付物
	•	systemd 开机触发恢复：OnBootSec=2min
	•	可选周期巡检自愈：每 5~10 分钟（带熔断/有限重试）
	•	页面 Recover 卡片可展示：
	•	是否启用
	•	上次恢复时间/结果
	•	失败原因摘要
	•	报告入口/证据包入口

核心能力（通用版）
	•	E Recover
	•	就绪判定：网络/挂载/时间等
	•	恢复策略：有限重试（1~2 次）+ 熔断避免重启风暴
	•	失败兜底：自动 collect（日志/状态/配置摘要）

完成标志：你说的“机房来电后立刻检测，没起来就拉起”，在通用环境下能跑通。

⸻

Milestone 3：通用服务器“可验收”（证据包能力闭环）

目标：不管部署的是啥系统，只要甲方要求“补丁包/版本证据”，OpsKit 都能给出一套不可争辩的证据包。

交付物
	•	opskit accept：生成 acceptance-YYYYMMDD.tar.gz
	•	页面 Accept/Handover 卡片可展示：
	•	最近一次证据包时间
	•	前端/后端/配置的 hash 摘要
	•	报告入口 + 下载入口
	•	opskit handover：汇总 A~F 报告，形成交付包目录/报告

核心能力（通用版）
	•	F Accept
	•	前端/后端/配置：hash、路径、生效状态（尽力做到一致性校验）
	•	证据清单（JSON）+ 人类可读报告（HTML）
	•	F Handover
	•	汇总 A~F 阶段关键信息，形成“交付总报告”

完成标志：你能把“验收证据”从人工截图写 Word，变成一键生成、可归档。

⸻

进入模板化（在通用能力跑通之后再做）

到这里你已经有一个“通用 OpsKit”。接下来才进入你说的：

“对不同服务器增加不同配置模板”

我建议把模板做成 Stack Profiles（用途配置模板），每个模板只定义三件事：
	1.	关键服务单元（systemd units / compose service）
	2.	健康检查规则（端口/HTTP/日志关键字/集群状态）
	3.	恢复顺序与依赖（start order / wait condition）

⸻

Milestone 4：ELK 模板（你最常用、最可控）

目标：ELK 变成“声明式”部署与运维，页面上能显示 ELK 专属指标。
	•	Deploy：离线包 → systemd 注册 → 启动 → 校验
	•	Operate：ES 集群健康、磁盘水位、关键索引/日志异常 Top
	•	Recover：按依赖顺序拉起 + 集群健康校验
	•	Accept：输出 ES/Kibana/Logstash 版本与配置证据

⸻

Milestone 5：OA（基于金蝶）模板（先“接管运维”，后“辅助部署”）

目标：不跟金蝶内部实现纠缠，先把“可控部分”做实。
	•	先做“接管式模板”：
	•	识别 OA 相关 units、端口、日志路径
	•	健康检查 = unit + 端口 + 可选 HTTP
	•	Recover = 顺序拉起 + 有限重试
	•	Accept = jar/war/配置证据 + 运行态摘要
	•	以后再扩展“部署辅助”（如果你确实掌握可通用的部署方法）

⸻

Milestone 6：MinIO 模板（中等频率，适合标准化）
	•	Deploy：离线二进制/容器方式都可
	•	Operate：健康、存储容量、磁盘告警
	•	Recover：启动 + API 健康检查
	•	Accept：版本、配置、证书摘要（脱敏）

⸻

最终收敛：单文件 + 单命令部署（贯穿全程）

你从第一天就按这个目标做，最后不会返工。

最终体验
	•	scp opskit <server>
	•	./opskit install
	•	浏览器打开页面 → A~F 卡片 + 状态 + 报告入口齐全
	•	后续：
	•	./opskit run preflight|recover|accept
	•	或定时器自动刷新

⸻

建议的推进顺序
	1.	Milestone 1：通用 Preflight + Operate + 页面骨架 + install
	2.	Milestone 2：Recover 自愈闭环
	3.	Milestone 3：Accept/Handover 证据闭环
	4.	Milestone 4：ELK 模板（第一个真正模板）
	5.	Milestone 5：OA 接管式模板
	6.	Milestone 6：MinIO 模板

⸻

通用能力一键自检（进入模板化前）

为确保“先通用后模板”不偏移，仓库提供：

	•	examples/generic-manage/run-af.sh

默认流程：

	•	template validate
	•	install（本地示例默认 --no-systemd）
	•	run AF
	•	status
	•	accept
	•	handover

建议在每次进入模板开发前先跑一次，确认通用 A~F 产物链稳定。
