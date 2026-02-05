package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemdUnitExistsCheck struct{}

func (c *systemdUnitExistsCheck) Kind() string { return "systemd_unit_exists" }

func (c *systemdUnitExistsCheck) Run(ctx context.Context, req Request) (Result, error) {
	unit := toString(req.Params["unit"], "")
	if unit == "" {
		return Result{}, fmt.Errorf("systemd_unit_exists requires params.unit")
	}
	sev := schema.SeverityFail
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "systemctl", Args: []string{"list-unit-files", "--full", "--no-legend", unit}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemctl unavailable: " + err.Error(), Advice: "run on systemd host"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	if strings.Contains(out.Stdout, unit) {
		return Result{CheckID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "systemd unit exists: " + unit}, nil
	}

	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: "systemd unit not found: " + unit, Advice: "install or correct unit name"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: issue.Message, Issue: issue}, nil
}
