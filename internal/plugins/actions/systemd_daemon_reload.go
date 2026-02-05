package actions

import (
	"context"

	"opskit/internal/schema"
)

type systemdDaemonReloadAction struct{}

func (a *systemdDaemonReloadAction) Kind() string { return "systemd_daemon_reload" }

func (a *systemdDaemonReloadAction) Run(ctx context.Context, req Request) (Result, error) {
	res, err := runCmd(ctx, req, "systemctl", "daemon-reload")
	if err != nil {
		return Result{}, err
	}
	if res.ExitCode != 0 {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "systemctl daemon-reload failed", Advice: res.Stderr}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "systemd daemon-reload done"}, nil
}
