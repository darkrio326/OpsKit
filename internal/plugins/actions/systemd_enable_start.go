package actions

import (
	"context"
	"fmt"

	"opskit/internal/schema"
)

type systemdEnableStartAction struct{}

func (a *systemdEnableStartAction) Kind() string { return "systemd_enable_start" }

func (a *systemdEnableStartAction) Run(ctx context.Context, req Request) (Result, error) {
	unit := toString(req.Params["unit"], "")
	if unit == "" {
		return Result{}, fmt.Errorf("systemd_enable_start requires params.unit")
	}
	res, err := runCmd(ctx, req, "systemctl", "enable", "--now", unit)
	if err != nil {
		return Result{}, err
	}
	if res.ExitCode != 0 {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "systemctl enable --now failed: " + unit, Advice: res.Stderr}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "enabled and started unit: " + unit}, nil
}
