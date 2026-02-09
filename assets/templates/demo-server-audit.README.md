# demo-server-audit

`demo-server-audit.json` 是一个去生产化的通用审计模板。
该模板默认覆盖 **A / D / F** 阶段，用于验证巡检、状态输出与证据链，不包含生产部署动作。

- Delivery Level: `Demo`

## 适用范围

- 生命周期阶段：`A` / `D` / `F`
- 核心能力：主机基础巡检、运行态检查、证据输出
- D 阶段默认检查：磁盘容量、inode、只读挂载、内存、负载、NTP 同步、时钟偏移、DNS 解析

## 变量说明

- `INSTALL_ROOT`（`group=paths`，必填）：state/reports 输出根目录
- `EVIDENCE_DIR`（`group=paths`，必填）：证据输出目录

推荐使用示例变量文件：

```bash
examples/vars/demo-server-audit.json
```

若手工覆盖 `INSTALL_ROOT`，建议同时覆盖 `EVIDENCE_DIR` 以保持路径一致。

## 模板校验

文本模式：

```bash
./opskit template validate --vars-file examples/vars/demo-server-audit.json assets/templates/demo-server-audit.json
```

机器可读模式（脚本/CI 推荐）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-server-audit.json assets/templates/demo-server-audit.json
```

## 执行示例（可选）

```bash
./opskit run A --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit run D --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit accept --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
```

## 预期输出

- `state/overall.json`
- `state/lifecycle.json`
- `reports/*.html`
- `evidence/overall-json-hash.json`
- `evidence/state-dir-hash.json`

说明：服务相关检查为占位并默认关闭，避免误导为生产策略。

## 常见失败

- `run D -> FAILED`：`fs_readonly` 命中只读挂载（检查挂载参数或调整阈值）
- `status=1`：存在 FAIL 检查项；先确认 `state/` 与 `reports/` 是否已生成
- 输出目录写入失败：改用可写 `--output`（如 `/data/...` 或 `./.tmp/...`）
- `template validate --json` 返回 `template_unresolved_var`：补齐变量并通过 `--vars` 或 `--vars-file` 传入

## 交付门禁信息

### 接管职责（一句话）

该模板接管“单机通用服务器的审计与验收输出”职责（A/D/F）。

### 模板不做什么

- 不做应用部署、升级、迁移
- 不做跨节点一致性编排
- 不依赖外部控制面或公网 API

### 单机自洽前提

- 主机可本地执行基础系统命令
- `--output` 对应目录可写
- 不要求依赖其他节点状态

### 最短命令链（A -> D -> Accept）

```bash
./opskit run A --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo-server-audit
./opskit run D --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo-server-audit
./opskit accept --template assets/templates/demo-server-audit.json --vars-file examples/vars/demo-server-audit.json --output ./.tmp/opskit-demo-server-audit
```

### vars 仅表达差异（不承载逻辑）

- 变量样例：`examples/vars/demo-server-audit/vars.example.yaml`
- 变量只用于路径/主机参数差异，不用于流程分支控制

校验失败示例（缺失必填变量）：

```bash
./opskit template validate --json --vars-file examples/vars/demo-server-audit.json --vars "EVIDENCE_DIR=" assets/templates/demo-server-audit.json
```

### 失败可交付说明

即便 `run A` 或 `run D` 失败，仍可执行 `accept` 并输出标准 `state/*.json`、`reports/*`、`bundles/*`，用于审计留痕与问题移交。
