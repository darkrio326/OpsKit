# OpsKit Roadmap（v0.x）

> 当前发布目标：`v0.3.x Preview`，优先开放“框架 + 通用能力 + demo 模板”。
>
> 最新发布：`v0.3.4-preview.1`（一致性索引增强、检查兼容性加固、发布门禁增强）

## Milestone 0：框架与工程骨架

- 建立 A~F 生命周期执行框架
- 完成 state JSON 基础模型与原子写
- CLI 基础命令可运行

## Milestone 1：通用巡检可视化

- A/D 阶段可执行并输出报告
- UI 可读取 `overall/lifecycle/services/artifacts` JSON
- `status` 可刷新状态并输出摘要

## Milestone 2：统一执行与安全边界

- 引入 `executil` 统一命令执行入口
- 建立全局锁，避免并发冲突（退出码 `4`）
- 基础脱敏与审计链路可用

## Milestone 3：可验收交付（已完成）

- `accept` 生成可复核证据包（manifest/hashes/reports/snapshots）
- `handover` 汇总 A~F 报告与产物
- 通用模板 + 变量体系稳定可用

## Milestone 4：模板能力增强

- 增加更多去生产化 demo 模板
- 模板校验与参数约束更严格
- 模板文档与验收规范完善

## Milestone 5：Recover 与 Operate 深化

- Recover 策略化（重试/熔断/失败采集细化）
- 通用检查项扩展（网络/时间/systemd 等）
- 页面展示恢复趋势与异常上下文

## Milestone 6：交付与生态扩展

- 多实例能力预研与目录隔离方案
- 模板仓库化与版本治理
- 发布流程标准化（release note/资产签名/兼容矩阵）
