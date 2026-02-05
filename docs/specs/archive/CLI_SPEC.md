命令行接口规范

1. 命令总览
	•	opskit install
	•	opskit run <A|B|C|D|E|F>
	•	opskit status
	•	opskit accept
	•	opskit handover
	•	opskit template validate <file>

⸻

2. 全局参数
	•	--template <id|path>
	•	--vars key=value[,key=value]
	•	--dry-run
	•	--fix
	•	--force
	•	--output <dir>


3. 退出码规范


code
含义
0
成功
1
失败
2
前置条件不满足
3
部分成功（WARN）
4
需要人工确认

4. 输出约定
	•	stdout：人类可读摘要
	•	JSON：写入 /var/lib/opskit/state
	•	日志：/var/log/opskit/opskit.log