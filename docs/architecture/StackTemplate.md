1) Stack Template 的目标与边界

目标

一个 Stack Template 只描述三类东西：
	1.	关键对象：这个栈由哪些服务/组件组成（systemd unit / compose service）
	2.	判定规则：怎么判断“健康/不健康/降级”（端口、HTTP、日志、命令输出）
	3.	动作策略：部署/启动顺序、恢复顺序、重试/熔断策略、证据采集范围

边界（必须明确）
	•	Template 不能改变 OpsKit 通用阶段 A/B/D/E/F 的基本规则，只能追加检查项/健康项/证据项。
	•	Template 不等同于“部署脚本”。对像 OA（金蝶）这种你不可控的系统，模板应该先走 接管式（Manage-only），而不是强部署。
	•	Template 必须支持“离线/内网”约束：不能依赖外网 API、在线仓库。

⸻

2) Template 的运行模型：Stack → Components → Checks → Actions

核心概念
	•	Stack：一个用途（elk / oa / minio）
	•	Component：栈内组件（elasticsearch / kibana / oa-backend / nginx）
	•	Check：健康判定（port/http/command/log）
	•	Action：可执行动作（start/stop/restart/deploy/recover/collect）
	•	Evidence：验收证据（版本/包hash/配置hash/运行态摘要）
	•	Dependencies：依赖与顺序（先 A 再 B，等待条件）

⸻

3) Template 结构规范（字段语义设计）

3.1 模板元信息（必须）
	•	id：唯一标识（elk、oa-kingdee、minio）
	•	name：展示名
	•	mode：运行模式
	•	manage：接管式（不负责安装部署，只负责巡检/恢复/证据）
	•	deploy：可部署式（包含离线部署流程）
	•	targets：适用范围
	•	OS：kylin v10
	•	arch：aarch64/x86_64
	•	criticality：对 Overall Status 的影响
	•	critical：这个栈挂了算红
	•	noncritical：挂了算黄

3.2 组件定义（必须）

每个 component 包含：
	•	id：组件名
	•	type：systemd / compose / process（未来可扩）
	•	unit/serviceName：systemd unit 或 compose service
	•	ports：关键端口列表
	•	health：健康检查规则集合
	•	startPolicy：启动策略（启动/重启/等待）
	•	deps：依赖组件列表
	•	evidence：该组件提供的证据采集项

3.3 健康检查规则（Check 规范）

每个 check 都要有：
	•	kind：unit / port / http / command / log_keyword
	•	id：check 名称
	•	severity：fail（影响红）/ warn（影响黄）/ info
	•	timeout：超时
	•	retry：重试次数（通常 0~1）
	•	successCriteria：成功判定条件（比如 http 200、命令返回包含某字符串）

关键点：判定逻辑必须是“客观可验证”，不要写成经验口吻。

3.4 恢复策略（Recover 规范）

Stack 级别定义：
	•	recover.enabled：是否参与自动恢复
	•	recover.order：组件启动顺序（默认按 deps 拓扑排序）
	•	recover.maxAttempts：每次恢复最多重试次数
	•	recover.cooldown：失败冷却时间（熔断）
	•	recover.readiness：环境就绪条件（网络/挂载/时间等）
	•	recover.failureAction：失败后动作（collect/mark_failed/notify）

3.5 证据采集（Accept/Evidence 规范）

每个 stack/component 的 evidence：
	•	kind：file_hash / dir_hash / command_output / config_snapshot / process_args
	•	path：文件/目录路径（可变量）
	•	mask：脱敏规则（比如配置中 password/token 替换为 ***）
	•	label：展示名

验收最关键：证据必须可复核（hash + 路径 + 时间）。

⸻

4) 执行流程（OpsKit 如何使用 Template）

Template 不直接“控制全生命周期”，它只在这些环节被调用：

A Preflight（通用）+ Template 扩展
	•	通用：系统基础
	•	模板追加：检查模板所需端口是否冲突、关键目录是否存在/可写

C Deploy（仅 deploy 模式）
	•	模板定义离线包输入、解包、配置渲染、unit 注册、启动校验
	•	输出 deploy 报告与证据

D Operate（核心）
	•	按模板的 checks 计算：
	•	component 健康
	•	stack 健康
	•	overall 绿/黄/红

E Recover（核心）
	•	readiness 满足 → 按 order/dep 启动 → health 校验 → 重试/熔断

F Accept/Handover（核心）
	•	按模板 evidence 采集 → 生成验收包与交付汇总

⸻

5) Template 的验收标准（很重要，防止模板越写越歪）

一个新模板要进入“可用模板库”，必须满足：
	1.	不依赖外网
	2.	检查项可复核（端口/HTTP/命令输出明确）
	3.	恢复策略可控（有限重试 + 熔断）
	4.	证据可审计（hash/路径/时间齐全）
	5.	模式匹配
	•	deploy 模板必须能独立完成部署并产出证据
	•	manage 模板必须能接管巡检/恢复/证据，且不破坏现有系统

⸻

6) 三个模板示例（设计级）

下面给你三个“按这个规范就能写出来”的模板逻辑。

⸻

6.1 ELK 模板（deploy 模式，强可控）

