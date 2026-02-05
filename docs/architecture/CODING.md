下面给你一份 到包/模块级别 的实现顺序清单，目标是：v1 完整可交付（单二进制 + install + 静态页+JSON + A~F + 单服务 Deploy 模板 + 单产品 Manage 模板）。
我按“先跑通骨架→先让页面有数据→再补 Deploy/Recover/Accept”的最省返工路线排的。

约定：你可以用 Java / Go / .NET 都行，我这里用“包/模块”抽象，不绑定语言。
目录命名给你一个建议结构，你按自己语言习惯改也行，关键是依赖方向别乱。

⸻

总体架构分层（先定依赖方向）

核心依赖链（禁止反向依赖）：
	•	core（通用模型/工具）
↓
	•	schema（JSON/模板 Schema）
↓
	•	engine（阶段执行引擎）
↓
	•	checks / actions / evidence（能力插件）
↓
	•	stages（A~F 编排）
↓
	•	templates（模板加载/变量替换）
↓
	•	reporting（HTML/JSON/包）
↓
	•	cli（命令行入口）
↓
	•	installer（install/systemd/ui）
↓
	•	ui（静态资源，仅拷贝）

⸻

v1 实现顺序清单（到包/模块级别）

0. repo 初始化与骨架（必须最先）

0.1 core/
	•	core.time：统一时间格式（ISO8601）
	•	core.fs：路径/权限/目录工具（ensurePaths、chmod/chown）
	•	core.proc：执行外部命令封装（超时、stdout/stderr、退出码）
	•	core.log：日志抽象（info/warn/error + 文件/console）
	•	core.errors：错误分类（PreconditionFailed、ActionFailed、PartialSuccess）

0.2 schema/
	•	schema.enums：Status/Severity/OverallStatus 枚举
	•	schema.state：overall/lifecycle/services/artifacts 的数据结构
	•	schema.template：template 的数据结构（只支持 v1 子集）
	•	schema.validation：基础校验（必填字段、枚举合法性）

✅ 里程碑：能在本地构造 state 对象并写出 4 个 JSON 文件（哪怕是假数据）

⸻

1. 状态输出与页面联通（先让“看得见”）

1.1 state/
	•	state.store：state 目录写入/读取（原子写：tmp→rename）
	•	state.retention：历史保留策略（先按 N 份保留）
	•	state.summarizer：从 checks 结果汇总 overallStatus

1.2 ui/（静态资源）
	•	先放一个最简静态页：读取 overall.json/lifecycle.json/services.json/artifacts.json
	•	不做复杂交互，只展示卡片和列表

✅ 里程碑：opskit status 能刷新 JSON，浏览器打开页面能看到内容变化

⸻

2. 模板系统（v1 最小子集）

2.1 templates/
	•	templates.loader：加载 template.json（从内置或外部路径）
	•	templates.vars：变量替换（${VAR}）
	•	templates.defaults：默认变量注入（INSTALL_ROOT、CONF_DIR 等）
	•	templates.merge：允许命令行 --vars 覆盖 template variables
	•	templates.validate：v1 子集校验（必填、类型）

✅ 里程碑：能加载你定义的 single-service-deploy 模板并渲染变量

⸻

3. 引擎与执行编排（把 A~F 跑起来）

3.1 engine/
	•	engine.context：运行上下文（模板、变量、输出目录、运行用户）
	•	engine.result：CheckResult/ActionResult/EvidenceResult
	•	engine.runner：顺序执行 steps，收集结果，生成 stage 结果
	•	engine.policy：失败策略（fail fast / continue on warn）

3.2 stages/
	•	stages.A_preflight
	•	stages.B_baseline
	•	stages.C_deploy（manage & deploy 分支）
	•	stages.D_operate
	•	stages.E_recover
	•	stages.F_accept_handover

✅ 里程碑：opskit run A 能生成 lifecycle.json 里 A 的状态和报告引用（报告先占位也行）

⸻

4. v1 checks/actions 插件（先覆盖最核心的）

按你模板里出现的 kind 优先实现，别多写。

4.1 checks/（v1 必做）
	•	checks.system_info
	•	checks.mount_check
	•	checks.port_conflict
	•	checks.port_listening
	•	checks.systemd_unit_exists
	•	checks.systemd_unit_active
	•	checks.disk_usage
	•	checks.memory_available
	•	（可选）checks.http_check（ES/MinIO 很有用，但可后置）

4.2 actions/（v1 必做）
	•	actions.ensure_paths
	•	actions.ensure_user_group（deploy 需要）
	•	actions.ensure_ownership
	•	actions.untar
	•	actions.sha256_verify
	•	actions.render_templates（先支持简单变量替换）
	•	actions.systemd_install_unit（把模板 unit 写入 systemd dir）
	•	actions.systemd_daemon_reload
	•	actions.systemd_enable_start
	•	actions.systemd_start / actions.systemd_stop（stop v1 默认禁用或需确认）
	•	actions.capture_inventory（units/ports/paths 摘要）
	•	actions.declare_stack（写入 stack 元信息到 state）

✅ 里程碑：用 deploy 模板能从离线 tar 包部署一个“hello-service”并拉起 systemd（哪怕只是你自己写的简单服务）

⸻

5. Recover 子系统（自愈闭环）

5.1 recover/
	•	recover.readiness：network_ready / mount_ready
	•	recover.executor：按 order 启动→verify→重试→熔断
	•	recover.circuit_breaker：cooldown 记录（写入 state）
	•	recover.onboot：生成 systemd on-boot service 需要的入口参数

5.2 collect/
	•	collect.bundle：采集 journal、ss、df、free、指定 paths 目录摘要
	•	collect.redaction：脱敏规则（基础 key/value 掩码）

✅ 里程碑：重启后 recover 自动触发一次；失败会生成 collect 包并在 artifacts.json 可见

⸻

6. Evidence / Accept / Handover（把“能卖”的闭环做出来）

6.1 evidence/
	•	evidence.command_output
	•	evidence.file_hash
	•	evidence.dir_hash（先做“文件列表+hash”，不要递归太深，支持排除）
	•	evidence.process_args（根据 PROCESS_MATCH 查进程并输出参数摘要）
	•	evidence.redaction（同 collect）

6.2 reporting/
	•	reporting.html：最简 HTML（标题+摘要+表格）
	•	reporting.json：把 stage 结果写成机器可读报告（可选）
	•	reporting.index：写入 artifacts.json（报告与包索引）
	•	reporting.bundle：打包 acceptance-YYYYMMDD.tar.gz

6.3 handover/
	•	handover.summary：汇总 A~F（状态、关键指标、报告链接）
	•	handover.generator：生成交付总报告 + 交付包

✅ 里程碑：opskit accept 生成 tar.gz，里面有 hashes、报告、快照；页面能下载

⸻

7. Installer（一键 install 成为真实产品）

7.1 installer/
	•	installer.layout：固定目录布局（state/reports/evidence/ui）
	•	installer.systemd：写入/更新 opskit-web、patrol.timer、recover.service
	•	installer.webserver：
	•	方案 A：内置静态文件 server（推荐，单二进制最干净）
	•	方案 B：生成一个极简 systemd + python -m http.server（不推荐，依赖 python）
	•	installer.bootstrap：install 后立即跑一次 A + D 生成首屏

✅ 里程碑：新服务器 scp 过去，执行 ./opskit install，页面立刻可用且有状态

⸻

8. CLI（最后封口，按文档固化）

8.1 cli/
	•	cli.root：解析命令与参数
	•	cli.commands.install
	•	cli.commands.run
	•	cli.commands.status
	•	cli.commands.accept
	•	cli.commands.handover
	•	cli.commands.template_validate
	•	cli.confirmation：危险动作二次确认（stop/disable/高风险写）

✅ 里程碑：CLI_SPEC 的命令与退出码全部达标

⸻

v1 必须交付的“模板集合”（随代码一起带）

1) 单服务 Deploy 模板（最少 2 个示例）
	•	elasticsearch-single-deploy
	•	minio-single-deploy
（OCR 可后补，但结构相同）

2) 单产品 Manage 模板（最少 1 个示例）
	•	kingdee-single-manage（或 yongzhong-fcs）

✅ 里程碑：模板 validate 通过；在页面上能切换/识别 activeTemplates

⸻

建议的实现节奏（不估时长，只给顺序）

