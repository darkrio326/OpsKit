# 银河麒麟 V10 离线部署与使用（从 GitHub Release 开始）

本文面向离线服务器场景：服务器不能访问外网，需先在联网机器下载 OpsKit Release 包，再拷贝到麒麟 V10 服务器运行。

## 1. 适用范围

- 操作系统：银河麒麟 V10（`linux/amd64` 或 `linux/arm64`）
- 网络：目标服务器离线
- 目标：跑通 `run A` / `run D` / `accept`，生成 `state`、`reports`、`bundles`，并通过 `web` 查看状态

## 2. 在联网机器下载 Release 资产

以 `v0.3.6` 为例（发布时请替换为实际版本）：

```bash
VERSION=v0.3.6
```

1. 打开 Release 页面：
   - `https://github.com/darkrio326/OpsKit/releases/tag/${VERSION}`
2. 下载以下文件（按服务器架构选择二进制）：
   - `opskit-${VERSION}-linux-amd64` 或 `opskit-${VERSION}-linux-arm64`
   - `checksums.txt`
   - `release-metadata.json`

可选：在联网机器先校验哈希（推荐）

```bash
cd /path/to/downloads
sha256sum -c checksums.txt
```

如果是 macOS：

```bash
# 手动比对 checksums.txt 中对应行
shasum -a 256 opskit-${VERSION}-linux-amd64
```

可选：查看 `release-metadata.json`，确认 `gitCommit/goVersion/generatedAt` 与发布说明一致。

## 3. 传输到离线麒麟服务器并安装

将下载好的文件通过 U 盘或内网文件传输到服务器，例如 `/tmp/opskit-release`。

```bash
# 1) 进入目录
cd /tmp/opskit-release

# 2) 识别服务器架构
uname -m
# x86_64 -> 用 amd64
# aarch64 -> 用 arm64

# 3) 安装二进制（以 amd64 为例）
install -m 0755 opskit-${VERSION}-linux-amd64 /usr/local/bin/opskit

# 4) 验证
opskit --help
```

无 root 权限可安装到当前用户目录：

```bash
mkdir -p "$HOME/bin"
cp opskit-${VERSION}-linux-amd64 "$HOME/bin/opskit"
chmod +x "$HOME/bin/opskit"
export PATH="$HOME/bin:$PATH"
opskit --help
```

## 4. 最小可运行流程（离线可用）

创建输出目录：

```bash
mkdir -p /data/opskit-demo
```

先用内置模板跑一条最小链路（不依赖仓库文件）：

```bash
opskit install --template generic-manage-v1 --dry-run --no-systemd --output /data/opskit-demo
opskit run A --template generic-manage-v1 --dry-run --output /data/opskit-demo
opskit run D --template generic-manage-v1 --dry-run --output /data/opskit-demo
opskit accept --template generic-manage-v1 --dry-run --output /data/opskit-demo
opskit status --output /data/opskit-demo
```

如果需要执行非 dry-run（真实检查/动作），去掉 `--dry-run`。

## 4.1 `v0.3.x` 用户侧回归清单（推荐）

建议在麒麟离线机按以下顺序做一次完整回归：

```bash
scripts/kylin-offline-validate.sh \
  --bin /usr/local/bin/opskit \
  --output /data/opskit-regression \
  --json-status-file /data/opskit-regression/status.json \
  --summary-json-file /data/opskit-regression/summary.json \
  --clean
```

严格模式（要求 `run A/D/accept/status` 均 exit=0）：

```bash
scripts/kylin-offline-validate.sh \
  --bin /usr/local/bin/opskit \
  --output /data/opskit-regression \
  --json-status-file /data/opskit-regression/status.json \
  --summary-json-file /data/opskit-regression/summary.json \
  --strict-exit \
  --clean
```

说明：

- 离线首轮验收建议先使用默认模式（允许 `0/1/3`），确认状态、报告、证据产物链路完整。
- 严格模式适合“应全绿”的目标环境；若首轮即要求全绿，需先完成环境基线治理（挂载、服务、时间同步等）。

若需手工逐步执行，可使用以下命令：

```bash
# 0) 输出目录
export OPSKIT_OUT=/data/opskit-regression
mkdir -p "$OPSKIT_OUT"

# 1) 模板可用性（离线建议用内置模板 ID）
opskit template validate generic-manage-v1 --output "$OPSKIT_OUT"

# 2) A / D / accept（真实执行）
opskit run A --template generic-manage-v1 --output "$OPSKIT_OUT"
opskit run D --template generic-manage-v1 --output "$OPSKIT_OUT"
opskit accept --template generic-manage-v1 --output "$OPSKIT_OUT"

# 3) 刷新状态并检查退出码
opskit status --output "$OPSKIT_OUT"
echo "status exit=$?"
```