Stack：elk（critical）

Components
	1.	elasticsearch（systemd）

	•	health：
	•	unit active（fail）
	•	port 9200（fail）
	•	command：GET /_cluster/health 或本地 curl（warn/fail按状态）
	•	evidence：
	•	es 二进制/包 hash
	•	config hash（elasticsearch.yml 脱敏）
	•	JVM args 摘要
	•	cluster health snapshot（command_output）

	2.	kibana（systemd）

	•	health：
	•	unit active（warn或fail按你的业务重要性）
	•	port 5601（warn）
	•	http /api/status（warn/fail）
	•	evidence：kibana.yml hash、版本输出

	3.	logstash（可选）

	•	health：unit active、port（如有）
	•	evidence：pipeline 配置 hash

Recover 策略
	•	readiness：网络 OK、/data OK、时间 OK
	•	order：elasticsearch → kibana → logstash
	•	maxAttempts：1
	•	cooldown：10min
	•	failureAction：collect（journal + config摘要 + df/free）

⸻

6.2 OA（金蝶）模板（manage 模式，先接管）

Stack：oa-kingdee（critical）

Components（你先按你实际 unit 录入）
	1.	oa-app（systemd）

	•	health：
	•	unit active（fail）
	•	port 8080（fail）或实际端口
	•	log_keyword（warn）：最近 N 分钟有没有“Started/Server started”或成功关键字（可选）
	•	evidence：
	•	部署包（war/jar）hash（如果路径固定）
	•	启动参数摘要（process_args）
	•	配置目录 hash（脱敏）

	2.	nginx（systemd）

	•	health：unit + port 80/443（warn/fail 看你是否必须）
	•	evidence：nginx.conf hash（脱敏）

Recover 策略
	•	readiness：网络、时间、挂载
	•	order：数据库/中间件（如果你也纳入模板）→ oa-app → nginx
	•	maxAttempts：1
	•	cooldown：10min
	•	failureAction：collect（journal + 关键目录列表）

注意：manage 模式不“部署 OA”，只做接管、证明、恢复。

⸻

6.3 MinIO 模板（deploy 或 manage 都行，建议 deploy）

Stack：minio（noncritical or critical 看你业务）

Components
	1.	minio（systemd 或 compose）

	•	health：
	•	unit active（fail）
	•	port 9000/console端口（fail/warn）
	•	http /minio/health/live（fail）
	•	evidence：
	•	minio 版本输出
	•	配置/env hash（脱敏）
	•	数据目录路径与挂载证明

Recover
	•	readiness：数据盘挂载必须就绪
	•	order：minio
	•	failureAction：collect

⸻

7) 模板与“通用服务器”的关系（你提的策略落地）

你说“先通用全新服务器，再加模板”，对应到设计就是：
	•	通用服务器：A/B/D/E/F 的基础能力 + 页面骨架
	•	模板库：只负责把 C/D/E/F “细化到具体栈”

而且强制规则：
	•	没模板也能跑（通用 operate 只看系统资源 + 少量通用 services）
	•	加模板后才有业务栈指标（ELK/OA/MinIO）

⸻

8) 最后给你一个“模板最小规范”（便于你实现时不膨胀）

如果你想极简落地，我建议 v1 模板只支持：
	•	Component.type：systemd
	•	Checks.kind：unit + port + http + command（不做 log）
	•	Evidence.kind：file_hash + command_output
	•	Recover：order + maxAttempts + cooldown + readiness(network+mount)


==============================================
1) 单机单产品模板：定位与模式

模式
	•	mode = manage（接管式）
	•	适用场景：服务器上已安装/已部署产品，我们要把它变成“可运维、可自愈、可验收”的交付态

模板必须描述的“5 件事”
	1.	关键服务单元（systemd units）
	2.	关键端口（对外/对内）
	3.	健康判定规则（最少 unit+port，最好加 http）
	4.	恢复顺序与重试策略
	5.	证据采集范围（可验收：包/配置/运行态）

⸻

2) 模板最小字段（v1 必填 / 可选）

这是你实现时“不会纠结”的最小子集。

必填
	•	product.id：如 kingdee / yongzhong-fcs
	•	product.name
	•	criticality：critical（通常生产业务都是 critical）
	•	components[]：至少 1 个
	•	component.id
	•	component.type = systemd
	•	component.unit
	•	component.ports[]（可为空，但建议填）
	•	component.healthChecks[]（至少 unit active）
	•	recover：
	•	enabled
	•	order[]（组件顺序）
	•	maxAttempts（建议 1）
	•	cooldown（建议 10min）
	•	readiness（至少 network+mount）
	•	evidence[]：至少 2 类（包hash、配置hash或启动参数）

可选（v2 再加）
	•	日志关键字健康（log_keyword）
	•	进程级健康（process）
	•	自动 collect 包
	•	变更/补丁追踪（watch）

⸻

3) 对应 A~F 阶段：单产品模板应做什么

通用阶段不变，模板只“追加/细化”业务相关部分。

⸻

A Preflight（模板追加项）

目的：确认这台机器“跑这个产品”具备基本条件，且不会跟现网冲突。