顺序就是：
	1.	core + schema + state（先能写 JSON）
	2.	ui 联通（先能看）
	3.	templates（先能读模板）
	4.	engine + stages 骨架（先能跑 A）
	5.	checks/actions 核心集合（补 A/B/D）
	6.	deploy actions（补 C deploy）
	7.	recover 子系统（补 E）
	8.	evidence/report/bundle（补 F）
	9.	installer（补 install 一键）
	10.	cli 封口（参数、退出码、确认）

⸻

v1 完整实现的“最终验收清单”（你开发时当打勾表）
	•	opskit install 一条命令完成：目录+systemd+首屏 JSON
	•	页面显示 overall + A~F 卡片 + 服务列表 + 报告/包入口
	•	opskit run A / run D 可用
	•	Deploy 模板可从离线包部署 ES 或 MinIO，并通过 unit/port 校验
	•	重启后 Recover 自动触发，有熔断，有 collect 包
	•	opskit accept 生成验收证据 tar.gz，可复核 hash
	•	opskit handover 生成交付汇总报告
	•	CLI 退出码符合规范
	•	默认不做高风险写操作（stop/删除/改系统参数）

⸻

如果你希望我再往下“压一层”，我可以把 checks/actions/evidence 的 kind 列表按你 v1 模板实际用到的做成一个 实现优先级表（先实现哪些 kind，哪些 v2），这样你写插件时完全不会发散。


===============================================
最推荐：Go（Golang）做 v1 主栈

为什么适合 OpsKit
	•	天然单二进制：不需要 JRE / dotnet runtime / python 解释器（你最在意的“减少依赖”）。
	•	交叉编译友好：本机就能编 linux/arm64，适配麒麟服务器很舒服。
	•	标准库够用：CLI、HTTP 静态页、JSON、tar/zip、sha256、模板渲染都能用标准库搞定，第三方依赖可以做到很少。
	•	稳定性强：做运维工具，Go 这种“保守稳”的风格反而是优点。

关键实现建议（对应你 v1 路线）

目标：把“外部命令依赖”也降到最低，只在不得不用时才调用：
	•	端口/监听检查：不依赖 ss，直接读 /proc/net/tcp* 解析（完全零依赖）。
	•	内存/负载：读 /proc/meminfo、/proc/loadavg
	•	磁盘：statfs（Go syscall/unix）
	•	挂载：读 /proc/mounts
	•	默认路由：读 /proc/net/route
	•	进程参数：读 /proc/<pid>/cmdline

这样你在“极简系统”上也能跑，避免“缺少 ss/df/free”之类的依赖坑。

systemd 相关：
	•	systemctl：基本每台麒麟 V10 都有，OpsKit 要启停服务、注册 unit，调用 systemctl 是合理的最小外部依赖。
	•	journalctl：用于 collect/证据包采集，建议“可选依赖”：有就采集，没有就降级为采集日志目录文件。

静态页面：
	•	用 Go embed 把 UI 打进二进制里，opskit install 只需要把资源释放到 /usr/share/opskit/ui（或甚至不落地，直接内置 web server 提供）。
	•	内置 net/http 起一个只读静态站点，默认绑定 127.0.0.1，可配置内网 IP。

离线包处理：
	•	tar 解包：Go archive/tar + compress/gzip
	•	sha256：Go crypto/sha256
	•	渲染配置：Go text/template

Go 的“唯一注意点”
	•	争取 CGO=0（纯 Go 静态编译），减少 glibc/依赖问题。
	•	对麒麟 ARM64：GOOS=linux GOARCH=arm64 CGO_ENABLED=0

⸻

备选：Rust（更“硬核”但开发成本更高）

Rust 同样能做成单二进制、依赖少，性能也更强。但对你这个项目来说：
	•	你会花更多时间在工程细节与依赖管理上
	•	写 CLI、HTTP、模板渲染、打包等开发效率不如 Go

如果你团队 Rust 很强，Rust 也行；否则 v1 不如 Go 快。

⸻

不建议作为 v1 主栈的原因（直说）

Java
	•	你做的是“交付给服务器跑的工具”，Java 就不可避免要带 JRE 或要求服务器有 JRE（依赖大、运维麻烦、信创环境版本碎）。
	•	单文件目标很难做到“真单文件”。

.NET
	•	类似：要 runtime，离线/信创环境下 runtime 管理是额外负担。

Python/Shell
	•	初期快，后期不可控：依赖散、兼容性差、交付不可审计，很难做到你想要的“产品化交付”。

⸻

给你一个“OpsKit v1 推荐技术栈清单”（Go 版）

尽量少依赖、能落地：
	•	语言：Go
	•	CLI：优先标准库 flag（最少依赖）；如果你想体验更好再用 cobra（可选）
	•	配置：JSON（模板本身就是 JSON）；运行配置也用 JSON/YAML 你选，但我建议 v1 全 JSON，最稳
	•	Web：Go net/http + embed（内置静态页）
	•	OS 采集：尽量通过 /proc + statfs，少调用外部命令
	•	systemd：systemctl（必需外部依赖），journalctl（可选）
	•	打包：Go 内置 tar/sha256 + 你定义的 evidence/collect 目录结构
	•	日志：写文件 + 同时 stderr（便于 systemd journal 收集）

⸻

一句话建议

Go + /proc 直读 + systemctl 最小依赖 + embed 静态 UI
这是在麒麟 ARM64 上做“单文件交付运维产品”的性价比路线。

===============================================

1) Go 版 OpsKit：推荐工程目录结构

目标：单二进制、最少第三方依赖、可在麒麟 ARM64 离线跑。

opskit/
  cmd/
    opskit/
      main.go                 # 入口：解析子命令/参数
  internal/
    core/
      timeutil/               # ISO8601、时间戳
      fsutil/                 # ensure dir、原子写、权限
      procutil/               # /proc 读取：meminfo/loadavg/cmdline
      netutil/                # /proc/net route/tcp 解析、端口检查
      executil/               # 外部命令执行封装（systemctl/journalctl）
      logutil/                # 日志封装（file+stderr）
      redaction/              # 脱敏规则（key/value）
      errors/                 # 统一错误类型 + exit code mapping
    schema/
      state/                  # overall/lifecycle/services/artifacts 结构体
      template/               # template JSON 结构体（v1 子集）
      validate/               # 校验：必填字段、类型、枚举
    state/
      store/                  # 读写 state/*.json、原子写
      retention/              # 保留策略（N份/天）
      summarize/              # 汇总 overallStatus
    template/
      loader/                 # 从内置/外部加载
      vars/                   # ${VAR} 替换、--vars 覆盖、默认变量
      render/                 # 配置渲染 text/template
    engine/
      context/                # RunContext（模板、变量、路径、输出）
      runner/                 # 执行器：跑 checks/actions/evidence
      result/                 # Result 定义（pass/warn/fail）
      policy/                 # fail-fast / continue
    checks/                   # “读操作”为主
      sysinfo/
      mount/
      port/
      systemd/
      disk/
      memory/
      http/                   # 可选（v1 可先关）
    actions/                  # “写操作”为主（受安全模型约束）
      ensurepaths/
      usergroup/
      ownership/
      sha256/
      untar/
      systemd/
      inventory/
      declare/
    recover/
      readiness/              # network_ready/mount_ready
      breaker/                # 熔断状态存储
      executor/               # 启动->verify->retry
      collect/                # 失败采集包
    evidence/
      cmdout/
      filehash/
      dirhash/
      processargs/
    reporting/
      html/                   # 最简 HTML
      index/                  # 写 artifacts.json
      bundle/                 # 打包 acceptance/collect
    installer/
      layout/                 # 目录布局
      systemd/                # 写 service/timer 文件
      web/                    # 内置静态站点/或落地 ui
      bootstrap/              # install 后跑一次 A+D
    ui/
      embed/                  # Go embed 静态资源入口
  assets/
    ui/                       # 静态页面原文件（构建时 embed）
    templates/
      single-service-deploy.json
      single-product-manage.json
      es-single-deploy.json
      minio-single-deploy.json
  docs/
    product-design/
      README.md
    specs/
      README.md
      SPEC_CORE.md
      SPEC_OPS_SECURITY.md
      SPEC_TEST_ACCEPTANCE.md
      archive/                # 历史拆分规范归档（可选保留）
  Makefile (可选)
  go.mod

依赖方向（要严格守）
	•	core 最底层，不依赖其他内部包
	•	schema 只依赖 core
	•	engine/checks/actions/recover/evidence/reporting/installer 依赖 core+schema
	•	cmd/opskit 只负责组装调用，不写业务逻辑

