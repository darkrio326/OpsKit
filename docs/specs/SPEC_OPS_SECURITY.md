# OpsKit 运行与安全规范（Ops & Security Spec）

## 1. 交付与安装

### 1.1 交付目标

- 单二进制 `opskit`
- 离线可运行

### 1.2 install 最小行为

`opskit install` 至少完成：

- 创建运行目录（state/reports/evidence/bundles/cache/ui）
- 安装/更新静态 UI
- 可选安装 systemd 单元
- 初始化首屏状态（建议 A + D）

### 1.3 卸载原则

- 停止并禁用 opskit 相关服务/定时器
- 删除 opskit 目录（可选择保留 evidence/bundles）

## 2. 容器验证（麒麟 V10）

- 参考部署说明：`docs/deployment/DOCKER_KYLIN_V10_DEPLOY.md`
- 标准入口：`make -C examples/generic-manage docker-kylin-e2e`
- 产物检查：`state/*.json`、`reports/*`、`bundles/*`、`summary.json`

## 3. 权限与操作分级

- `Level 0` 只读操作（状态采集、查询）
- `Level 1` 安全写操作（OpsKit 自身目录与状态写入）
- `Level 2` 中风险操作（systemd unit 安装、enable/start）
- `Level 3` 高风险操作（删除数据、系统参数变更）v1 默认禁止

## 4. 执行与审计约束

- 所有外部命令应统一通过 `executil` 执行
- 统一超时、输出截断、错误映射
- 关键操作必须可审计（命令、结果、时间、阶段）
- 禁止插件直接执行未受控命令

## 5. 脱敏规范

- 默认脱敏关键字：`password`、`token`、`secret`
- 配置输出优先 hash 或脱敏快照
- 报告与证据包不得明文暴露敏感信息

## 6. 失败边界与人工确认

- 高风险动作默认不自动执行
- 需要人工确认的动作必须在 CLI 层显式触发
- 并发冲突与锁占用必须显式提示，并返回退出码 `4`