检查项（单产品模板通用清单）
	•	systemd 可用
	•	关键 unit 是否存在（有则显示为“可接管”，无则提示“未安装/未注册”）
	•	关键端口是否被占用（冲突数）
	•	关键目录是否存在（部署目录/配置目录/日志目录/数据目录）
	•	挂载点是否就绪（如产品依赖 /data）

Preflight 验收标准（模板级）
	•	能识别出：这个产品“是否已部署”“是否可接管”
	•	能明确给出：端口冲突/目录缺失/挂载不就绪等原因

⸻

B Baseline（模板追加项）

目的：把运维空间打好，让后续巡检/恢复/证据能落地。

模板追加动作
	•	为产品建立标准化的“证据采集目录”（存放 hash/快照）
	•	建立对产品日志目录的只读权限可见（不改产品本身权限，最多给建议）
	•	固化产品关键路径（以后 Accept 用）

Baseline 验收标准（模板级）
	•	OpsKit 能在不破坏现网的前提下，稳定读取到：
	•	unit 状态、端口、关键路径

⸻

C Deploy（manage 模式下的定义：接管声明）

注意：manage 模式不“部署”，但 C 阶段仍然存在，它负责“用途声明+接管边界”。

模板在 C 阶段做什么
	•	将该服务器标记为：product=<kingdee|fcs>
	•	建立“产品组件清单”（units/ports/paths）
	•	生成 deploy(manage) 报告：说明我们接管的范围与能力

Deploy（manage）验收标准
	•	页面上能看到该服务器“用途明确”
	•	报告里明确：
	•	我们管理哪些 unit
	•	我们怎么判活
	•	我们如何恢复
	•	我们采集哪些证据

⸻

D Operate（巡检）

这是单产品模板的核心价值之一。

检查项（单产品模板 v1）

对每个 component（systemd unit）：
	•	unit active（FAIL）
	•	端口监听（FAIL 或 WARN，按关键程度）
	•	可选：HTTP 探活（FAIL 或 WARN）
对系统资源（复用通用）：
	•	磁盘使用率阈值（WARN）
	•	内存余量（WARN）

产出
	•	组件级状态：healthy/degraded/unhealthy
	•	产品级状态：healthy/degraded/unhealthy
	•	Overall status：是否拉红（critical = 红）

Operate 验收标准
	•	断开产品服务后，页面 1 个巡检周期内能变红
	•	恢复后，页面能回绿
	•	状态变化有记录（至少最近 N 次）

⸻

E Recover（断电/异常恢复）

这是你之前场景的核心卖点。

恢复策略（单产品模板 v1）
	•	readiness：
	•	network ready
	•	挂载就绪（如果依赖 /data）
	•	order：按模板指定组件顺序
	•	maxAttempts：1（只重试一次）
	•	cooldown：10min（失败后熔断）
	•	failureAction：
	•	标记失败 + 生成 collect 包（建议默认开启 collect）

Recover 验收标准
	•	重启服务器后能自动触发恢复一次
	•	产品未起时能尝试拉起
	•	拉起失败不会无限循环
	•	失败时能生成证据包（给人工排障）

⸻

F Accept / Handover（验收证据与交付）

这就是你说的“补丁包/验收证据”。

单产品模板的证据清单（v1）

至少包含：
	•	服务清单证据：unit 列表 + 当前状态
	•	版本证据（三选一或多选）：
	•	command_output：产品版本命令（如果可用）
	•	file_hash：核心 war/jar/可执行文件 hash
	•	dir_hash：部署目录摘要 hash（不建议全量，选关键目录）
	•	配置证据：
	•	配置文件 hash（脱敏）
	•	运行态证据：
	•	进程启动参数摘要（可脱敏）
	•	监听端口列表快照

交付汇总
	•	A~F 报告链接汇总
	•	生成 acceptance-YYYYMMDD.tar.gz

Accept/Handover 验收标准
	•	一键生成证据包
	•	包内信息满足“可复核、可归档、可解释”
	•	能回答验收常见问题：
	•	当前版本是什么
	•	部署路径在哪里
	•	运行是否正常
	•	最近一次恢复/异常是什么时候

⸻

4) 金蝶 vs 永中FCS：如何落到同一模板

它们只在“组件清单/端口/路径/探活方式”上不同，模板框架完全一致。

你需要为每个产品填的“产品参数表”

（这就是你以后快速做更多模板的方式）
	1.	产品名称/ID
	2.	systemd units（通常 1~3 个）
	3.	关键端口（外部访问、内部依赖）
	4.	部署目录（可选，但最好有）
	5.	配置目录（可选）
	6.	日志目录（可选）
	7.	探活方式（端口/HTTP/命令）
	8.	版本/包证据获取方式（hash路径或命令）

做第一个模板时，你就把这张表固化下来，以后加模板就是填表。

⸻

5) 单产品模板的“整体验收标准”（你进入实现的门槛）

当你说“模板做完了”，至少要满足：
	•	✅ 页面上这个产品的阶段卡片 A~F 全部能从 NOT_STARTED 变成有状态
	•	✅ Operate 能准确拉红/回绿
	•	✅ Recover 能在重启后自动触发且有熔断
	•	✅ Accept 能生成证据包（可审计）
	•	✅ 不破坏现网：所有默认动作不做危险写操作（除非显式开关）

