package checks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemdBasicsCheck struct{}

func (c *systemdBasicsCheck) Kind() string { return "systemd_basics" }

func (c *systemdBasicsCheck) Run(ctx context.Context, req Request) (Result, error) {
	candidates := []string{
		"sshd.service",
		"ssh.service",
		"systemd-timesyncd.service",
		"chronyd.service",
		"network-online.target",
	}
	if v, ok := req.Params["candidate_units"]; ok {
		candidates = toStringSlice(v, candidates)
	}
	requireAny := toBool(req.Params["require_any"], false)
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	discovered := 0
	active := 0
	inactiveUnits := []string{}
	for _, unit := range candidates {
		exists, isActive, err := probeUnit(ctx, req, unit)
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				issue := &schema.Issue{
					ID:       req.ID,
					Severity: schema.SeverityWarn,
					Message:  "systemctl not available for systemd basics check",
					Advice:   "verify base services manually",
				}
				return Result{
					CheckID:  req.ID,
					Status:   schema.StatusWarn,
					Severity: schema.SeverityWarn,
					Message:  issue.Message,
					Issue:    issue,
				}, nil
			}
			return Result{}, err
		}
		if !exists {
			continue
		}
		discovered++
		if isActive {
			active++
		} else {
			inactiveUnits = append(inactiveUnits, unit)
		}
	}

	metrics := []schema.Metric{
		{Label: "systemd_units_discovered", Value: fmt.Sprintf("%d", discovered)},
		{Label: "systemd_units_active", Value: fmt.Sprintf("%d", active)},
	}
	if discovered == 0 {
		if requireAny {
			msg := "no baseline systemd units discovered"
			issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "ensure host exposes expected systemd units"}
			return Result{
				CheckID:  req.ID,
				Status:   statusFromSeverity(sev),
				Severity: sev,
				Message:  msg,
				Issue:    issue,
				Metrics:  metrics,
			}, nil
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "no baseline units discovered (skipped)",
			Metrics:  metrics,
		}, nil
	}
	if len(inactiveUnits) == 0 {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "baseline systemd units active",
			Metrics:  metrics,
		}, nil
	}

	msg := "inactive baseline units: " + strings.Join(inactiveUnits, ",")
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "investigate and recover inactive systemd baseline units"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}

func probeUnit(ctx context.Context, req Request, unit string) (bool, bool, error) {
	out, err := req.Exec.Run(ctx, executil.Spec{
		Name:    "systemctl",
		Args:    []string{"show", unit, "--property=LoadState", "--property=ActiveState", "--value"},
		Timeout: 8 * time.Second,
	})
	if err != nil {
		return false, false, err
	}
	lines := strings.Split(strings.TrimSpace(out.Stdout), "\n")
	if len(lines) == 0 {
		return false, false, nil
	}
	loadState := strings.TrimSpace(lines[0])
	activeState := ""
	if len(lines) > 1 {
		activeState = strings.TrimSpace(lines[1])
	}
	if loadState == "" || strings.EqualFold(loadState, "not-found") {
		return false, false, nil
	}
	return true, strings.EqualFold(activeState, "active"), nil
}
