安全模型与操作边界

1. 权限模型
	•	默认以 root 运行
	•	页面默认只读
	•	所有写操作可审计

⸻

2. 操作分级
	•	Level 0（只读）：ss / df / systemctl status
	•	Level 1（安全写）：创建目录、写 OpsKit 文件
	•	Level 2（中风险）：install unit、enable/start 服务
	•	Level 3（高风险）：删除数据、改系统参数（v1 禁止）

⸻

3. 脱敏规则
	•	password / token / secret → ******
	•	配置文件只做 hash 或脱敏输出