-----
{
  "templateVersion": "1.0",
  "stack": {
    "id": "product-single-manage",
    "name": "Single Product (Manage-only)",
    "mode": "manage",
    "criticality": "critical",
    "description": "Single host runs a single product. OpsKit manages health checking, recovery, and acceptance evidence; does not perform installation by default.",
    "targets": {
      "os": ["Kylin V10"],
      "arch": ["aarch64", "x86_64"],
      "offline": true
    },
    "variables": {
      "PRODUCT_NAME": "TODO_PRODUCT_NAME",
      "PRODUCT_ID": "TODO_PRODUCT_ID",
      "HOST_ROLE": "single",
      "DATA_MOUNT": "/data",
      "DEPLOY_DIR": "TODO_DEPLOY_DIR",
      "CONFIG_DIR": "TODO_CONFIG_DIR",
      "LOG_DIR": "TODO_LOG_DIR"
    }
  },

  "lifecycle": {
    "stages": [
      {
        "id": "A",
        "name": "Preflight",
        "enabled": true,
        "purpose": "Assess whether the host is ready to run the product and can be safely managed by OpsKit.",
        "checks": [
          {
            "id": "A.sysinfo",
            "kind": "system_info",
            "severity": "info",
            "enabled": true,
            "params": {}
          },
          {
            "id": "A.mounts",
            "kind": "mount_check",
            "severity": "warn",
            "enabled": true,
            "params": {
              "requiredMounts": ["${DATA_MOUNT}"]
            }
          },
          {
            "id": "A.firewall",
            "kind": "firewall_status",
            "severity": "info",
            "enabled": true,
            "params": {}
          },
          {
            "id": "A.units.present",
            "kind": "systemd_unit_exists",
            "severity": "fail",
            "enabled": true,
            "params": {
              "units": ["TODO_PRODUCT_UNIT_1"]
            }
          },
          {
            "id": "A.ports.conflict",
            "kind": "port_conflict",
            "severity": "fail",
            "enabled": true,
            "params": {
              "ports": [8080]
            }
          },
          {
            "id": "A.paths",
            "kind": "path_exists",
            "severity": "warn",
            "enabled": true,
            "params": {
              "paths": ["${DEPLOY_DIR}", "${CONFIG_DIR}", "${LOG_DIR}"]
            }
          }
        ]
      },

      {
        "id": "B",
        "name": "Baseline",
        "enabled": true,
        "purpose": "Initialize OpsKit working directories and minimal permissions needed for observability and evidence.",
        "actions": [
          {
            "id": "B.init.opskit_dirs",
            "kind": "ensure_paths",
            "enabled": true,
            "params": {
              "paths": [
                "/var/lib/opskit/state",
                "/var/lib/opskit/reports",
                "/var/lib/opskit/evidence"
              ]
            }
          },
          {
            "id": "B.snapshot.baseline",
            "kind": "baseline_snapshot",
            "enabled": true,
            "params": {
              "include": ["${DEPLOY_DIR}", "${CONFIG_DIR}", "${LOG_DIR}"]
            }
          }
        ]
      },

      {
        "id": "C",
        "name": "Deploy (Manage Declaration)",
        "enabled": true,
        "purpose": "Declare product ownership for management and generate a manage-scope deploy report (no installation).",
        "actions": [
          {
            "id": "C.declare.stack",
            "kind": "declare_stack",
            "enabled": true,
            "params": {
              "productId": "${PRODUCT_ID}",
              "productName": "${PRODUCT_NAME}",
              "scope": "manage_only"
            }
          },
          {
            "id": "C.inventory.capture",
            "kind": "capture_inventory",
            "enabled": true,
            "params": {
              "includeUnits": true,
              "includePorts": true,
              "includePaths": true
            }
          }
        ]
      },

      {
        "id": "D",
        "name": "Operate",
        "enabled": true,
        "purpose": "Periodic health checks for the product and host to compute overall status.",
        "checks": [
          {
            "id": "D.unit.active",
            "kind": "systemd_unit_active",
            "severity": "fail",
            "enabled": true,
            "params": {
              "units": ["TODO_PRODUCT_UNIT_1"]
            }
          },
          {
            "id": "D.ports.listening",
            "kind": "port_listening",
            "severity": "fail",
            "enabled": true,
            "params": {
              "ports": [8080]
            }
          },
          {
            "id": "D.resources.disk",
            "kind": "disk_usage",
            "severity": "warn",
            "enabled": true,
            "params": {
              "mounts": ["/", "${DATA_MOUNT}"],
              "warnPercent": 80,
              "failPercent": 95
            }
          },
          {
            "id": "D.resources.memory",
            "kind": "memory_available",
            "severity": "warn",
            "enabled": true,
            "params": {
              "warnMB": 1024
            }
          }
        ]
      },

      {
        "id": "E",
        "name": "Recover",
        "enabled": true,
        "purpose": "Auto-recovery on boot or on degraded/unhealthy status, with limited retries and circuit breaker.",
        "recover": {
          "enabled": true,
          "readiness": [
            { "kind": "network_ready", "params": { "requireDefaultRoute": true } },
            { "kind": "mount_ready", "params": { "requiredMounts": ["${DATA_MOUNT}"] } }
          ],
          "order": ["product"],
          "maxAttempts": 1,
          "cooldownSeconds": 600,
          "components": {
            "product": {
              "start": {
                "kind": "systemd_start",
                "params": { "units": ["TODO_PRODUCT_UNIT_1"] }
              },
              "verify": [
                {
                  "kind": "systemd_unit_active",
                  "params": { "units": ["TODO_PRODUCT_UNIT_1"] }
                },
                {
                  "kind": "port_listening",
                  "params": { "ports": [8080] }
                }
              ],
              "retryPolicy": { "attempts": 1, "waitSeconds": 10 }
            }
          },
          "failureAction": {
            "kind": "collect_bundle",
            "params": {
              "include": ["journal", "units", "ports", "df", "free", "paths"],
              "paths": ["${CONFIG_DIR}", "${LOG_DIR}"],
              "redaction": { "enabled": true }
            }
          }
        }
      },

      {
        "id": "F",
        "name": "Accept / Handover",
        "enabled": true,
        "purpose": "Generate acceptance evidence and a handover report for auditing and project acceptance.",
        "evidence": [
          {
            "id": "F.units.snapshot",
            "kind": "command_output",
            "enabled": true,
            "params": {
              "label": "systemd-units",
              "command": "systemctl status TODO_PRODUCT_UNIT_1 --no-pager"
            }
          },
          {
            "id": "F.ports.snapshot",
            "kind": "command_output",
            "enabled": true,
            "params": {
              "label": "listening-ports",
              "command": "ss -lntp"
            }
          },
          {
            "id": "F.binary.hash",
            "kind": "file_hash",
            "enabled": false,
            "params": {
              "label": "product-binary",
              "path": "TODO_PRODUCT_BINARY_OR_WAR_PATH",
              "algo": "sha256"
            }
          },
          {
            "id": "F.config.hash",
            "kind": "dir_hash",
            "enabled": false,
            "params": {
              "label": "config-dir",
              "path": "${CONFIG_DIR}",
              "algo": "sha256",
              "redaction": { "enabled": true }
            }
          },
          {
            "id": "F.process.args",
            "kind": "process_args",
            "enabled": true,
            "params": {
              "label": "process-args",
              "match": "TODO_PROCESS_MATCH"
            }
          }
        ],
        "handover": {
          "includeStages": ["A", "B", "C", "D", "E", "F"],
          "outputs": ["html", "json", "tar_gz"],
          "bundleNamePattern": "acceptance-${PRODUCT_ID}-${YYYYMMDD}"
        }
      }
    ]
  },

  "components": [
    {
      "id": "product",
      "name": "${PRODUCT_NAME}",
      "type": "systemd",
      "unit": "TODO_PRODUCT_UNIT_1",
      "ports": [8080],
      "paths": {
        "deployDir": "${DEPLOY_DIR}",
        "configDir": "${CONFIG_DIR}",
        "logDir": "${LOG_DIR}"
      },
      "healthChecks": [
        { "id": "hc.unit", "kind": "systemd_unit_active", "severity": "fail", "params": { "units": ["TODO_PRODUCT_UNIT_1"] } },
        { "id": "hc.port", "kind": "port_listening", "severity": "fail", "params": { "ports": [8080] } }
      ]
    }
  ]
}

