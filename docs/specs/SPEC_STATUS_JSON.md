# OpsKit `status --json` 输出契约规范

## 1. 目标与范围

本规范定义 `opskit status --json` 的机器可读输出契约，供：

- 自动化验收脚本
- CI/CD 门禁
- 外部状态采集器

使用范围仅限 `status --json` 的 `stdout` JSON，不包含普通文本模式。

## 2. 调用约定

示例：

```bash
opskit status --output /data/opskit-demo --json
```

行为约定：

- `stdout` 输出 JSON
- 退出码仍按生命周期状态返回（`0|1|3`）
- 调用方必须同时读取 `stdout` 与进程退出码

## 3. 顶层字段（`schemaVersion = v1`）

字段：

- `command`：固定 `opskit status`
- `exitCode`：与该次命令退出码一致（`0|1|3`）
- `schemaVersion`：当前固定 `v1`
- `generatedAt`：状态生成时间（ISO8601）
- `overall`：等价 `state/overall.json`
- `lifecycle`：等价 `state/lifecycle.json`
- `services`：等价 `state/services.json`
- `artifacts`：等价 `state/artifacts.json`

## 4. 示例（节选）

```json
{
  "command": "opskit status",
  "exitCode": 1,
  "schemaVersion": "v1",
  "generatedAt": "2026-02-07T18:01:45+08:00",
  "overall": {
    "overallStatus": "DEGRADED",
    "lastRefreshTime": "2026-02-07T18:01:45+08:00",
    "activeTemplates": ["generic-manage-v1"],
    "openIssuesCount": 2
  }
}
```

## 5. 兼容性策略

- `v1` 内仅允许**追加字段**，不得改变现有字段语义
- 删除字段、修改类型、改变语义属于破坏性变更，必须升级 `schemaVersion`
- 调用方应忽略未知字段，避免因新增字段导致解析失败

## 6. 调用方建议

- 优先以 `exitCode` 与进程退出码交叉校验
- 对 `generatedAt` 做新鲜度判断，避免读取过旧状态
- 若退出码为 `1`，不应直接判定“程序不可用”，应结合 `lifecycle` 与 `artifacts` 判断是否已生成可用证据
