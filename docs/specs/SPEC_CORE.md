# OpsKit 核心规范（Core Spec）

## 1. 规范范围

本规范统一约束：

- 命令行接口（CLI）
- 退出码与输出约定
- 状态 JSON 模型
- v1 设计冻结边界

## 2. CLI 规范

### 2.1 命令总览

- `opskit install`
- `opskit run <A|B|C|D|E|F|AF>`
- `opskit status`
- `opskit accept`
- `opskit handover`
- `opskit template validate <file|id> [--json]`
- `opskit template list [--json]`
- `opskit web`

### 2.2 通用参数

- `--template <id|path>`
- `--vars key=value[,key=value]`
- `--dry-run`
- `--fix`
- `--output <dir>`

### 2.3 退出码

- `0` 成功
- `1` 失败
- `2` 前置条件不满足
- `3` 部分成功（WARN）
- `4` 需人工介入（如全局锁冲突）

验收门禁口径：

- 默认 non-strict：`run A/D/accept/status` 允许 `0|1|3`
- strict（可选）：`run A/D/accept/status` 要求全 `0`
- 详细规则见：`docs/specs/SPEC_TEST_ACCEPTANCE.md`

### 2.4 输出约定

- `stdout`：人类可读摘要
- `opskit status --json`：机器可读 JSON（稳定顶层字段）
- `opskit template validate --json`：模板校验机器可读 JSON（错误路径与建议）
- `opskit template list --json`：模板目录机器可读 JSON（内置模板元数据）
- `scripts/release-check.sh --summary-json-file`：发布门禁机器可读 summary JSON
- 状态 JSON：写入 `<output>/state`
- 日志：按运行环境输出到文件或控制台

`status --json` 顶层字段约定：

- `command`：当前固定 `opskit status`
- `exitCode`：与该次 status 命令退出码一致（`0|1|3`）
- `health`：状态汇总（`ok|warn|fail`，分别对应 `0|3|1`）
- `schemaVersion`：当前固定 `v1`
- `generatedAt`：本次状态刷新时间（ISO8601）
- `overall` / `lifecycle` / `services` / `artifacts`：对应状态对象

详细契约见：`docs/specs/SPEC_STATUS_JSON.md`
模板校验 JSON 契约见：`docs/specs/SPEC_TEMPLATE_VALIDATE_JSON.md`
模板目录 JSON 契约见：`docs/specs/SPEC_TEMPLATE_LIST_JSON.md`
发布门禁 summary 契约见：`docs/specs/SPEC_RELEASE_CHECK_JSON.md`

### 2.5 模板变量约定（v1 追加）

`template.vars.<NAME>` 支持以下常用字段：

- `type`：`string|int|number|bool|array|object`
- `required`：是否必填
- `default`：默认值（需满足 `type/enum`）
- `example`：示例值（需满足 `type/enum`）
- `enum`：允许值集合
- `group`：变量分组（可选）
- `description`：变量说明

`group` 校验规则：

- 可选；为空表示未分组
- 若填写，需满足正则 `^[a-z][a-z0-9_]{0,31}$`
- 推荐组名：`service`、`paths`、`runtime`、`network`、`security`、`evidence`
- 为保证兼容性，`v1` 不限制固定枚举组名，仅限制格式

## 3. 状态模型规范

### 3.1 基本原则

- 页面只读 JSON
- JSON 字段仅追加，不破坏既有语义
- 时间格式统一 ISO8601

### 3.2 枚举约定

- `status`: `NOT_STARTED|RUNNING|PASSED|WARN|FAILED|SKIPPED`
- `severity`: `info|warn|fail`

### 3.3 必备状态文件

- `overall.json`
- `lifecycle.json`
- `services.json`
- `artifacts.json`

### 3.4 字段语义（摘要）

- `overall.json`：全局健康、刷新时间、激活模板、问题数、recover 摘要
- `lifecycle.json`：A~F 阶段状态、指标、问题、报告引用
- `services.json`：服务/组件健康明细
- `artifacts.json`：报告和 bundle 索引

## 4. v1 设计冻结边界

### 4.1 v1 范围内

- A~F 生命周期完整跑通
- 单二进制 + install
- 静态 UI + JSON 状态
- 模板支持：单服务 Deploy、单产品 Manage
- Recover（有限重试+熔断）与 Accept/Handover 证据链

### 4.2 v1 范围外

- 多节点集群编排
- 在线依赖下载
- 自动漏洞修复
- 复杂权限系统与登录平台

## 5. 兼容性原则

- 允许新增字段、模板与检查项
- 禁止修改既有字段语义
- 禁止推翻 A~F 模型
