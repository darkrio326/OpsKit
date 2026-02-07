# GitHub 发布说明（v0.3.x Preview）

## 1. 发布目标

面向开源仓库发布 OpsKit 的“框架 + 通用能力 + demo 模板”，用于离线/内网场景验证与二次开发参考。

## 2. 发布策略（允许发布）

- 核心框架：CLI、engine、state、plugin registry
- 通用能力：A/D/F 及 Recover/Accept 基础链路
- UI：读取 state JSON 的静态页面
- demo 模板：`assets/templates/demo-*.json`
- 设计与规范文档：`docs/` 下公开文档

## 3. 不发布内容（必须剔除）

- 生产级 deploy 模板（中间件一键部署脚本）
- 客户环境信息（IP、路径、账号、主机名）
- 真实日志、真实证据包、真实快照
- 客户定制策略与内部 SOP

## 4. Tag 与版本建议

- 推荐 tag：`v0.3.x-preview`（例如 `v0.3.0-preview.1`）
- 版本定位：预览版，不承诺生产 SLA

## 5. Release 资产建议

- `opskit-linux-arm64`
- `opskit-linux-amd64`
- （可选）`checksums.txt`

示例构建命令：

```bash
mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/opskit-linux-arm64 ./cmd/opskit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/opskit-linux-amd64 ./cmd/opskit
```

一键脚本（推荐）：

```bash
scripts/release.sh --version v0.3.0-preview.1 --clean
```

生成 `checksums.txt`（二选一）：

```bash
# Linux
(cd dist && sha256sum opskit-linux-arm64 opskit-linux-amd64 > checksums.txt)

# macOS
(cd dist && shasum -a 256 opskit-linux-arm64 opskit-linux-amd64 > checksums.txt)
```

校验示例：

```bash
# Linux
(cd dist && sha256sum -c checksums.txt)
```

## 6. 发布前检查清单

- 删除/排除真实环境产物：`.tmp/`、日志、证据包
- 检查文档无客户信息与敏感数据
- 验证 demo 模板可通过 `template validate`
- 至少完成一次 `run A`、`run D`、`accept` 验证流程
- 优先执行 `scripts/release-check.sh` 一键回归
- 建议在发版前增加离线门禁：`scripts/release-check.sh --with-offline-validate`
- 对“环境应全部健康”的发布，可使用严格模式：`scripts/release-check.sh --with-offline-validate --offline-strict-exit`
- 建议完成一次麒麟 V10 Docker e2e（`make -C examples/generic-manage docker-kylin-e2e`）
- 更新 `CHANGELOG.md` 对应版本条目
- 如发布二进制，确保 `checksums.txt` 已生成并随附件上传

### 6.1 `release-check` 结果快速判读

- 出现 `release-check passed` 且退出码为 `0`：门禁通过
- `release-check summary` 中 `steps` 与各步骤耗时可用于定位慢步骤
- dry-run 阶段输出中的 `checks/actions/evidence` 计数可用于确认模板执行计划未异常漂移
- 如任一步骤失败，脚本立即退出；先修复失败项，再重新执行整套门禁

## 7. Release 文案模板

- 可直接使用：`docs/RELEASE_NOTES_v0.3.0-preview.1.md`
- 当前版草稿：`docs/RELEASE_NOTES_v0.3.4-preview.1.md`
- 下一版草稿：`docs/RELEASE_NOTES_v0.3.5-preview.1.md`
- 下一版任务单：`docs/RELEASE_PLAN_v0.3.5-preview.1.md`
