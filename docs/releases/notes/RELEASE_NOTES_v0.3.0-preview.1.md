# OpsKit v0.3.0-preview.1

> 预览版本（Preview）。面向离线/内网场景验证，不承诺生产 SLA。

## 🚀 本次发布亮点

- 完成 Milestone 3（通用服务器“可验收”）核心链路
- A~F 生命周期框架可运行（Preflight / Baseline / Deploy / Operate / Recover / Accept-Handover）
- 统一 state JSON + 静态 UI 状态页（overall/lifecycle/services/artifacts）
- 模板驱动执行（模板加载、变量替换、阶段编排）
- 统一命令执行器（executil）与基础审计边界
- Accept 证据包能力闭环（manifest + hashes + reports + snapshots）
- 全局锁并发保护（冲突返回 exit code `4`）

## 📦 Release 资产

- `opskit-linux-arm64`
- `opskit-linux-amd64`
- （可选）`checksums.txt`

## ⚡ 快速体验

```bash
# 1) 构建
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/opskit-linux-arm64 ./cmd/opskit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/opskit-linux-amd64 ./cmd/opskit

# 2) 最小运行链路
go build -o opskit ./cmd/opskit
./opskit run A --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit run D --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo
./opskit accept --template assets/templates/demo-server-audit.json --output ./.tmp/opskit-demo

# 3) 查看状态页
./opskit web --output ./.tmp/opskit-demo --listen 127.0.0.1:18080
```

## ✅ 当前能力范围

- 通用巡检、状态汇总、证据生成
- 去生产化 demo 模板（`assets/templates/`）
- 文档化规范（产品设计、规格、安全、发布说明）

## ⚠️ 当前不包含/不承诺

- 生产级中间件一键部署模板
- 客户定制模板与客户环境适配脚本
- 登录权限系统（账号/RBAC）
- 多节点集群编排与分布式协调

## 🔒 安全与合规提示

- 默认建议仅本机监听 UI（`127.0.0.1`）
- 高风险动作需人工确认（如 stop/disable）
- 发布内容不包含客户信息、生产模板、真实日志/证据包

## 🧭 后续计划

- Milestone 4：模板能力增强（更多 demo 模板与校验约束）
- Milestone 5：Recover/Operate 深化（策略化恢复与更丰富检查）
- Milestone 6：交付与生态扩展（模板库化、多实例预研）

---

问题反馈与建议欢迎提交 Issue/PR。
