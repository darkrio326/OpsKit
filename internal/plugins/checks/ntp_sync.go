package checks

import (
	"context"

	"opskit/internal/schema"
)

type ntpSyncCheck struct{}

func (c *ntpSyncCheck) Kind() string { return "ntp_sync" }

func (c *ntpSyncCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	ok, source, detail := detectTimeSync(ctx, req)
	metrics := []schema.Metric{
		{Label: "ntp_sync", Value: ternary(ok, "synced", "unsynced")},
		{Label: "ntp_source", Value: source},
	}
	if ok {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "ntp synchronization healthy (" + source + ")",
			Metrics:  withHealthyMetrics(metrics),
		}, nil
	}
	if source == "unknown" {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "ntp sync status unknown: " + detail,
			Advice:   "verify timedatectl/chronyc on target host",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  withDegradedMetrics(metrics, "time_sync_probe_unavailable"),
		}, nil
	}

	msg := "ntp synchronization is not ready"
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "ensure ntp/chrony sync before deployment or handover"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  withHealthyMetrics(metrics),
	}, nil
}
