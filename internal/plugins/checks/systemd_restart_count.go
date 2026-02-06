package checks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemdRestartCountCheck struct{}

func (c *systemdRestartCountCheck) Kind() string { return "systemd_restart_count" }

func (c *systemdRestartCountCheck) Run(ctx context.Context, req Request) (Result, error) {
	unit := toString(req.Params["unit"], "")
	if unit == "" {
		return Result{}, fmt.Errorf("systemd_restart_count requires params.unit")
	}
	maxRestarts := toInt(req.Params["max_restarts"], 3)
	if maxRestarts < 0 {
		maxRestarts = 0
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "systemctl", Args: []string{"show", unit, "--property", "NRestarts", "--value"}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemd restart count degraded: " + err.Error(), Advice: "run on systemd host or install systemctl"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}
	if out.ExitCode != 0 {
		detail := strings.TrimSpace(out.Stderr)
		if detail == "" {
			detail = "systemctl show returned non-zero"
		}
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemd restart count degraded: " + detail, Advice: "verify unit name and systemd status"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	nRestartsText := strings.TrimSpace(out.Stdout)
	nRestarts, parseErr := strconv.Atoi(nRestartsText)
	if parseErr != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemd restart count parse failed: " + parseErr.Error(), Advice: "inspect systemctl output format"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	metrics := []schema.Metric{
		{Label: "restart_unit", Value: unit},
		{Label: "restart_count", Value: strconv.Itoa(nRestarts)},
		{Label: "restart_max", Value: strconv.Itoa(maxRestarts)},
	}
	if nRestarts <= maxRestarts {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("systemd restart count within threshold: %s (%d/%d)", unit, nRestarts, maxRestarts),
			Metrics:  metrics,
		}, nil
	}

	msg := fmt.Sprintf("systemd restart count too high: %s (%d/%d)", unit, nRestarts, maxRestarts)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "inspect recent crashes and stabilize service before handover"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}
