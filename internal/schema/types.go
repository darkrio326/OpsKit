package schema

type Status string

type Severity string

type OverallStatus string

const (
	StatusNotStarted Status = "NOT_STARTED"
	StatusRunning    Status = "RUNNING"
	StatusPassed     Status = "PASSED"
	StatusWarn       Status = "WARN"
	StatusFailed     Status = "FAILED"
	StatusSkipped    Status = "SKIPPED"
)

const (
	SeverityInfo Severity = "info"
	SeverityWarn Severity = "warn"
	SeverityFail Severity = "fail"
)

const (
	OverallHealthy   OverallStatus = "HEALTHY"
	OverallDegraded  OverallStatus = "DEGRADED"
	OverallUnhealthy OverallStatus = "UNHEALTHY"
	OverallUnknown   OverallStatus = "UNKNOWN"
)

type OverallState struct {
	OverallStatus   OverallStatus   `json:"overallStatus"`
	LastRefreshTime string          `json:"lastRefreshTime"`
	ActiveTemplates []string        `json:"activeTemplates"`
	OpenIssuesCount int             `json:"openIssuesCount"`
	RecoverSummary  *RecoverSummary `json:"recoverSummary,omitempty"`
}

type RecoverSummary struct {
	LastStatus     Status `json:"lastStatus,omitempty"`
	LastRunTime    string `json:"lastRunTime,omitempty"`
	LastTrigger    string `json:"lastTrigger,omitempty"`
	LastReasonCode string `json:"lastReasonCode,omitempty"`
	SuccessCount   int    `json:"successCount,omitempty"`
	FailureCount   int    `json:"failureCount,omitempty"`
	WarnCount      int    `json:"warnCount,omitempty"`
	CircuitOpen    bool   `json:"circuitOpen,omitempty"`
	CooldownUntil  string `json:"cooldownUntil,omitempty"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Issue struct {
	ID       string   `json:"id"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	Advice   string   `json:"advice,omitempty"`
}

type StageState struct {
	StageID     string   `json:"stageId"`
	Name        string   `json:"name"`
	Status      Status   `json:"status"`
	LastRunTime string   `json:"lastRunTime,omitempty"`
	Metrics     []Metric `json:"metrics,omitempty"`
	Issues      []Issue  `json:"issues,omitempty"`
	ReportRef   string   `json:"reportRef,omitempty"`
}

type LifecycleState struct {
	Stages []StageState `json:"stages"`
}

type CheckState struct {
	CheckID  string   `json:"checkId"`
	Result   string   `json:"result"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

type ServiceState struct {
	ServiceID string       `json:"serviceId"`
	Unit      string       `json:"unit"`
	Health    string       `json:"health"`
	Checks    []CheckState `json:"checks"`
}

type ServicesState struct {
	Services []ServiceState `json:"services"`
}

type ArtifactRef struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type ArtifactsState struct {
	Reports []ArtifactRef `json:"reports"`
	Bundles []ArtifactRef `json:"bundles"`
}