============================================
{
  "templateVersion": "1.0",
  "stack": {
    "id": "single-service-manage",
    "name": "Single Service (Manage-only)",
    "mode": "manage",
    "criticality": "critical",
    "description": "Single host runs a single service. OpsKit manages health checking, recovery, and acceptance evidence. No installation by default.",
    "targets": {
      "os": ["Kylin V10"],
      "arch": ["aarch64", "x86_64"],
      "offline": true
    },
    "variables": {
      "SERVICE_ID": "TODO_SERVICE_ID",
      "SERVICE_NAME": "TODO_SERVICE_NAME",
      "SYSTEMD_UNIT": "TODO_SYSTEMD_UNIT",
      "PORTS": [0],
      "DATA_MOUNT": "/data",
      "DEPLOY_DIR": "TODO_DEPLOY_DIR",
      "CONFIG_DIR": "TODO_CONFIG_DIR",
      "LOG_DIR": "TODO_LOG_DIR",
      "HEALTH_HTTP_URL": "TODO_HEALTH_HTTP_URL",
      "HEALTH_HTTP_EXPECT": "TODO_HEALTH_HTTP_EXPECT",
      "PROCESS_MATCH": "TODO_PROCESS_MATCH"
    }
  },

  "lifecycle": {
    "stages": [
      {
        "id": "A",
        "name": "Preflight",
        "enabled": true,
        "purpose": "Assess whether the host is ready to run and manage the service safely.",
        "checks": [
          {
            "id": "A.sysinfo",
            "kind": "system_info",
            "severity": "info",
            "enabled": true,
            "params": {}
          },
          {
            "id": "A.mounts",
            "kind": "mount_check",
            "severity": "warn",
            "enabled": true,
            "params": {
              "requiredMounts": ["${DATA_MOUNT}"]
            }
          },
          {
            "id": "A.unit.exists",
            "kind": "systemd_unit_exists",
            "severity": "fail",
            "enabled": true,
            "params": {
              "units": ["${SYSTEMD_UNIT}"]
            }
          },
          {
            "id": "A.ports.conflict",
            "kind": "port_conflict",
            "severity": "fail",
            "enabled": true,
            "params": {
              "ports": "${PORTS}"
            }
          },
          {
            "id": "A.paths",
            "kind": "path_exists",
            "severity": "warn",
            "enabled": true,
            "params": {
              "paths": ["${DEPLOY_DIR}", "${CONFIG_DIR}", "${LOG_DIR}"]
            }
          }
        ]
      },

      {
        "id": "B",
        "name": "Baseline",
        "enabled": true,
        "purpose": "Initialize OpsKit directories and minimal observability baselines.",
        "actions": [
          {
            "id": "B.init.opskit_dirs",
            "kind": "ensure_paths",
            "enabled": true,
            "params": {
              "paths": [
                "/var/lib/opskit/state",
                "/var/lib/opskit/reports",
                "/var/lib/opskit/evidence"
              ]
            }
          },
          {
            "id": "B.snapshot.baseline",
            "kind": "baseline_snapshot",
            "enabled": true,
            "params": {
              "include": ["${DEPLOY_DIR}", "${CONFIG_DIR}", "${LOG_DIR}"]
            }
          }
        ]
      },

      {
        "id": "C",
        "name": "Deploy (Manage Declaration)",
        "enabled": true,
        "purpose": "Declare service management scope (no installation) and capture inventory.",
        "actions": [
          {
            "id": "C.declare.stack",
            "kind": "declare_stack",
            "enabled": true,
            "params": {
              "serviceId": "${SERVICE_ID}",
              "serviceName": "${SERVICE_NAME}",
              "scope": "manage_only"
            }
          },
          {
            "id": "C.inventory.capture",
            "kind": "capture_inventory",
            "enabled": true,
            "params": {
              "includeUnits": true,
              "includePorts": true,
              "includePaths": true
            }
          }
        ]
      },

      {
        "id": "D",
        "name": "Operate",
        "enabled": true,
        "purpose": "Periodic health checks for the service and basic host resources.",
        "checks": [
          {
            "id": "D.unit.active",
            "kind": "systemd_unit_active",
            "severity": "fail",
            "enabled": true,
            "params": {
              "units": ["${SYSTEMD_UNIT}"]
            }
          },
          {
            "id": "D.ports.listening",
            "kind": "port_listening",
            "severity": "fail",
            "enabled": true,
            "params": {
              "ports": "${PORTS}"
            }
          },
          {
            "id": "D.http.health",
            "kind": "http_check",
            "severity": "warn",
            "enabled": false,
            "params": {
              "url": "${HEALTH_HTTP_URL}",
              "expectContains": "${HEALTH_HTTP_EXPECT}",
              "timeoutSeconds": 3
            }
          },
          {
            "id": "D.resources.disk",
            "kind": "disk_usage",
            "severity": "warn",
            "enabled": true,
            "params": {
              "mounts": ["/", "${DATA_MOUNT}"],
              "warnPercent": 80,
              "failPercent": 95
            }
          },
          {
            "id": "D.resources.memory",
            "kind": "memory_available",
            "severity": "warn",
            "enabled": true,
            "params": {
              "warnMB": 1024
            }
          }
        ]
      },

      {
        "id": "E",
        "name": "Recover",
        "enabled": true,
        "purpose": "Auto-recovery on boot or on unhealthy status, with limited retries and circuit breaker.",
        "recover": {
          "enabled": true,
          "readiness": [
            { "kind": "network_ready", "params": { "requireDefaultRoute": true } },
            { "kind": "mount_ready", "params": { "requiredMounts": ["${DATA_MOUNT}"] } }
          ],
          "order": ["service"],
          "maxAttempts": 1,
          "cooldownSeconds": 600,
          "components": {
            "service": {
              "start": {
                "kind": "systemd_start",
                "params": { "units": ["${SYSTEMD_UNIT}"] }
              },
              "verify": [
                {
                  "kind": "systemd_unit_active",
                  "params": { "units": ["${SYSTEMD_UNIT}"] }
                },
                {
                  "kind": "port_listening",
                  "params": { "ports": "${PORTS}" }
                }
              ],
              "retryPolicy": { "attempts": 1, "waitSeconds": 10 }
            }
          },
          "failureAction": {
            "kind": "collect_bundle",
            "params": {
              "include": ["journal", "units", "ports", "df", "free", "paths"],
              "paths": ["${CONFIG_DIR}", "${LOG_DIR}"],
              "redaction": { "enabled": true }
            }
          }
        }
      },

      {
        "id": "F",
        "name": "Accept / Handover",
        "enabled": true,
        "purpose": "Generate acceptance evidence and a handover report for auditing and acceptance.",
        "evidence": [
          {
            "id": "F.unit.status",
            "kind": "command_output",
            "enabled": true,
            "params": {
              "label": "systemd-status",
              "command": "systemctl status ${SYSTEMD_UNIT} --no-pager"
            }
          },
          {
            "id": "F.ports.snapshot",
            "kind": "command_output",
            "enabled": true,
            "params": {
              "label": "listening-ports",
              "command": "ss -lntp"
            }
          },
          {
            "id": "F.version",
            "kind": "command_output",
            "enabled": false,
            "params": {
              "label": "service-version",
              "command": "TODO_VERSION_COMMAND"
            }
          },
          {
            "id": "F.binary.hash",
            "kind": "file_hash",
            "enabled": false,
            "params": {
              "label": "service-binary-or-jar",
              "path": "TODO_BINARY_OR_JAR_PATH",
              "algo": "sha256"
            }
          },
          {
            "id": "F.config.hash",
            "kind": "dir_hash",
            "enabled": false,
            "params": {
              "label": "config-dir",
              "path": "${CONFIG_DIR}",
              "algo": "sha256",
              "redaction": { "enabled": true }
            }
          },
          {
            "id": "F.process.args",
            "kind": "process_args",
            "enabled": true,
            "params": {
              "label": "process-args",
              "match": "${PROCESS_MATCH}"
            }
          }
        ],
        "handover": {
          "includeStages": ["A", "B", "C", "D", "E", "F"],
          "outputs": ["html", "json", "tar_gz"],
          "bundleNamePattern": "acceptance-${SERVICE_ID}-${YYYYMMDD}"
        }
      }
    ]
  },

  "components": [
    {
      "id": "service",
      "name": "${SERVICE_NAME}",
      "type": "systemd",
      "unit": "${SYSTEMD_UNIT}",
      "ports": "${PORTS}",
      "paths": {
        "deployDir": "${DEPLOY_DIR}",
        "configDir": "${CONFIG_DIR}",
        "logDir": "${LOG_DIR}"
      },
      "healthChecks": [
        {
          "id": "hc.unit",
          "kind": "systemd_unit_active",
          "severity": "fail",
          "params": { "units": ["${SYSTEMD_UNIT}"] }
        },
        {
          "id": "hc.port",
          "kind": "port_listening",
          "severity": "fail",
          "params": { "ports": "${PORTS}" }
        },
        {
          "id": "hc.http",
          "kind": "http_check",
          "severity": "warn",
          "enabled": false,
          "params": {
            "url": "${HEALTH_HTTP_URL}",
            "expectContains": "${HEALTH_HTTP_EXPECT}",
            "timeoutSeconds": 3
          }
        }
      ]
    }
  ]
}

