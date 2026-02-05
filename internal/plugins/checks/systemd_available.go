package checks

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemdAvailableCheck struct{}

func (c *systemdAvailableCheck) Kind() string { return "systemd_available" }

func (c *systemdAvailableCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}
	out, err := req.Exec.Run(ctx, executil.Spec{Name: "systemctl", Args: []string{"--version"}, Timeout: 8 * time.Second})
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			msg := "systemctl not found"
			issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "install systemd tooling or adjust host baseline"}
			return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue}, nil
		}
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemctl check degraded: " + err.Error(), Advice: "verify systemd status manually"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}
	if out.ExitCode != 0 {
		msg := strings.TrimSpace(out.Stderr)
		if msg == "" {
			msg = "systemctl returned non-zero"
		}
		issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "check init system and systemd service state"}
		return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue}, nil
	}
	return Result{
		CheckID:  req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  "systemd available",
		Metrics:  []schema.Metric{{Label: "systemd", Value: "available"}},
	}, nil
}
