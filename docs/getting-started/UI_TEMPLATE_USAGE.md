# UI 模板使用说明（只选择模板与查看结果）

## 1. 定位

OpsKit UI 仅用于：

- 选择模板视图（或无模板基础视图）
- 查看阶段状态（A-F）
- 查看证据入口（report / bundle）

OpsKit UI 不用于：

- 编辑模板
- 编辑 vars
- 承载执行逻辑

## 2. 使用步骤

1. 启动 Web：

```bash
opskit web --output /data/opskit-demo --listen 127.0.0.1:18080 --status-interval 15s
```

2. 打开页面：`http://127.0.0.1:18080/ui/`
3. 在右上角选择模板视图（或“基础通用能力”）。
4. 查看下方 A-F 卡片与状态大屏。
5. 点击证据入口进入 `reports/` 与 `bundles/`。

## 3. 自动刷新说明

- 前端自动刷新：周期拉取 `state/*.json`
- 后台状态刷新：由 `web --status-interval` 定时执行状态更新
- 页面会显示“后台状态刷新已启用/未启用”

## 4. 截图占位（建议）

![UI-模板选择与阶段卡片](docs/assets/screenshots/ui-template-stage.png)
![UI-状态大屏与证据入口](docs/assets/screenshots/ui-dashboard-evidence.png)
