# OpsKit v0.3.2-preview.1 发布任务清单（草案）

> 主线：模板能力深化（变量系统与错误提示更可用）。

## 1. 版本目标

- 支持 `--vars-file`（文件式变量输入）
- 变量类型与枚举校验更严格（含数组/对象的最小支持）
- 模板校验报错更友好（路径化、上下文提示）
- 文档示例统一为“可直接运行”

## 2. 范围边界

### In Scope

- 模板变量系统（解析/合并/校验）
- CLI 变量输入与错误提示
- demo 模板与示例文档调整

### Out of Scope

- 生产级 deploy 模板
- 多节点/集群编排
- 登录/权限系统

## 3. 任务拆分（可执行）

### A. CLI 变量输入增强

- 新增 `--vars-file <path>`（JSON 或 key=value 格式）
- 与 `--vars` 合并：CLI 变量优先级最高
- 输出清晰错误：缺少文件、格式错误、字段类型不匹配

验收：

```bash
./opskit template validate assets/templates/demo-server-audit.json --vars-file ./examples/vars/demo-server-audit.json
```

### B. 变量类型系统（最小支持）

- `string/int/bool` 完整校验
- 支持 `array`（字符串数组）与 `object`（map[string]string）
- `enum` 与 `default` 组合校验

验收：

```bash
GOCACHE=$PWD/.gocache go test ./internal/schema ./internal/templates
```

### C. 模板校验错误提示

- 报错格式：`template.<path>: <reason>`
- 变量未解析、类型不匹配、未知字段均带路径

验收：

```bash
./opskit template validate assets/templates/demo-hello-service.json
```

### D. demo 模板与文档更新

- 更新 demo 模板变量说明（含 vars-file 示例）
- 在 README.zh-CN.md 给出可复制的 vars-file 示例

## 4. 发布准入（Go/No-Go）

- `go test ./...` 通过
- demo 模板 `template validate` 通过
- `scripts/release.sh --version v0.3.2-preview.1 --clean` 产物可校验
- 文档无敏感信息

## 5. 建议时间盒

- Day 1-2：vars-file + 类型校验
- Day 3：错误提示与 demo 文档
- Day 4：回归与发布
