# DELIVERY_GATE（按模板交付服务器门禁）

本文是 `09-模板设计指南` 与 `ROADMAP(v0.x)` 的交付收口门禁；模板对外交付必须同时满足两者约束。

## 1. 什么叫“模板交付”（一句话）

在**单台服务器**范围内，用户选择模板并准备变量后，OpsKit 可稳定执行 `A -> D -> Accept`，并输出可复核的 `state/report/bundle`，即视为“按模板交付服务器”。

## 2. 必须满足的 8 个条件

### 条件 1：模板明确表达服务器用途

- 黑箱接管：`demo-blackbox-middleware-manage`（含 FCS/金蝶变量预设）
- 自部署中间件：`demo-minio-deploy` / `demo-elk-deploy` / `demo-powerjob-deploy`
- 产物要求：每个模板 README 必须有“接管职责一句话 + 模板不做什么”

### 条件 2：模板可独立执行 A / D / Accept

- 不依赖人工口头判断
- 可在无外部控制面的单机环境执行
- 失败允许，但必须能落盘统一状态
- 产物要求：README 提供最短命令链（`run A -> run D -> accept`）
- accept 必须可重复执行，不依赖前一次运行的中间态。

执行门禁脚本：

```bash
scripts/template-delivery-check.sh --clean
```

通过标准：

- `summary.tsv` 中每个模板 `status_json=1`
- `summary.tsv` 中每个模板 `report=1`、`bundle=1`
- 退出码允许 `0/1/3`（失败可交付）

### 条件 3：vars 只表达差异，不承载逻辑

- vars 仅承载：`unit / port / path / package` 等环境差异
- 禁止在 vars 中表达 if/else、执行顺序、策略分支
- 产物要求：`vars.example.yaml` + “缺失必填变量校验失败”示例命令

### 条件 4：模板单机自洽

- 不隐式依赖其他节点状态
- 不假设外部控制平面
- 不假设公网
- 产物要求：README 明确“单机运行前提”

### 条件 5：失败可交付

- Deploy 失败、Service 未启动仍可执行 `accept`
- 失败原因可定位，退出码稳定（`0/1/3/4`）
- 证据结构完整（`state/*.json` + `reports/*` + `bundles/*`）
- 产物要求：README 给出失败可交付说明

Exit Code Semantics:
- 0: Success, template requirements satisfied
- 1: Check failure, requirements not satisfied (deliverable)
- 3: Recoverable failure, partial requirements not satisfied (deliverable)
- 4: Execution blocked (lock / precondition)

失败场景下 bundle 结构示例（最小）：

```text
<output>/
  state/
    overall.json
    lifecycle.json
    services.json
    artifacts.json
  reports/
    preflight-*.html
    operate-*.html
    accept-*.html
  bundles/
    acceptance-*.tar.gz
  evidence/
    acceptance-consistency-*.json
```

### 条件 6：新项目接入不改引擎

- 新中间件接入优先复用现有 `deploy/manage` 模板结构
- 只新增模板、vars、必要插件
- 不新增阶段，不改 `executil` 入口
- 产物要求：接入步骤说明（见第 4 节）

### 条件 7：UI 只做模板选择与结果查看

- UI 不编辑模板
- UI 不承载执行逻辑
- UI 只展示：模板选择、状态结果、证据入口
- 产物要求：`docs/getting-started/UI_TEMPLATE_USAGE.md`

### 条件 8：README 明确三种使用模式

- 无模板：临时接管 / 排障 / 补验收
- 标准模板：按用途交付（推荐）
- 定制模板：特殊项目（高级）

## 3. 什么情况不算“模板交付”

- 仅无模板临时执行（未形成模板资产）
- 仅实验性模板（无 README、无 vars 示例、无失败可交付说明）
- 依赖人工口头判断才能解释结果

## 4. 新项目接入步骤（不改引擎，3-5 步）

1. 选择最接近的基线模板：黑箱选 `manage`，自部署选 `deploy`。  
2. 复制模板与 vars 示例，只改 `unit/port/path/package` 差异。  
3. 执行 `template validate` 与 `run A -> run D -> accept`，确认失败也有交付产物。  
4. 仅当现有插件不足时新增插件；禁止修改阶段语义和 `executil`。  
5. 补齐模板 README（用途/边界/最短命令/Delivery Level）后再进入发布。  

## 5. 对外交付声明的判定标准

只有当“目标模板集合”全部通过本门禁 8 条，并在目标环境完成一次可复核验收（含失败可交付验证），才可对外宣称：

> OpsKit 支持按模板交付服务器。

## 5.1 当前标准模板清单（v0.x）

- `generic-manage-v1`（Pilot）
- `single-service-deploy-v1`（Pilot）
- `demo-server-audit`（Demo）
- `demo-runtime-baseline`（Demo）
- `demo-blackbox-middleware-manage`（Demo，含 FCS/金蝶预设）
- `demo-generic-selfhost-deploy`（Demo）
- `demo-hello-service`（Demo）
- `demo-minio-deploy`（Demo）
- `demo-elk-deploy`（Demo）
- `demo-powerjob-deploy`（Demo）

## 6. Delivery Level 约定

- `Demo`：演示可用，边界清晰，可复核但不承诺生产参数完备
- `Pilot`：试点可用，模板契约稳定，可用于项目试运行
- `Production-Candidate`：候选生产，需补齐兼容矩阵与发布审计
- 任一门禁条件不满足，模板级别自动降级（不允许人工提升）。
