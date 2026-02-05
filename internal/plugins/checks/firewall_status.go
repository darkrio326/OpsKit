package checks

import (
	"context"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type firewallStatusCheck struct{}

func (c *firewallStatusCheck) Kind() string { return "firewall_status" }

func (c *firewallStatusCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	state, detail := detectFirewallState(ctx, req)
	metrics := []schema.Metric{{Label: "firewall_status", Value: state}}
	switch state {
	case "enabled":
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "firewall enabled",
			Metrics:  metrics,
		}, nil
	case "disabled":
		issue := &schema.Issue{ID: req.ID, Severity: sev, Message: "firewall disabled", Advice: "enable host firewall policy before production traffic"}
		return Result{
			CheckID:  req.ID,
			Status:   statusFromSeverity(sev),
			Severity: sev,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  metrics,
		}, nil
	default:
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "firewall status unknown: " + detail, Advice: "verify firewall service manually"}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  metrics,
		}, nil
	}
}

func detectFirewallState(ctx context.Context, req Request) (string, string) {
	if out, ok := runNonEmpty(ctx, req, "firewall-cmd", "--state"); ok {
		if strings.Contains(strings.ToLower(out), "running") {
			return "enabled", "firewalld running"
		}
		return "disabled", "firewalld not running"
	}

	if out, ok := runNonEmpty(ctx, req, "ufw", "status"); ok {
		l := strings.ToLower(out)
		if strings.Contains(l, "status: active") {
			return "enabled", "ufw active"
		}
		if strings.Contains(l, "status: inactive") {
			return "disabled", "ufw inactive"
		}
	}

	if out, ok := runNonEmpty(ctx, req, "systemctl", "is-active", "firewalld"); ok {
		switch strings.TrimSpace(strings.ToLower(out)) {
		case "active":
			return "enabled", "firewalld service active"
		case "inactive", "failed":
			return "disabled", "firewalld service " + strings.TrimSpace(out)
		}
	}

	return "unknown", "no supported firewall probe succeeded"
}

func runNonEmpty(ctx context.Context, req Request, name string, args ...string) (string, bool) {
	out, err := req.Exec.Run(ctx, executil.Spec{Name: name, Args: args, Timeout: 8 * time.Second})
	if err != nil || out.ExitCode != 0 {
		return "", false
	}
	merged := strings.TrimSpace(out.Stdout)
	if merged == "" {
		merged = strings.TrimSpace(out.Stderr)
	}
	return merged, merged != ""
}
