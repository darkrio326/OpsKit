# demo-server-audit

`demo-server-audit.json` 是一个去生产化的通用审计模板。
该模板默认覆盖 **A / D / F** 阶段，用于验证巡检、状态输出与证据链，不包含生产部署动作。

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