⸻

2) 依赖策略：尽量做到“0~1 个第三方库”

v1 我建议你直接定规则：

v1 “只允许”的外部依赖
	•	无（完全标准库）是最理想
	•	如果你确实想 CLI 体验好一点：cobra（但我建议 v1 先别上）

为什么标准库足够
	•	CLI：flag + 手写 subcommand 分发
	•	HTTP 静态站点：net/http
	•	embed：embed
	•	tar/sha256/template：archive/tar、crypto/sha256、text/template
	•	JSON：encoding/json

⸻

3) 关键实现策略（麒麟 + 离线 + 少依赖）

3.1 “尽量不调用外部命令”

能读 /proc 就别 exec：
	•	端口监听：读 /proc/net/tcp、/proc/net/tcp6（解析十六进制端口）
	•	路由：读 /proc/net/route（判断默认路由）
	•	挂载：读 /proc/mounts
	•	内存：读 /proc/meminfo
	•	负载：读 /proc/loadavg
	•	进程参数：读 /proc/<pid>/cmdline

3.2 systemd 作为最小外部依赖（可接受）

这两个命令在麒麟 V10 基本都有：
	•	systemctl：启停服务、daemon-reload、enable
	•	journalctl：可选采集；没有则降级

⸻

4) Go 版“模块与阶段”映射表（写代码时不迷路）

阶段 A Preflight
	•	checks/sysinfo, checks/mount, checks/port, checks/systemd
	•	输出：reporting/html + state/store

阶段 B Baseline
	•	actions/ensurepaths, actions/usergroup, actions/ownership

阶段 C Deploy（deploy 模板）
	•	actions/sha256, actions/untar, template/render, actions/systemd
	•	verify 复用 checks/systemd, checks/port, 可选 checks/http

阶段 D Operate
	•	复用 checks：systemd/port/disk/memory
	•	state/summarize 计算 overallStatus

阶段 E Recover
	•	recover/readiness, recover/executor, recover/breaker, recover/collect