回归通过建议至少满足：

- `opskit status` 退出码为 `0/1/3`（`1` 代表存在 FAIL，`3` 代表存在 WARN）
- `status.json` 中 `health` 与退出码匹配：`ok=0`、`warn=3`、`fail=1`
- `state/lifecycle.json` 含 `summary(total/pass/warn/fail/skip)` 字段
- `state/artifacts.json` 存在 `acceptance-consistency-*.json` 报告索引
- `reports/accept-*.html` 可打开并看到 consistency 摘要
- `status.json` 存在且包含 `command/schemaVersion/exitCode/health`

可选检查命令（系统有 `grep` 即可）：

```bash
grep -R "\"summary\"" "$OPSKIT_OUT/state/lifecycle.json"
grep -R "acceptance-consistency-" "$OPSKIT_OUT/state/artifacts.json"
grep -R "\"consistency\"" "$OPSKIT_OUT/reports"/accept-*.html
grep -R "\"schemaVersion\"" "$OPSKIT_OUT/status.json"
grep -R "\"exitCode\"" "$OPSKIT_OUT/status.json"
grep -R "\"health\"" "$OPSKIT_OUT/status.json"
```

## 5. 启动 Web UI 查看状态

```bash
opskit web --output /data/opskit-demo --listen 127.0.0.1:18080
```

本机访问：`http://127.0.0.1:18080`

远程访问建议用 SSH 隧道：

```bash
ssh -L 18080:127.0.0.1:18080 user@kylin-host
```

然后在本地浏览器打开：`http://127.0.0.1:18080`

## 6. 产物说明

运行后重点查看：

- `/data/opskit-demo/state/overall.json`
- `/data/opskit-demo/state/lifecycle.json`
- `/data/opskit-demo/state/services.json`
- `/data/opskit-demo/state/artifacts.json`
- `/data/opskit-demo/status.json`（一键回归脚本默认输出）
- `/data/opskit-demo/reports/*.html`
- `/data/opskit-demo/bundles/*.tar.gz`
- `/data/opskit-demo/ui/index.html`

## 7. 常见问题与解决方案

1. `Permission denied`（二进制无法执行）

```bash
chmod +x /usr/local/bin/opskit
```

2. `Exec format error`（架构不匹配）

- 原因：在 `x86_64` 机器上运行了 `arm64` 二进制，或反之
- 处理：用 `uname -m` 确认架构，重新替换对应文件

3. `command not found: opskit`

- 原因：安装目录不在 `PATH`
- 处理：使用绝对路径运行，或把安装目录加入 `PATH`

4. `template ... not found` / `template validate` 失败

- 离线仅有二进制时，请优先使用内置模板 ID：`generic-manage-v1`
- 若使用外部模板路径，需确认模板文件已随包拷贝到服务器
- 若仅复制了二进制但没有仓库文件，`template validate templates/builtin/...` 会失败，此时请改用模板 ID

5. `run A` / `run D` 返回失败或 WARN

- 这通常是环境检查触发（例如端口扫描工具缺失、挂载点不满足、systemd 不可用）
- 先执行：

```bash
opskit status --output /data/opskit-demo
```

- 再查看 `state/lifecycle.json` 中对应 stage 的 `issues` 与 `reports/*.html`

6. `exit code 4`（并发冲突）

- 原因：已有另一个 OpsKit 进程在同一输出目录运行
- 处理：等待前一任务完成，或改用不同 `--output` 目录

7. Web UI 打不开

- 检查进程是否启动，端口是否监听
- 确认使用 `http://` 而不是 `https://`
- 远程访问请使用 SSH 隧道，或按安全策略显式开放监听地址

8. `acceptance-consistency` 没有出现在 `artifacts.json`

- 先确认是否执行了 `opskit accept`
- 再检查 `--output` 是否和 `status/web` 使用的是同一目录
- 若中途清理过目录，请重新执行 `run A/D` + `accept` + `status`

9. `status` 返回 `1`

- 这通常表示某些检查项为 FAIL（例如挂载/端口/systemd 条件不满足）
- 对离线回归来说，优先确认状态文件、报告和一致性产物已生成，再决定是否调整模板阈值

## 8. 功能使用速查

- 模板检查：`opskit template validate <template.json>`
- 安装阶段：`opskit install ...`
- 阶段执行：`opskit run <A|B|C|D|E|F|AF> ...`
- 状态刷新：`opskit status --output <dir>`
- 验收打包：`opskit accept ...`
- 移交包：`opskit handover --output <dir>`
- UI 服务：`opskit web --output <dir> --listen 127.0.0.1:18080`

## 9. 安全建议

- 默认使用 `127.0.0.1` 监听 UI
- 不要在仓库或共享目录保留真实客户数据、日志、证据包
- 生产环境使用前请完成本地安全与合规评估
