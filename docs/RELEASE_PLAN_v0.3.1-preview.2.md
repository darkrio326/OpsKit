# OpsKit v0.3.1-preview.2 发布任务清单（草案）

> 目标：先做“稳定性 + 验证性”增强，不扩展生产部署边界。

## 1. 版本目标

- 统一并补齐 `run/status/accept` 退出码覆盖与测试
- 强化模板校验（变量必填、类型、枚举、未知字段提示）
- 固化 Accept 证据包可复核性（manifest/hashes/reports 一致性）
- 固化银河麒麟 v10 Docker 验证路径（脚本 + 文档）

## 2. 范围边界

### In Scope

- 通用能力与框架稳定性
- demo 模板质量与文档可用性
- 发布流程标准化（脚本、checksum、release notes）

### Out of Scope

- 生产级中间件一键部署
- 客户定制模板/环境适配
- 多节点集群与分布式调度

## 3. 任务拆分（可执行）

### A. 退出码与错误分类

- 对齐 `internal/core/exitcode` 与 CLI 返回行为
- 覆盖典型场景：参数错误、模板错误、锁冲突（4）、执行失败、部分成功
- 新增/补全单测：`cmd/opskit`、`internal/engine` 相关路径

验收：

```bash
GOCACHE=$PWD/.gocache go test ./...
```

### B. 模板校验增强

- 增加变量 schema 校验（必填、类型、枚举、默认值处理）
- 明确 unknown key 报错信息（带路径）
- 更新 demo 模板 README 示例输入

验收：

```bash
go run ./cmd/opskit template validate assets/templates/demo-server-audit.json
go run ./cmd/opskit template validate assets/templates/demo-hello-service.json
```

### C. 证据包一致性

- 增加 `accept` 结果稳定性检查（重复执行 hash 一致）
- 补充脱敏回归样例（password/token/secret）
- 校验证据目录结构与索引完整性

验收：

```bash
go run ./cmd/opskit accept --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-accept-test
```

### D. Docker 验证链路

- 固化麒麟 v10 容器执行脚本（通用 AF 路径）
- 文档同步到 `docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- 在 `docs/GITHUB_RELEASE.md` 记录最小 e2e 验证要求

## 4. 发布准入（Go/No-Go）

- 所有测试通过：`go test ./...`
- demo 模板校验通过
- `scripts/release.sh --version v0.3.1-preview.2 --clean` 产物可校验
- Release Notes、CHANGELOG、README 版本信息已更新
- 无客户信息、真实日志、真实证据包进入仓库

## 5. 建议时间盒

- Day 1-2：退出码 + 模板校验
- Day 3：证据包一致性 + 脱敏回归
- Day 4：Docker e2e + 文档
- Day 5：打包发布与回归
