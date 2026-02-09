# OpsKit v0.4.1-preview.1 发布任务清单

> 目标：完成 M4 的“模板交付收口”，冻结 v0.4.x 设计边界，并为后续模板扩展开发提供稳定基线。

## 1. 版本目标

- 收口模板交付门禁（Delivery Gate）并固化评审口径
- 统一模板文档与变量样例规范（用途、边界、最短命令）
- 冻结 v0.4.x 设计边界：不新增阶段、不改 executil 语义
- 保持“无模板/标准模板/定制模板”三种模式入口清晰

## 2. 范围边界

### In Scope

- 新增 `v0.4.x` 设计冻结说明
- 文档入口统一到 `v0.4.1-preview.1`
- 模板 README 与 vars 示例口径统一
- Release 计划与说明滚动到新版本

### Out of Scope

- 生产级中间件部署脚本
- 多节点编排与集群控制面
- 新生命周期阶段与核心执行语义变更

## 3. 功能点拆分（固定 5 项）

### 功能点 1：版本入口切换（完成）

- 新增 `docs/releases/plans/RELEASE_PLAN_v0.4.1-preview.1.md`
- 新增 `docs/releases/notes/RELEASE_NOTES_v0.4.1-preview.1.md`
- README/docs 索引切换到当前版本入口

验收：

```bash
rg -n "v0.4.1-preview.1" README.md README.zh-CN.md docs/README.md docs/releases/README.md docs/GITHUB_RELEASE.md
```

### 功能点 2：v0.4.x 设计冻结（完成）

- 新增 `docs/product-design/10-v0.4.x设计冻结说明.md`
- 明确冻结范围、允许变更、禁止变更、变更申请流程
- 产品设计索引增加冻结文档入口

验收：

```bash
rg -n "冻结|Freeze|变更申请" docs/product-design/10-v0.4.x设计冻结说明.md docs/product-design/README.md
```

### 功能点 3：模板交付门禁收口（完成）

- `DELIVERY_GATE` 作为模板交付评审门禁
- 继续沿用“失败可交付”口径（`0/1/3` 可交付，`4` 阻断）
- 对模板 README 要求追加 `Delivery Level` 约定

验收：

```bash
scripts/template-delivery-check.sh --clean
```

### 功能点 4：文档与发布流程一致性（完成）

- `ROADMAP` 与 `GITHUB_RELEASE` 同步到 `v0.4.1-preview.1`
- 归档历史 release 文档，根目录仅保留当前版本 plan/notes

验收：

```bash
ls -1 docs/RELEASE_PLAN_v*.md docs/RELEASE_NOTES_v*.md
```

### 功能点 5：M4 后续开发基线（完成）

- 固化“先黑箱管理模板，再自部署模板”的推进顺序
- 预留 0.4.x 后续 preview 仅做模板库扩展与插件补齐

验收：

```bash
rg -n "M4|黑箱|自部署|模板交付" ROADMAP.md docs/releases/notes/RELEASE_NOTES_v0.4.1-preview.1.md
```

## 4. 发布准入（Go/No-Go）

- `go test ./...` 全绿
- `scripts/release-check.sh --with-offline-validate` 通过
- `scripts/template-delivery-check.sh --clean` 通过
- 文档版本口径一致（README/ROADMAP/GITHUB_RELEASE/RELEASE_NOTES）

## 5. 交付物

- 当前版本计划：`docs/releases/plans/RELEASE_PLAN_v0.4.1-preview.1.md`
- 当前版本说明：`docs/releases/notes/RELEASE_NOTES_v0.4.1-preview.1.md`
- 设计冻结文档：`docs/product-design/10-v0.4.x设计冻结说明.md`
