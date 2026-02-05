package actions

import (
	"context"
	"fmt"

	"opskit/internal/schema"
)

type systemdStopAction struct{}

func (a *systemdStopAction) Kind() string { return "systemd_stop" }

func (a *systemdStopAction) Run(ctx context.Context, req Request) (Result, error) {
	unit := toString(req.Params["unit"], "")
	if unit == "" {
		return Result{}, fmt.Errorf("systemd_stop requires params.unit")
	}
	res, err := runCmd(ctx, req, "systemctl", "stop", unit)
	if err != nil {
		return Result{}, err
	}
	if res.ExitCode != 0 {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "systemctl stop failed: " + unit, Advice: res.Stderr}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "stopped unit: " + unit}, nil
}
