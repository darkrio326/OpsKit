最小验收测试用例

A Preflight
	•	启动占用端口 → A FAILED
	•	修复端口 → A PASSED

C Deploy
	•	缺少离线包 → exit 2
	•	启动失败 → 生成日志与建议

D Operate
	•	stop 服务 → 页面变红
	•	start 服务 → 页面回绿

E Recover
	•	reboot 服务器 → 自动恢复一次
	•	连续失败 → 不反复重启（熔断生效）

F Accept
	•	生成 tar.gz
	•	校验 hashes.txt 可复核

⸻

✅ 到这里，你已经具备：
	•	冻结边界
	•	稳定数据模型
	•	可编码的 CLI 设计
	•	可审计的安全模型
	•	可测试的验收路径