快速套用：ES / OCR / MinIO 你需要填哪些

你只要针对每个服务填这些变量即可：

ES（Elasticsearch）
	•	SERVICE_ID: elasticsearch
	•	SERVICE_NAME: Elasticsearch
	•	SYSTEMD_UNIT: elasticsearch
	•	PORTS: [9200, 9300]
	•	HEALTH_HTTP_URL: http://127.0.0.1:9200/_cluster/health
	•	HEALTH_HTTP_EXPECT: "status"
	•	PROCESS_MATCH: org.elasticsearch.bootstrap.Elasticsearch
	•	可选证据：
	•	version command: curl -s http://127.0.0.1:9200 | head
	•	binary/jar 路径、config dir（/etc/elasticsearch 等）

MinIO
	•	SERVICE_ID: minio
	•	SYSTEMD_UNIT: minio（按你 unit 名）
	•	PORTS: [9000, 9001]
	•	HEALTH_HTTP_URL: http://127.0.0.1:9000/minio/health/live
	•	HEALTH_HTTP_EXPECT: ""（只要 200 即可，你也可以改为 expectStatus=200 的形式）
	•	PROCESS_MATCH: minio server
	•	config/log/data 路径按你的部署习惯填

OCR（你自己的服务或第三方）
	•	SERVICE_ID: ocr
	•	SYSTEMD_UNIT: ocr-service
	•	PORTS: [xxxx]
	•	HEALTH_HTTP_URL: 你服务的 /actuator/health 或 /health
	•	PROCESS_MATCH: jar 名或 main class