阶段 F Accept/Handover
	•	evidence/* 采集
	•	reporting/bundle 打 acceptance tar.gz
	•	reporting/index 更新 artifacts.json
	•	handover（可放到 reporting 下一个子包也行）

⸻

5) “单二进制 + embed UI”的交付形态（推荐）

你最终交付给服务器的只有一个文件：opskit
	•	UI 静态资源：放在 assets/ui，编译时用 embed 打进二进制
	•	opskit install：
	•	写 systemd unit/timer
	•	启一个内置静态站点（net/http）或把 ui 释放到 /usr/share/opskit/ui
	•	跑一次 A + D 生成首屏 JSON

这样服务器上不用 nginx、不用 node、不用 python。

⸻

6) 构建与发布建议（麒麟 ARM64）

你在开发机上直接交叉编译：
	•	GOOS=linux GOARCH=arm64 CGO_ENABLED=0
	•	输出 opskit-linux-arm64

同时再编一个 x86_64 备用：
	•	GOOS=linux GOARCH=amd64 CGO_ENABLED=0

CGO=0 能最大化减少 glibc 兼容问题，离线交付最省心。

⸻

7) v1 推荐“先做一个 ES deploy 模板”的最小落地集合

为了你写第一条 end-to-end 路线最快闭环，建议 v1 先保证这些 kind 都实现：
	•	checks：system_info、mount_check、port_conflict、systemd_unit_active、port_listening、disk_usage、memory_available
	•	actions：ensure_paths、ensure_user_group、ensure_ownership、sha256_verify、untar、render_templates、systemd_install_unit、daemon_reload、enable_start
	•	recover：network_ready、mount_ready、collect_bundle
	•	evidence：command_output、dir_hash、file_hash、process_args

===============================================
下面给你一套 Go 版 OpsKit v1 的接口设计（面向“插件化 checks/actions/evidence + A~F 阶段编排 + 状态/报告输出”）。这套接口的目标是：
	•	写插件不痛苦：每个 kind 一个实现，注册进去就能跑
	•	阶段只做编排：不关心具体实现细节
	•	结果统一：无论 check/action/evidence，都能汇总成 stage 状态与 overall 状态
	•	少依赖：全部标准库即可

我会给你：
	1.	核心类型（Status/Severity/Result）
	2.	Context（运行上下文）
	3.	Check/Action/Evidence 接口
	4.	Registry（注册与查找）
	5.	StageRunner（执行与汇总）
	6.	建议的最小实现习惯（避免返工）


1) 核心枚举与结果模型（统一出口）
package model

type Severity string

const (
	SevInfo Severity = "info"
	SevWarn Severity = "warn"
	SevFail Severity = "fail"
)

type StageStatus string

const (
	StageNotStarted StageStatus = "NOT_STARTED"
	StageRunning    StageStatus = "RUNNING"
	StagePassed     StageStatus = "PASSED"
	StageWarn       StageStatus = "WARN"
	StageFailed     StageStatus = "FAILED"
	StageSkipped    StageStatus = "SKIPPED"
)

type CheckOutcome string

const (
	OutcomePass CheckOutcome = "PASS"
	OutcomeWarn CheckOutcome = "WARN"
	OutcomeFail CheckOutcome = "FAIL"
	OutcomeSkip CheckOutcome = "SKIP"
)

type ItemResult struct {
	ID       string       `json:"id"`       // 配置里写的 check/action/evidence id
	Kind     string       `json:"kind"`     // system_info / untar / file_hash ...
	Outcome  CheckOutcome `json:"outcome"`  // PASS/WARN/FAIL/SKIP
	Severity Severity     `json:"severity"` // info/warn/fail（用于汇总策略）
	Message  string       `json:"message"`  // 简短可读结论
	Details  any          `json:"details"`  // 可选：结构化细节（端口列表、路径等）
	Error    string       `json:"error"`    // 可选：错误文本（不要放堆栈）
	Started  string       `json:"started"`  // ISO8601
	Ended    string       `json:"ended"`    // ISO8601
}

type StageResult struct {
	StageID   string       `json:"stageId"`   // A/B/C...
	Name      string       `json:"name"`
	Status    StageStatus  `json:"status"`
	LastRun   string       `json:"lastRun"`
	Metrics   []Metric     `json:"metrics"`   // 页面卡片的 2~4 个关键指标
	Issues    []Issue      `json:"issues"`    // WARN/FAIL 的摘要
	Items     []ItemResult `json:"items"`     // 原始条目结果（可用于详情页）
	ReportRef string       `json:"reportRef"` // artifacts 索引 id 或路径
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Issue struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

2) RunContext（执行上下文：模板变量、路径、执行策略）

package engine

import "context"

type RunContext struct {
	Ctx context.Context

	// 运行标识
	RunID     string // 每次 run 的唯一 id
	StageID   string // 当前阶段
	TemplateID string

	// 变量（已完成默认值注入与覆盖合并）
	Vars map[string]string

	// 文件系统布局
	StateDir   string // /var/lib/opskit/state
	ReportsDir string // /var/lib/opskit/reports
	EvidenceDir string // /var/lib/opskit/evidence
	CacheDir   string // /var/lib/opskit/cache
	UIDir      string // /usr/share/opskit/ui (可选)

	// 执行策略
	DryRun bool
	Fix    bool
	Force  bool

	// 安全策略（后面 SECURITY_MODEL 会用到）
	AllowWriteLevel int // 0..3（v1 先用 0/1/2）
}

建议：Vars 最终都变成 string，模板里复杂结构（ports 数组等）由具体插件自己解析（JSON array -> []int），避免 Context 变复杂。

3) 配置节点（从 Template JSON 解析出来的“待执行项”）

package plan

import "encoding/json"

// Node 是模板里一个 check/action/evidence 的抽象
type Node struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`
	Enabled  bool            `json:"enabled"`
	Severity string          `json:"severity"` // info/warn/fail（check/evidence 用；action 也可用于汇总）
	Params   json.RawMessage `json:"params"`   // 交给具体插件 decode
}

4) Check / Action / Evidence 接口（插件化核心）

4.1 Check 接口
package plugin

import (
	"encoding/json"
	"opskit/internal/engine"
	"opskit/internal/model"
)

type Check interface {
	Kind() string
	Run(rc *engine.RunContext, nodeID string, severity model.Severity, params json.RawMessage) model.ItemResult
}

4.2 Action 接口

package plugin

import (
	"encoding/json"
	"opskit/internal/engine"
	"opskit/internal/model"
)

type Action interface {
	Kind() string
	Run(rc *engine.RunContext, nodeID string, severity model.Severity, params json.RawMessage) model.ItemResult
}

4.3 Evidence 接口
package plugin

import (
	"encoding/json"
	"opskit/internal/engine"
	"opskit/internal/model"
)

type Evidence interface {
	Kind() string
	Collect(rc *engine.RunContext, nodeID string, severity model.Severity, params json.RawMessage) model.ItemResult
}

为什么 check/action/evidence 都返回 ItemResult：阶段汇总逻辑完全一致；你只需要在 Stage 编排中决定跑哪些 node。

5) Registry（按 kind 查找实现；阶段执行只依赖 registry）
package registry

import (
	"fmt"
	"opskit/internal/plugin"
)

type Registry struct {
	checks   map[string]plugin.Check
	actions  map[string]plugin.Action
	evidence map[string]plugin.Evidence
}

func New() *Registry {
	return &Registry{
		checks:   map[string]plugin.Check{},
		actions:  map[string]plugin.Action{},
		evidence: map[string]plugin.Evidence{},
	}
}

func (r *Registry) RegisterCheck(c plugin.Check) {
	r.checks[c.Kind()] = c
}
func (r *Registry) RegisterAction(a plugin.Action) {
	r.actions[a.Kind()] = a
}
func (r *Registry) RegisterEvidence(e plugin.Evidence) {
	r.evidence[e.Kind()] = e
}

func (r *Registry) GetCheck(kind string) (plugin.Check, error) {
	c, ok := r.checks[kind]
	if !ok {
		return nil, fmt.Errorf("check kind not found: %s", kind)
	}
	return c, nil
}
func (r *Registry) GetAction(kind string) (plugin.Action, error) {
	a, ok := r.actions[kind]
	if !ok {
		return nil, fmt.Errorf("action kind not found: %s", kind)
	}
	return a, nil
}
func (r *Registry) GetEvidence(kind string) (plugin.Evidence, error) {
	e, ok := r.evidence[kind]
	if !ok {
		return nil, fmt.Errorf("evidence kind not found: %s", kind)
	}
	return e, nil
}

v1 先用 map 就够。后面要做“别名 / 版本”再扩展。

⸻

6) StageRunner：执行 + 汇总（最关键的“稳定内核”）
6.1 执行策略接口（可选，但很有用）
package engine

import "opskit/internal/model"

type Policy interface {
	ShouldContinue(item model.ItemResult) bool
}

type DefaultPolicy struct{}

func (p DefaultPolicy) ShouldContinue(item model.ItemResult) bool {
	// v1 简化：FAIL 且 severity=fail -> 直接停止
	if item.Outcome == model.OutcomeFail && item.Severity == model.SevFail {
		return false
	}
	return true
}

6.2 Runner（分别执行 checks/actions/evidence）
package engine

import (
	"opskit/internal/model"
	"opskit/internal/plan"
	"opskit/internal/registry"
)

type StageRunner struct {
	Reg    *registry.Registry
	Policy Policy
}

func (sr *StageRunner) RunChecks(rc *RunContext, nodes []plan.Node) []model.ItemResult {
	var out []model.ItemResult
	for _, n := range nodes {
		if !n.Enabled {
			continue
		}
		c, err := sr.Reg.GetCheck(n.Kind)
		if err != nil {
			out = append(out, model.ItemResult{
				ID: n.ID, Kind: n.Kind, Outcome: model.OutcomeFail, Severity: model.SevFail,
				Message: "check kind 未注册", Error: err.Error(),
			})
			continue
		}
		item := c.Run(rc, n.ID, toSeverity(n.Severity), n.Params)
		out = append(out, item)
		if sr.Policy != nil && !sr.Policy.ShouldContinue(item) {
			break
		}
	}
	return out
}

func (sr *StageRunner) RunActions(rc *RunContext, nodes []plan.Node) []model.ItemResult {
	var out []model.ItemResult
	for _, n := range nodes {
		if !n.Enabled {
			continue
		}
		a, err := sr.Reg.GetAction(n.Kind)
		if err != nil {
			out = append(out, model.ItemResult{
				ID: n.ID, Kind: n.Kind, Outcome: model.OutcomeFail, Severity: model.SevFail,
				Message: "action kind 未注册", Error: err.Error(),
			})
			continue
		}
		item := a.Run(rc, n.ID, toSeverity(n.Severity), n.Params)
		out = append(out, item)
		if sr.Policy != nil && !sr.Policy.ShouldContinue(item) {
			break
		}
	}
	return out
}

func (sr *StageRunner) CollectEvidence(rc *RunContext, nodes []plan.Node) []model.ItemResult {
	var out []model.ItemResult
	for _, n := range nodes {
		if !n.Enabled {
			continue
		}
		e, err := sr.Reg.GetEvidence(n.Kind)
		if err != nil {
			out = append(out, model.ItemResult{
				ID: n.ID, Kind: n.Kind, Outcome: model.OutcomeFail, Severity: model.SevFail,
				Message: "evidence kind 未注册", Error: err.Error(),
			})
			continue
		}
		item := e.Collect(rc, n.ID, toSeverity(n.Severity), n.Params)
		out = append(out, item)
		// evidence 失败一般不阻断全流程（你也可以用 policy 控制）
	}
	return out
}

func toSeverity(s string) model.Severity {
	switch s {
	case "info":
		return model.SevInfo
	case "warn":
		return model.SevWarn
	default:
		return model.SevFail
	}
}

6.3 汇总 StageStatus（统一算法）
package engine

import "opskit/internal/model"

// 汇总规则建议：
// - 任意 OutcomeFail 且 Severity=fail => StageFailed
// - 否则任意 OutcomeWarn 或 OutcomeFail(但 severity=warn/info) => StageWarn
// - 否则 StagePassed
func SummarizeStage(stageID, name string, items []model.ItemResult) model.StageResult {
	res := model.StageResult{
		StageID: stageID,
		Name:    name,
		Status:  model.StagePassed,
		Items:   items,
	}

	hasWarn := false
	for _, it := range items {
		if it.Outcome == model.OutcomeFail && it.Severity == model.SevFail {
			res.Status = model.StageFailed
			res.Issues = append(res.Issues, model.Issue{ID: it.ID, Severity: it.Severity, Message: it.Message})
			continue
		}
		if it.Outcome == model.OutcomeWarn || (it.Outcome == model.OutcomeFail && it.Severity != model.SevFail) {
			hasWarn = true
			res.Issues = append(res.Issues, model.Issue{ID: it.ID, Severity: it.Severity, Message: it.Message})
		}
	}

	if res.Status != model.StageFailed && hasWarn {
		res.Status = model.StageWarn
	}
	return res
}
7) 插件实现习惯（强烈建议你现在就统一）

7.1 插件都做 “params decode + rc.Vars 替换”
	params 用 json.RawMessage，在插件内部定义 struct：
    type Params struct {
	Ports []int `json:"ports"`
}
需要变量替换（${PORTS} 这种）你可以在 template 层先做替换；或者插件里再做一次“string -> parse”。

7.2 ItemResult 的 message 要“短 + 明确”
	•	✅ “端口 9200 未监听”
	•	✅ “unit elasticsearch inactive”
	•	❌ “error happened”

7.3 禁止在插件里直接写 state/report
	•	插件只产出 ItemResult
	•	state/report 统一在 stage/runner 层写（否则很快变成一团）

⸻

8) v1 最小 kind 对应插件名（你写文件时不纠结）
	•	checks:
	•	system_info
	•	mount_check
	•	port_conflict
	•	port_listening
	•	systemd_unit_exists
	•	systemd_unit_active
	•	disk_usage
	•	memory_available
	•	http_check（可选）
	•	actions:
	•	ensure_paths
	•	ensure_user_group
	•	ensure_ownership
	•	sha256_verify
	•	untar
	•	render_templates
	•	systemd_install_unit
	•	systemd_daemon_reload
	•	systemd_enable_start
	•	systemd_start
	•	evidence:
	•	command_output
	•	file_hash
	•	dir_hash
	•	process_args

===============================================================
下面给你一套Template JSON → Node 解析结构的设计（Go 版），目标是：
	•	模板 JSON 保持你前面定义的结构（A~F stages 内有 checks/actions/evidence/deploy/recover）
	•	解析后得到一个统一的 ExecutionPlan：每个 stage 拿到要跑的 []plan.Node
	•	支持 ${VAR} 替换（在解析或加载阶段完成）
	•	v1 只支持你最小子集：stages[].checks/actions/evidence + deploy.steps/verify + recover.components.*.verify（先够用）

我会给你：
	1.	模板结构体（反序列化层）
	2.	解析后的 Plan 结构体（执行层）
	3.	解析流程（TemplateLoader → Normalizer → PlanBuilder）
	4.	关键实现细节（变量替换与 JSON RawMessage）

⸻

1) Template JSON 反序列化结构（template/schema 层）

这层的职责：把 JSON 读进来，字段尽量贴近 JSON，不要做业务逻辑。
package tpl

import "encoding/json"

type Template struct {
	TemplateVersion string `json:"templateVersion"`
	Stack           Stack  `json:"stack"`
	Lifecycle       Lifecycle `json:"lifecycle"`
	Components      []Component `json:"components,omitempty"`
}

type Stack struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Mode        string              `json:"mode"`        // manage | deploy
	Criticality string              `json:"criticality"` // critical | noncritical
	Description string              `json:"description,omitempty"`
	Targets     map[string]any      `json:"targets,omitempty"`
	Variables   map[string]any      `json:"variables,omitempty"` // 原始类型可能是 string/number/array
}

type Lifecycle struct {
	Stages []Stage `json:"stages"`
}

type Stage struct {
	ID      string `json:"id"`   // A/B/C/D/E/F
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Purpose string `json:"purpose,omitempty"`

	Checks   []Node `json:"checks,omitempty"`
	Actions  []Node `json:"actions,omitempty"`
	Evidence []Node `json:"evidence,omitempty"`

	Deploy  *DeployBlock  `json:"deploy,omitempty"`  // 仅 C 用
	Recover *RecoverBlock `json:"recover,omitempty"` // 仅 E 用
	Handover *HandoverBlock `json:"handover,omitempty"` // 仅 F 用（可选）
}

type Node struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`
	Enabled  *bool           `json:"enabled,omitempty"`  // nil -> 默认 true
	Severity string          `json:"severity,omitempty"` // info/warn/fail（可为空 -> 默认 fail）
	Params   json.RawMessage `json:"params,omitempty"`
	// v1 先不做 timeout/retry，后续可扩展
}

// ---- Deploy Block（C 阶段）----
type DeployBlock struct {
	Enabled bool `json:"enabled"`
	Inputs  []DeployInput `json:"inputs,omitempty"`
	Steps   []Node `json:"steps,omitempty"`   // 直接复用 Node：untar/render/systemd...
	Verify  []Node `json:"verify,omitempty"`  // 复用 Node：systemd_unit_active/port_listening/http_check
}

type DeployInput struct {
	ID       string          `json:"id"`
	Kind     string          `json:"kind"`        // bundle_present/sha256_verify
	Required bool            `json:"required"`
	Params   json.RawMessage `json:"params,omitempty"`
}

// ---- Recover Block（E 阶段）----
type RecoverBlock struct {
	Enabled bool `json:"enabled"`
	Readiness []Node `json:"readiness,omitempty"` // kind: network_ready/mount_ready
	Order   []string `json:"order,omitempty"`
	MaxAttempts int  `json:"maxAttempts,omitempty"`
	CooldownSeconds int `json:"cooldownSeconds,omitempty"`

	Components map[string]RecoverComponent `json:"components,omitempty"`

	FailureAction *Node `json:"failureAction,omitempty"` // kind: collect_bundle
}

type RecoverComponent struct {
	Start Node   `json:"start"`
	Verify []Node `json:"verify,omitempty"`
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
}

type RetryPolicy struct {
	Attempts int `json:"attempts"`
	WaitSeconds int `json:"waitSeconds"`
}

type HandoverBlock struct {
	IncludeStages []string `json:"includeStages,omitempty"`
	Outputs []string `json:"outputs,omitempty"` // html/json/tar_gz
	BundleNamePattern string `json:"bundleNamePattern,omitempty"`
}

type Component struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit,omitempty"`
	Ports any   `json:"ports,omitempty"` // 可能是数组或变量字符串
	Paths map[string]any `json:"paths,omitempty"`
}

要点：
	•	Node.Params 用 json.RawMessage 保留原始参数结构，交给插件自己 decode。
	•	Enabled 用 *bool，这样能区分“没写（默认 true）” vs “写 false”。

2) 执行层 Plan 结构（engine/plan 层）

这层的职责：提供“某个 stage 要跑哪些 nodes”，并且已经做完：

	•	stage enabled 判定
	•	node enabled 默认值处理
	•	severity 默认值处理
	•	（可选）变量替换已完成
package plan

import "encoding/json"

type ExecutionPlan struct {
	TemplateID string
	Mode       string // manage|deploy
	Stages     map[string]StagePlan // key: "A".."F"
	// recover/deploy 的结构化部分也可以放这里，供 E/C 阶段使用
	Deploy  *DeployPlan
	Recover *RecoverPlan
	Handover *HandoverPlan
}

type Node struct {
	ID       string
	Kind     string
	Enabled  bool
	Severity string          // "info"|"warn"|"fail"
	Params   json.RawMessage
	// 可选：SourceStage/Group 用于调试
	Group   string          // "checks"|"actions"|"evidence"|"deploy.steps"...
}

type StagePlan struct {
	StageID string
	Name    string
	Enabled bool

	Checks   []Node
	Actions  []Node
	Evidence []Node
}

type DeployPlan struct {
	Enabled bool
	Inputs  []DeployInputPlan
	Steps   []Node
	Verify  []Node
}

type DeployInputPlan struct {
	ID string
	Kind string
	Required bool
	Params json.RawMessage
}

type RecoverPlan struct {
	Enabled bool
	Readiness []Node
	Order []string
	MaxAttempts int
	CooldownSeconds int
	Components map[string]RecoverComponentPlan
	FailureAction *Node
}

type RecoverComponentPlan struct {
	Start Node
	Verify []Node
	RetryPolicy *RetryPolicyPlan
}

type RetryPolicyPlan struct {
	Attempts int
	WaitSeconds int
}

type HandoverPlan struct {
	IncludeStages []string
	Outputs []string
	BundleNamePattern string
}

3) 解析流程：Template → NormalizedTemplate → ExecutionPlan

3.1 Loader：读 JSON + 变量替换（建议在这里做）
	•	读取 template.json 字符串
	•	合并变量：
	•	template.stack.variables
	•	默认变量（INSTALL_ROOT 等）
	•	CLI --vars 覆盖
	•	将 ${VAR} 替换到整个 JSON 文本（字符串级替换）
✅ 最简单、v1 足够
⚠️ 要注意：替换后仍是合法 JSON（例如 ${PORTS} 如果是 [9200,9300] 这种数组，就必须确保替换位置不是带引号的字符串里）

强烈建议的 v1 规则：
	•	数组/对象变量在 JSON 里不要加引号，例如：
	•	"ports": ${PORTS}（PORTS = [9200,9300]）
	•	这样替换后仍是合法 JSON

3.2 Normalizer：补默认值（enabled/severity）
	•	stage.enabled 缺省 -> true（或模板里强制写 enabled）
	•	node.enabled 缺省 -> true
	•	node.severity 缺省 -> fail（除非明确写 warn/info）

3.3 PlanBuilder：把 Stage/Deploy/Recover 提取成 ExecutionPlan
	•	遍历 template.lifecycle.stages
	•	为每个 stage 建 StagePlan
	•	同时把 stage.Deploy、stage.Recover 提取到 plan.Deploy / plan.Recover（便于 C/E 阶段使用）

⸻

4) PlanBuilder 伪代码（可直接照着写）
func BuildPlan(t *tpl.Template) (*plan.ExecutionPlan, error) {
	p := &plan.ExecutionPlan{
		TemplateID: t.Stack.ID,
		Mode:       t.Stack.Mode,
		Stages:     map[string]plan.StagePlan{},
	}

	for _, s := range t.Lifecycle.Stages {
		stageEnabled := s.Enabled // 建议模板强制写 enabled；否则这里可默认 true
		sp := plan.StagePlan{
			StageID: s.ID,
			Name:    s.Name,
			Enabled: stageEnabled,
			Checks:  normalizeNodes(s.Checks, "checks"),
			Actions: normalizeNodes(s.Actions, "actions"),
			Evidence: normalizeNodes(s.Evidence, "evidence"),
		}
		p.Stages[s.ID] = sp

		// 提取 Deploy / Recover / Handover
		if s.Deploy != nil {
			p.Deploy = &plan.DeployPlan{
				Enabled: s.Deploy.Enabled,
				Inputs:  normalizeDeployInputs(s.Deploy.Inputs),
				Steps:   normalizeNodes(s.Deploy.Steps, "deploy.steps"),
				Verify:  normalizeNodes(s.Deploy.Verify, "deploy.verify"),
			}
		}
		if s.Recover != nil {
			p.Recover = normalizeRecover(s.Recover)
		}
		if s.Handover != nil {
			p.Handover = &plan.HandoverPlan{
				IncludeStages: s.Handover.IncludeStages,
				Outputs: s.Handover.Outputs,
				BundleNamePattern: s.Handover.BundleNamePattern,
			}
		}
	}

	return p, nil
}

func normalizeNodes(nodes []tpl.Node, group string) []plan.Node {
	out := make([]plan.Node, 0, len(nodes))
	for _, n := range nodes {
		enabled := true
		if n.Enabled != nil {
			enabled = *n.Enabled
		}
		sev := n.Severity
		if sev == "" {
			sev = "fail"
		}
		out = append(out, plan.Node{
			ID: n.ID, Kind: n.Kind, Enabled: enabled, Severity: sev, Params: n.Params, Group: group,
		})
	}
	return out
}

func normalizeDeployInputs(in []tpl.DeployInput) []plan.DeployInputPlan {
	out := make([]plan.DeployInputPlan, 0, len(in))
	for _, x := range in {
		out = append(out, plan.DeployInputPlan{
			ID: x.ID, Kind: x.Kind, Required: x.Required, Params: x.Params,
		})
	}
	return out
}

func normalizeRecover(r *tpl.RecoverBlock) *plan.RecoverPlan {
	out := &plan.RecoverPlan{
		Enabled: r.Enabled,
		Readiness: normalizeNodes(r.Readiness, "recover.readiness"),
		Order: r.Order,
		MaxAttempts: r.MaxAttempts,
		CooldownSeconds: r.CooldownSeconds,
		Components: map[string]plan.RecoverComponentPlan{},
	}

	for name, c := range r.Components {
		out.Components[name] = plan.RecoverComponentPlan{
			Start: plan.Node{
				ID: c.Start.ID, Kind: c.Start.Kind, Enabled: true,
				Severity: defaultSev(c.Start.Severity), Params: c.Start.Params, Group: "recover.start",
			},
			Verify: normalizeNodes(c.Verify, "recover.verify"),
			RetryPolicy: &plan.RetryPolicyPlan{
				Attempts: safeInt(c.RetryPolicy, "Attempts", 1),
				WaitSeconds: safeInt(c.RetryPolicy, "WaitSeconds", 10),
			},
		}
	}

	if r.FailureAction != nil {
		n := *r.FailureAction
		out.FailureAction = &plan.Node{
			ID: n.ID, Kind: n.Kind, Enabled: true,
			Severity: defaultSev(n.Severity), Params: n.Params, Group: "recover.failureAction",
		}
	}
	// 默认值
	if out.MaxAttempts == 0 { out.MaxAttempts = 1 }
	if out.CooldownSeconds == 0 { out.CooldownSeconds = 600 }
	return out
}

func defaultSev(s string) string {
	if s == "" { return "fail" }
	return s
}

注意：Recover 的 Start 节点你可以要求模板里强制写 enabled=false/true，但 v1 简化为默认 enabled=true。

⸻

5) “阶段怎么取 nodes”——标准做法（stages 编排层）
	•	stages.A_preflight：
	•	plan.Stages["A"].Checks → runner.RunChecks()
	•	stages.B_baseline：
	•	plan.Stages["B"].Actions → runner.RunActions()
	•	stages.C_deploy：
	•	如果 mode=manage：跑 plan.Stages["C"].Actions（declare_stack/inventory）
	•	如果 mode=deploy：跑 plan.Deploy.Inputs（校验）→ plan.Deploy.Steps（actions）→ plan.Deploy.Verify（checks）
	•	stages.D_operate：
	•	plan.Stages["D"].Checks
	•	stages.E_recover：
	•	plan.Recover.Readiness（checks 或者专用 readiness runner）
	•	按 plan.Recover.Order 执行 plan.Recover.Components[name].Start（action）→ .Verify（checks）
	•	stages.F_accept：
	•	plan.Stages["F"].Evidence → runner.CollectEvidence()
	•	然后 reporting.bundle/handover

⸻

6) v1 变量替换的“强规则”（避免 JSON 解析地狱）

为了让你实现最简单且稳定，我建议写进 SPEC：
	1.	非字符串类型变量（数组/对象/数字）必须在 JSON 里不加引号
	•	✅ "ports": ${PORTS}
	•	❌ "ports": "${PORTS}"
	2.	${VAR} 替换发生在 JSON 反序列化前（文本级替换）
	3.	变量缺失时：
	•	v1 默认：替换为 "" 或直接报错（我建议 报错，否则很难排障）

⸻
==============================================

下面给你一套 统一外部命令执行器 executil 的设计（Go / v1），专门服务 OpsKit 的场景：
	•	调 systemctl / journalctl / tar / sha256sum（可选）等
	•	超时、可取消、输出限流、脱敏、审计记录、退出码映射
	•	最重要：减少依赖 & 行为可控（别让插件各写各的 exec）

我按“接口 + 数据结构 + 默认策略 + 如何被插件使用”给你。

⸻

1) 设计目标与边界

目标
	•	插件只管“要跑什么命令”，不管超时/输出/脱敏/日志
	•	所有外部命令有统一审计：何时、谁、跑了啥、耗时、退出码、摘要输出
	•	可根据安全模型控制：禁止危险命令、必须二次确认、只读模式等

v1 边界
	•	只做 本机执行
	•	不做 SSH 远程（v2 再加）
	•	不做 shell 管道（尽量避免 bash -c）；必要时也集中控制

⸻

2) 核心类型设计

2.1 CommandSpec：命令定义（插件传入）
package executil

type CommandSpec struct {
	// 必填：建议传绝对路径或在 allowlist 中（例如 systemctl/journalctl）
	Path string   // "systemctl"
	Args []string // ["status", "nginx", "--no-pager"]

	// 行为控制
	TimeoutSeconds int  // 0 -> 用默认
	WorkingDir     string
	Env            map[string]string
	RunAsUser      string // v1 可先不实现（一般 opskit root 跑）
	UseShell       bool   // v1 默认 false；true 时执行 "bash -lc <cmd>"
	ShellCommand   string // 仅当 UseShell=true 使用

	// 输出控制
	MaxOutputBytes int    // 0 -> 默认，例如 1MB
	TrimOutput     bool   // true：输出过长截断
	// 脱敏
	Redact         bool
	RedactKeys     []string // 额外脱敏 key（如 password, token）
	// 审计标签
	Tag            string   // e.g. "systemd.status", "collect.journal"
}

建议 v1 规则
	•	插件优先使用 Path+Args，禁止直接拼 bash -c，除非模板里明确允许的极少数情况。
	•	Tag 必填（方便排障和审计）

⸻

2.2 ExecResult：执行结果（统一返回）
package executil

type ExecResult struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
	Tag  string   `json:"tag"`

	ExitCode int    `json:"exitCode"`
	DurationMs int64 `json:"durationMs"`
	TimedOut bool   `json:"timedOut"`

	Stdout string `json:"stdout"` // 可能截断
	Stderr string `json:"stderr"` // 可能截断
	// 便于日志展示：stdout+stderr 摘要（比如前 200 字）
	Summary string `json:"summary"`

	// 失败原因（非进程退出码层面）
	Error string `json:"error,omitempty"`
}

2.3 Executor：统一执行器接口
package executil

import "context"

type Executor interface {
	Run(ctx context.Context, spec CommandSpec) ExecResult
}

3) 安全与审计：Policy / Auditor

3.1 Policy：执行前检查（危险命令管控）
package executil

type PolicyDecision struct {
	Allowed bool
	Reason  string
	RequireConfirm bool // v1 可先落到 cli 层；executil 只给标记
}

type Policy interface {
	Check(spec CommandSpec) PolicyDecision
}

v1 默认 Policy（建议）
	•	allowlist（只允许这些可执行文件）：
	•	systemctl, journalctl, tar, sha256sum, ss, df, free（你若改为 /proc 直读就不需要 ss/df/free）
	•	denylist：
	•	rm, dd, mkfs, iptables, firewall-cmd, shutdown, reboot
	•	任何 UseShell=true → RequireConfirm=true 或直接 deny（你决定）

3.2 Auditor：审计记录（统一落盘）
package executil

type Auditor interface {
	Record(event AuditEvent)
}

type AuditEvent struct {
	TimeISO   string `json:"time"`
	StageID   string `json:"stageId"`
	TemplateID string `json:"templateId"`
	RunID     string `json:"runId"`

	Tag   string   `json:"tag"`
	Path  string   `json:"path"`
	Args  []string `json:"args"`
	ExitCode int   `json:"exitCode"`
	DurationMs int64 `json:"durationMs"`
	TimedOut bool   `json:"timedOut"`
	// 输出不要全量写审计（避免泄露/爆盘），写摘要
	Summary string `json:"summary"`
	Denied  bool   `json:"denied"`
	Reason  string `json:"reason,omitempty"`
}

审计建议写到：/var/log/opskit/exec-audit.jsonl（一行一个 JSON）

⸻

4) DefaultExecutor：实现要点（超时、限流、脱敏、截断）

4.1 默认参数
	•	TimeoutSeconds 默认 10（status 类）/ 60（deploy 解包类可更长，插件可覆盖）
	•	MaxOutputBytes 默认 1MB（collect/journal 可以更大但要谨慎）
	•	输出截断策略：截断尾部或截断中间都行，v1 简单截断尾部即可

4.2 输出限流实现建议

不要 CombinedOutput()（可能撑爆内存），而是：
	•	StdoutPipe / StderrPipe
	•	并发读，累计到 bytes buffer
	•	超过 MaxOutputBytes 就停止继续读（但要 drain / kill 进程避免阻塞）
	•	最简单做法：读到上限后，继续丢弃输出直到进程退出（需要 goroutine）

4.3 脱敏（Redaction）

v1 脱敏建议两层：
	1.	基于 key 的简单替换（正则可选）：(?i)(password|token|secret)\s*[:=]\s*\S+ → ***
	2.	对于命令行参数：如果 args 包含 --password / -p 之类，审计时只写 --password=***

⸻

5) 将 executil 融入 RunContext（插件使用方式）

在 engine.RunContext 增加：
type RunContext struct {
  ...
  Exec executil.Executor
  ExecPolicy executil.Policy
  Auditor executil.Auditor
}
或者把 policy/auditor 组合到 executor 里也行。推荐组合成一个：
type SecureExecutor struct {
  Inner executil.Executor
  Policy executil.Policy
  Auditor executil.Auditor
  // run metadata
  StageID string
  TemplateID string
  RunID string
}
6) 典型插件怎么用（示例）

systemd_unit_active check
spec := executil.CommandSpec{
  Path: "systemctl",
  Args: []string{"is-active", unit},
  TimeoutSeconds: 5,
  MaxOutputBytes: 8 * 1024,
  Redact: false,
  Tag: "systemd.is-active",
}
r := rc.Exec.Run(rc.Ctx, spec)

if r.ExitCode == 0 {
  // PASS
} else {
  // FAIL，message: unit inactive
}

systemd_enable_start action
// enable
rc.Exec.Run(rc.Ctx, executil.CommandSpec{
  Path: "systemctl",
  Args: []string{"enable", unit},
  TimeoutSeconds: 10,
  Tag: "systemd.enable",
})
// start
rc.Exec.Run(rc.Ctx, executil.CommandSpec{
  Path: "systemctl",
  Args: []string{"start", unit},
  TimeoutSeconds: 30,
  Tag: "systemd.start",
})

collect journal（可选依赖）
spec := executil.CommandSpec{
  Path: "journalctl",
  Args: []string{"-u", unit, "--no-pager", "--since", "today"},
  TimeoutSeconds: 10,
  MaxOutputBytes: 2 * 1024 * 1024,
  Redact: true,
  Tag: "collect.journal",
}
r := rc.Exec.Run(rc.Ctx, spec)
// r.Stdout 写入 evidence/collect 中的文件（注意已截断）
7) v1 强约束（建议写进 SECURITY_MODEL）

为了减少“脚本注入/不可控”，我建议 v1 定死：
	1.	默认禁止 UseShell=true（除非 --force 或专门白名单）
	2.	外部命令必须在 allowlist 内（systemctl/journalctl/tar/sha256sum）
	3.	对 stop/disable/restart 这种“可造成服务中断”的动作，policy 标记 RequireConfirm=true，由 CLI 做二次确认

⸻

8) 你可以直接照着落的“executil 最小模块清单”
	•	executil/types.go：CommandSpec/ExecResult
	•	executil/policy.go：AllowlistPolicy/Denylist
	•	executil/audit.go：JSONL Auditor
	•	executil/executor.go：DefaultExecutor（超时/限流/截断）
	•	executil/redact.go：Redaction（regex + args mask）


    =================================

    必补设计（建议在开码前定死）

1) 运行时目录布局与文件命名规范

你后面所有模块都会用到路径，不先定死会全仓返工。

必须明确：
	•	根目录：/var/lib/opskit/（state/reports/evidence/cache）
	•	state 文件：是否覆盖 or 按 runId 分版本？
	•	报告命名：reports/<stage>-<template>-<timestamp>.html
	•	证据包命名：bundles/acceptance-<service>-<YYYYMMDD>.tar.gz
	•	collect 包命名：bundles/collect-<stage>-<timestamp>.tar.gz
	•	保留策略：每类最多保留 N 份（默认 20）+ 超出删除规则

这份设计写成 1 页 “PATHS.md” 就够。

⸻

2) 执行状态机与并发/互斥策略（非常关键）

你会有 timer 周期巡检、手动 run、recover 开机触发——不做互斥会乱套。

必须定：
	•	是否允许并发 run？（我建议 v1 不允许）
	•	全局锁：/var/lib/opskit/state/lock（flock）
	•	冲突策略：
	•	已在运行 → 返回 exit 4（或 3）并提示“正在执行”
	•	“Recover 与 Deploy 冲突”的优先级：Deploy 优先，Recover 看到锁就退出

⸻

3) Kind/Params 的版本与兼容策略

你模板会长期演进，必须提前定“兼容规则”。

建议定：
	•	kind 名称永远不改；新增用新 kind
	•	params：只追加字段；旧字段保留语义
	•	templateVersion：v1 只支持 1.0，v2 再扩展
	•	未识别 kind 的处理：
	•	Check：FAIL（模板无效）
	•	Evidence：WARN（可降级）
	•	Action：FAIL（部署不完整）

⸻

4) 阶段语义与“失败是否阻断”规则（Policy）

你现在有 A~F，但每个阶段失败后该不该继续必须定。

建议 v1 规则（实现友好）：
	•	A（Preflight）FAIL → 阻断后续（exit 2）
	•	B（Baseline）FAIL → 阻断 C（deploy）
	•	C（Deploy）FAIL → 阻断 D/E/F（但仍可做 collect）
	•	D（Operate）FAIL → 不阻断（它本来就是判活）
	•	E（Recover）FAIL → 不阻断（但写入熔断与 collect）
	•	F（Accept）FAIL → 不阻断（部分证据失败仍可出包，但标记 WARN/FAILED）

⸻

5) 变量系统的规范（字符串 vs JSON 值）

你前面说 ${PORTS} 这种替换，必须写进规范，不然后面模板会写崩。

必须定：
	•	变量类型：
	•	String 变量：会自动 JSON escape 并加引号（如果你实现这个）
	•	JSON 变量：原样注入（数组/对象/数字），模板中必须不加引号
	•	缺失变量：直接 FAIL（而不是替换为空）
	•	变量优先级：
	•	CLI --vars > vars-file > 模板 variables > 默认变量

这会影响 loader/渲染/模板库，必须先写。

⸻

6) 服务标识与“模板实例化”规则

同一个模板可能部署多次（不同端口/目录/实例）。你得定义“实例 ID”。

建议：
	•	templateId 是“模板”
	•	instanceId 是“实例”（默认等于 serviceId）
	•	state/report/bundle 都按 instanceId 分目录，避免覆盖：
	•	/var/lib/opskit/instances/<instanceId>/state/...

如果你 v1 不做多实例，也要写清：v1 默认只支持单实例，但目录结构预留。

⸻

7) systemd 单元生成规范（最容易扯皮）

deploy 模板要装 unit，你要提前定：
	•	unit 放哪里：/etc/systemd/system/<unit>.service
	•	unit 模板变量：RUN_USER/SERVICE_BIN/SERVICE_ARGS/CONF_DIR/DATA_DIR/LOG_DIR
	•	unit 里必须带的：
	•	Restart=on-failure
	•	RestartSec=...
	•	LimitNOFILE=...（ES/MinIO 常用）
	•	日志策略：
	•	默认交给 journal，或同时写文件（你决定）

这份写成 SYSTEMD_GUIDE.md。

⸻

8) “证据包内容清单”与脱敏边界（验收要用）

要提前定：包里包含什么、什么必须脱敏、什么不能采集。

建议定 v1 证据包必须包含：
	•	manifest.json（包内容索引、生成时间、模板/实例信息）
	•	hashes.txt（sha256）
	•	reports/*.html
	•	snapshots/systemd.txt、snapshots/ports.txt、snapshots/proc_args.txt
	•	conf_hash.json（只 hash，不直接打包明文配置，除非你允许）

脱敏规则清单要写死（regex + key list）。

⸻

9) Web/UI 访问模型（最小安全设计）

你不做登录，但也得写清：
	•	默认只监听 127.0.0.1
	•	是否允许绑定内网 IP（需要显式配置）
	•	UI 只读，不提供危险操作按钮（v1）
	•	下载证据包是否需要 token？（v1 可以不做，但至少写明风险）

⸻

10) “离线包规范”与模板库约定（Deploy 的核心）

要卖 deploy 模板，离线包结构必须标准化，不然每次都写特例。

建议定义离线 bundle 标准（v1）：
	•	bundle.tar.gz
	•	bin/（可执行或启动脚本）
	•	systemd/<unit>.service（模板）
	•	templates/（配置模板）
	•	manifest.json（版本、构建时间、目标 OS/arch）
	•	SHA256 文件：bundle.tar.gz.sha256（可选）

写成 BUNDLE_SPEC.md，你以后 ES/MinIO/OCR 都按这个打包。

⸻

建议你在开码前补齐的“最小设计文档清单”（10 选 6）

如果你只想补最关键的 6 个，我建议：
	1.	PATHS.md（目录与命名）
	2.	LOCKING.md（互斥与并发）
	3.	POLICY.md（阶段失败阻断规则）
	4.	VARS.md（变量系统）
	5.	BUNDLE_SPEC.md（离线包标准）
	6.	SYSTEMD_GUIDE.md（unit 生成规范）


========
PATHS.md

OpsKit 运行时目录与命名规范（v1）

1. 设计目标
	•	所有模块统一路径来源
	•	不因模板/阶段不同导致路径分叉
	•	支持回溯、审计、清理

2. 根目录布局（v1 固定）前面有更完整的更详细的 这个参考
/var/lib/opskit/
├─ instances/
│  └─ default/
│     ├─ state/
│     ├─ reports/
│     ├─ evidence/
│     ├─ bundles/
│     └─ cache/
└─ logs/
   └─ opskit.log
v1 仅支持单实例，实例名固定为 default，但目录结构预留多实例能力。

⸻

3. state 目录
	•	用途：页面 JSON 状态源
	•	文件：
	•	overall.json
	•	lifecycle.json
	•	services.json
	•	artifacts.json
	•	写入策略：
	•	原子写（.tmp → rename）
	•	每次 run 覆盖当前状态

⸻

4. reports 目录
	•	用途：阶段性 HTML/JSON 报告
	•	命名规范：
	•	<stage>-<template>-<YYYYMMDDHHmmss>.html
	•	例：A-preflight-es-20260201103000.html

⸻

5. bundles / evidence
	•	bundles/：最终可交付 tar.gz
	•	evidence/：中间产物（hash、snapshot）
	•	命名：
	•	acceptance-<instance>-<YYYYMMDD>.tar.gz
	•	collect-<stage>-<YYYYMMDDHHmmss>.tar.gz

⸻

6. 保留策略（v1）
	•	每类最多保留 20 份
	•	超出删除最旧
	•	bundles 默认不自动清理（避免误删交付物）

⸻

LOCKING.md

执行互斥与并发控制（v1）

1. 设计目标
	•	避免 Deploy / Recover / Patrol 并发执行
	•	保证状态文件一致性

⸻

2. 全局锁
	•	锁文件：
	•	/var/lib/opskit/instances/default/state/opskit.lock
	•	实现方式：
	•	文件锁（flock / syscall.Flock）

⸻

3. 互斥规则（v1）
	•	任意时刻 只允许一个 opskit run
	•	包括：
	•	opskit run
	•	opskit accept
	•	opskit handover
	•	systemd timer 触发的 patrol / recover

⸻

4. 冲突处理
	•	若锁已被占用：
	•	CLI 返回 exit code = 4
	•	stdout 提示：
OpsKit is already running (stage=X, pid=Y)

5. 特殊优先级
	•	Deploy > Recover > Patrol
	•	Recover/Patrol 发现锁存在：
	•	直接退出
	•	不做重试

⸻

VARS.md

模板变量系统规范（v1）

1. 设计目标
	•	简单、可预测
	•	避免 JSON 替换错误

⸻

2. 变量来源与优先级
	1.	CLI --vars key=value
	2.	CLI --vars-file
	3.	template stack.variables
	4.	系统默认变量

高优先级覆盖低优先级。

⸻

3. 变量类型（强约束）

3.1 String 变量
	•	用于路径、命令参数
	•	示例：
"unit": "${SYSTEMD_UNIT}"
3.2 JSON 变量（数组 / 对象 / 数字）
	•	模板中禁止加引号
	•	示例：
"ports": ${PORTS}
PORTS = [9200, 9300]
4. 替换规则
	•	发生在 JSON 反序列化之前（文本级）
	•	未定义变量 → 直接 FAIL
	•	不支持嵌套表达式（v1）

⸻

5. 默认系统变量（v1）
	•	INSTANCE_ID = default
	•	DATA_MOUNT = /data
	•	INSTALL_ROOT = /opt/deploy/${SERVICE_ID}

⸻

POLICY.md

生命周期阶段执行与阻断策略（v1）

1. 设计目标
	•	行为确定
	•	不靠“感觉”决定是否继续

⸻

2. 阶段阻断规则
阶段
FAIL 行为
A Preflight
阻断后续，exit 2
B Baseline
阻断 C
C Deploy
阻断 D/E/F
D Operate
不阻断
E Recover
不阻断
F Accept
不阻断

3. Severity 影响
	•	severity=fail + FAIL → 阻断
	•	severity=warn + FAIL → 记录 WARN，不阻断
	•	info 永不阻断

⸻

4. Recover 特殊规则
	•	Recover 永不触发 Deploy
	•	Recover 失败：
	•	记录熔断状态
	•	生成 collect 包

⸻

BUNDLE_SPEC.md

离线部署包规范（v1）

1. 设计目标
	•	统一 deploy 模板输入
	•	避免每个服务一套安装逻辑

⸻

2. 标准结构
bundle.tar.gz
├─ bin/
│  └─ service-binary
├─ systemd/
│  └─ <unit>.service
├─ templates/
│  └─ *.conf.tpl
└─ manifest.json

3. manifest.json（最小字段）
{
  "name": "elasticsearch",
  "version": "7.17.18",
  "os": "kylin-v10",
  "arch": "aarch64",
  "buildTime": "2026-02-01T10:00:00+08:00"
}

4. 校验规则
	•	bundle 文件必须存在
	•	若提供 sha256 → 必须匹配
	•	OS/arch 不匹配 → WARN（v1 不强阻断）

⸻

SYSTEMD_GUIDE.md

systemd 单元生成与管理规范（v1）

1. 安装位置
	•	/etc/systemd/system/<unit>.service

⸻

2. 必须包含的字段
	•	User=
	•	ExecStart=
	•	Restart=on-failure
	•	RestartSec=10
	•	LimitNOFILE=65536（可模板化）

⸻

3. 日志策略
	•	默认交由 journal
	•	日志目录由服务自身写入 ${LOG_DIR}

⸻

4. OpsKit 管理原则
	•	enable + start 由 OpsKit 执行
	•	stop / disable：
	•	v1 禁止自动执行
	•	必须人工确认（CLI）

⸻

✅ 到此为止，你已经：
	•	补齐所有“不可返工”的设计空洞
	•	可以放心进入编码阶段，不会被路径/变量/并发/阶段语义反复折磨
