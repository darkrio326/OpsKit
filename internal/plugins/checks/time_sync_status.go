package checks

import (
	"context"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type timeSyncStatusCheck struct{}

func (c *timeSyncStatusCheck) Kind() string { return "time_sync_status" }

func (c *timeSyncStatusCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	ok, source, detail := detectTimeSync(ctx, req)
	metrics := []schema.Metric{{Label: "time_sync", Value: ternary(ok, "synced", "unsynced")}, {Label: "time_sync_source", Value: source}}
	if ok {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "time synchronization healthy (" + source + ")",
			Metrics:  metrics,
		}, nil
	}
	if source == "unknown" {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "time sync status unknown: " + detail,
			Advice:   "verify ntp/chrony status manually",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  metrics,
		}, nil
	}
	msg := "time sync is not ready"
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "ensure ntp/chrony synchronization before production cutover"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}

func detectTimeSync(ctx context.Context, req Request) (bool, string, string) {
	// systemd host path
	out, err := req.Exec.Run(ctx, executil.Spec{Name: "timedatectl", Args: []string{"show", "-p", "NTPSynchronized", "--value"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		v := strings.ToLower(strings.TrimSpace(out.Stdout))
		if v == "yes" {
			return true, "timedatectl", "NTPSynchronized=yes"
		}
		if v == "no" {
			return false, "timedatectl", "NTPSynchronized=no"
		}
	}

	// chrony fallback
	out, err = req.Exec.Run(ctx, executil.Spec{Name: "chronyc", Args: []string{"tracking"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		s := strings.ToLower(out.Stdout)
		if strings.Contains(s, "leap status") && strings.Contains(s, "normal") {
			return true, "chronyc", "leap status normal"
		}
		if strings.Contains(s, "not synchronised") || strings.Contains(s, "not synchronized") {
			return false, "chronyc", "chrony not synchronised"
		}
		return false, "chronyc", "tracking output parsed as unsynced"
	}

	return false, "unknown", "timedatectl/chronyc unavailable"
}

func ternary(ok bool, yes, no string) string {
	if ok {
		return yes
	}
	return no
}