=============================================
OpsKit Deploy 模式说明

单机单服务（Single Service · Deploy）

⸻

1. Deploy 模式的定位

Deploy 模式用于 在离线、内网、信创环境下，
通过 OpsKit 自动完成单个服务的安装、配置、注册、启动、校验与交付验收。

适用场景包括但不限于：
	•	Elasticsearch（单节点）
	•	MinIO（单节点）
	•	OCR 服务（单实例）
	•	任何可通过 systemd 管理的单服务程序

Deploy 模式的目标不是“代替业务系统本身的安装文档”，
而是将 部署过程标准化、可复现、可验收、可恢复。

⸻

2. Deploy 模式的能力边界

OpsKit 负责的事情
	•	校验服务器是否满足部署条件（Preflight）
	•	准备运行环境与目录（Baseline）
	•	从 离线安装包 部署服务（Deploy）
	•	注册并管理 systemd 服务
	•	启动并验证服务健康状态
	•	在异常或断电后自动尝试恢复（Recover）
	•	生成 可审计的验收与交付证据（Accept / Handover）

OpsKit 不负责的事情
	•	修改业务逻辑配置（仅做模板化渲染）
	•	联网下载安装依赖
	•	多节点集群协调（本模板仅针对单机）
	•	业务级数据初始化（如建索引、建桶等）

