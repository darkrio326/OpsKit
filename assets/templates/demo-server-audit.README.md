# demo-server-audit.json 使用说明

用途：仅用于演示通用审计链路，默认只覆盖 **A / D / F** 阶段，不包含生产部署动作。

## 运行示例

```bash
./opskit run A --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit accept --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit status --output ./.tmp/opskit-demo
```

## 可选变量

模板内已声明 `vars` 校验，默认变量即可通过验证，必要时可覆盖：

- `INSTALL_ROOT`
- `EVIDENCE_DIR`

示例：

```bash
./opskit run A --template assets/templates/demo-server-audit.json --vars "INSTALL_ROOT=/tmp/opskit-demo"
```

说明：若显式覆盖 `INSTALL_ROOT`，建议同时覆盖 `EVIDENCE_DIR` 以保持路径一致。

## 最低可运行变量集

以下两种方式二选一即可：

- 不显式传 vars（使用模板默认路径）
- 使用 `--vars-file examples/vars/demo-server-audit.json`（推荐）

使用 `--vars-file`（推荐）：

```bash
./opskit run A \
  --template assets/templates/demo-server-audit.json \
  --vars-file examples/vars/demo-server-audit.json \
  --output ./.tmp/opskit-demo
```

## 预期输出

- `state/overall.json`
- `state/lifecycle.json`
- `reports/*.html`
- `evidence/overall-json-hash.json`
- `evidence/state-dir-hash.json`

说明：模板中的服务相关检查均为占位并默认禁用，避免误导为生产可用策略。
当前 D 阶段默认启用通用检查：磁盘容量、inode、只读挂载、内存、负载、NTP 同步、时钟偏移、DNS 解析（`localhost`）。

## 常见失败

- `run D -> FAILED`：`fs_readonly` 检查命中只读挂载（检查挂载参数或调整模板阈值）
- `status=1`：存在 FAIL 检查项，不代表程序不可用；优先确认 `state/` 与 `reports/` 已生成
- 输出目录写入失败：更换 `--output` 到可写路径（如 `/data/...` 或 `./.tmp/...`）
