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

type diskUsageCheck struct{}

func (c *diskUsageCheck) Kind() string { return "disk_usage" }

func (c *diskUsageCheck) Run(ctx context.Context, req Request) (Result, error) {
	mount := "/"
	if v, ok := req.Params["mount"]; ok {
		mount = toString(v, mount)
	}
	threshold := 90
	if v, ok := req.Params["max_used_percent"]; ok {
		threshold = toInt(v, threshold)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "df", Args: []string{"-P", mount}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "df command unavailable: " + err.Error(), Advice: "ensure coreutils available"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	lines := strings.Split(strings.TrimSpace(out.Stdout), "\n")
	if len(lines) < 2 {
		return Result{}, fmt.Errorf("unexpected df output for mount %s", mount)
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 5 {
		return Result{}, fmt.Errorf("invalid df row for mount %s", mount)
	}
	usedRaw := strings.TrimSuffix(fields[4], "%")
	used, convErr := strconv.Atoi(usedRaw)
	if convErr != nil {
		return Result{}, fmt.Errorf("parse disk usage: %w", convErr)
	}

	metrics := []schema.Metric{{Label: "mount", Value: mount}, {Label: "disk_used_percent", Value: strconv.Itoa(used)}}
	if used <= threshold {
		return Result{CheckID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("disk usage %s=%d%% (threshold=%d%%)", mount, used, threshold), Metrics: metrics}, nil
	}

	msg := fmt.Sprintf("disk usage high %s=%d%% (threshold=%d%%)", mount, used, threshold)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "free disk space or increase capacity"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue, Metrics: metrics}, nil
}