⸻

3. Deploy 生命周期说明（对应 A~F 阶段）

阶段 A：Preflight（部署前检查）

目的
确认服务器满足离线部署与运行该服务的最低要求，
避免部署后才发现端口冲突、磁盘不足等问题。

主要检查内容
	•	操作系统与架构（如 Kylin V10 / aarch64）
	•	数据盘是否已挂载（如 /data）
	•	目标端口是否被占用
	•	数据盘剩余空间是否满足最低要求
	•	基础工具是否可用（tar / systemctl / ss 等）

结果
	•	若存在 FAIL 级问题，Deploy 阶段不会继续执行
	•	所有检查结果会生成 Preflight 报告并在页面展示

⸻

阶段 B：Baseline（部署基线初始化）

目的
为服务运行准备一个 规范、可控、可观测 的基础环境。

主要动作
	•	创建服务安装目录（如 /opt/deploy/<service>）
	•	创建数据目录（如 /data/<service>/data）
	•	创建日志目录（如 /data/<service>/logs）
	•	创建并校验运行用户/用户组
	•	设置必要的目录属主与权限
	•	初始化 OpsKit 自身的状态、报告、证据目录

原则
	•	不修改系统全局配置
	•	不影响已有业务
	•	所有动作可重复执行，不破坏现状

⸻

阶段 C：Deploy（核心部署阶段）

目的
在 不依赖外网 的前提下，从离线包完成服务安装与启动。

C.1 离线包校验
	•	校验离线安装包是否存在
	•	可选校验 SHA256，确保包未被篡改

C.2 解包与安装
	•	将离线包解压至安装目录
	•	保留原始包作为验收证据

C.3 配置渲染
	•	基于模板生成配置文件
	•	注入变量（数据目录、日志目录、端口等）
	•	避免手工编辑配置文件

C.4 systemd 注册
	•	安装 systemd unit 文件
	•	指定运行用户、启动命令、参数
	•	执行 daemon-reload
	•	设置开机自启

C.5 启动与验证
	•	启动 systemd 服务
	•	校验：
	•	服务是否处于 active 状态
	•	端口是否正常监听
	•	（可选）HTTP 健康检查

结果
	•	Deploy 成功后，页面显示服务为 Running
	•	Deploy 阶段报告生成并归档

⸻

阶段 D：Operate（运行巡检）

目的
在服务上线后，持续判断服务是否处于健康状态。

巡检内容
	•	systemd 服务状态
	•	端口监听状态
	•	（可选）HTTP 健康接口
	•	系统资源使用情况（磁盘/内存）

输出
	•	服务级健康状态（Healthy / Degraded / Unhealthy）
	•	Overall Status（用于页面总览）
	•	巡检记录与历史状态

⸻

阶段 E：Recover（异常与断电恢复）

目的
在服务器重启或服务异常时，自动恢复服务运行状态，减少人工介入。

触发方式
	•	服务器启动后自动触发
	•	巡检发现服务异常时触发

恢复流程
	1.	等待环境就绪（网络、数据盘）
	2.	按模板定义顺序尝试启动服务
	3.	校验服务健康
	4.	失败时有限重试（默认 1 次）
	5.	超过重试次数后熔断，避免无限循环

失败兜底
	•	自动采集诊断信息（日志、端口、磁盘等）
	•	生成恢复失败证据包，供人工分析

⸻

阶段 F：Accept / Handover（验收与交付）

目的
为项目验收、审计、交付提供 可复核的技术证据。

生成的证据包括
	•	离线安装包 hash
	•	systemd 服务状态快照
	•	配置目录 hash（脱敏）
	•	服务启动参数摘要
	•	当前监听端口快照

交付物
	•	验收证据包（tar.gz）
	•	HTML / JSON 验收报告
	•	A~F 阶段汇总交付报告

⸻

4. Deploy 模式的验收标准

Deploy 模式被视为 完成并可交付，需满足：
	•	一条命令完成部署
	•	服务成功启动并通过健康校验
	•	页面可实时反映服务状态
	•	重启服务器后服务可自动恢复
	•	可生成完整、可审计的验收证据包

⸻

5. Deploy 模式与 Manage 模式的区别
能力
Manage
Deploy
安装服务
❌
✅
离线包处理
❌
✅
systemd 注册
❌
✅
巡检
✅
✅
自动恢复
✅
✅
验收证据
✅
✅
6. 适用与扩展

该 Deploy 模式模板适用于：
	•	单机单服务
	•	离线环境
	•	systemd 管理

在此基础上可扩展为：
	•	多服务组合（Stack）
	•	集群模板（后续阶段）
	•	业务级健康与数据校验

⸻

一句话总结（可以放在文档开头）

Deploy 模式让“一次人工部署”，变成“可重复、可恢复、可验收的标准交付过程”。
