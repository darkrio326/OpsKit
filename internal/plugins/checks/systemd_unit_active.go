package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemdUnitActiveCheck struct{}

func (c *systemdUnitActiveCheck) Kind() string { return "systemd_unit_active" }

func (c *systemdUnitActiveCheck) Run(ctx context.Context, req Request) (Result, error) {
	unit := toString(req.Params["unit"], "")
	if unit == "" {
		return Result{}, fmt.Errorf("systemd_unit_active requires params.unit")
	}
	sev := schema.SeverityFail
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "systemctl", Args: []string{"is-active", unit}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "systemctl unavailable: " + err.Error(), Advice: "run on systemd host"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	state := strings.TrimSpace(out.Stdout)
	if state == "active" {
		return Result{CheckID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "systemd unit active: " + unit}, nil
	}

	msg := fmt.Sprintf("systemd unit not active: %s (state=%s)", unit, state)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "start service and inspect unit logs"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue}, nil
}
