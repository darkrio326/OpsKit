状态 JSON 与数据模型规范

1. 设计原则
	•	页面 只读 JSON
	•	JSON 是 OpsKit 的内部 API
	•	新字段只能追加，不破坏旧页面

⸻

2. 通用字段约定
	•	时间：ISO8601（例：2026-02-01T10:12:33+08:00）
	•	status 枚举：
	•	NOT_STARTED
	•	RUNNING
	•	PASSED
	•	WARN
	•	FAILED
	•	SKIPPED
	•	severity 枚举：
	•	info
	•	warn
	•	fail

⸻

3. overall.json
{
  "overallStatus": "HEALTHY",
  "lastRefreshTime": "2026-02-01T10:12:33+08:00",
  "activeTemplates": ["elasticsearch"],
  "openIssuesCount": 1
}

4. lifecycle.json
{
  "stages": [
    {
      "stageId": "A",
      "name": "Preflight",
      "status": "PASSED",
      "lastRunTime": "2026-02-01T09:50:00+08:00",
      "metrics": [
        { "label": "端口冲突", "value": "0" }
      ],
      "issues": [],
      "reportRef": "preflight-20260201.html"
    }
  ]
}

5. services.json
{
  "services": [
    {
      "serviceId": "elasticsearch",
      "unit": "elasticsearch",
      "health": "healthy",
      "checks": [
        {
          "checkId": "unit.active",
          "result": "PASS",
          "severity": "fail",
          "message": "systemd unit active"
        }
      ]
    }
  ]
}

6. artifacts.json
{
  "reports": [
    { "id": "preflight", "path": "reports/preflight-20260201.html" }
  ],
  "bundles": [
    { "id": "acceptance", "path": "bundles/acceptance-es-20260201.tar.gz" }
  ]
}