package checks

import (
	"context"
	"os"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type selinuxStatusCheck struct{}

func (c *selinuxStatusCheck) Kind() string { return "selinux_status" }

func (c *selinuxStatusCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	state, detail := detectSELinuxState(ctx, req)
	metrics := []schema.Metric{{Label: "selinux_status", Value: state}}
	switch state {
	case "enforcing":
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "SELinux enforcing",
			Metrics:  metrics,
		}, nil
	case "permissive":
		issue := &schema.Issue{ID: req.ID, Severity: sev, Message: "SELinux permissive", Advice: "confirm this is expected for production baseline"}
		return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: issue.Message, Issue: issue, Metrics: metrics}, nil
	case "disabled":
		issue := &schema.Issue{ID: req.ID, Severity: sev, Message: "SELinux disabled", Advice: "confirm security baseline requirements"}
		return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: issue.Message, Issue: issue, Metrics: metrics}, nil
	default:
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "SELinux status unknown: " + detail, Advice: "verify SELinux state manually"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue, Metrics: metrics}, nil
	}
}

func detectSELinuxState(ctx context.Context, req Request) (string, string) {
	if b, err := os.ReadFile("/sys/fs/selinux/enforce"); err == nil {
		v := strings.TrimSpace(string(b))
		switch v {
		case "1":
			return "enforcing", "/sys/fs/selinux/enforce"
		case "0":
			return "permissive", "/sys/fs/selinux/enforce"
		}
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "getenforce", Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		switch strings.ToLower(strings.TrimSpace(out.Stdout)) {
		case "enforcing":
			return "enforcing", "getenforce"
		case "permissive":
			return "permissive", "getenforce"
		case "disabled":
			return "disabled", "getenforce"
		}
	}
	return "unknown", "getenforce unavailable"
}
