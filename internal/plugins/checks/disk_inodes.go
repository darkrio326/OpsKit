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

type diskInodesCheck struct{}

func (c *diskInodesCheck) Kind() string { return "disk_inodes" }

func (c *diskInodesCheck) Run(ctx context.Context, req Request) (Result, error) {
	mount := "/"
	if v, ok := req.Params["mount"]; ok {
		mount = toString(v, mount)
	}
	threshold := 90
	if v, ok := req.Params["max_iused_percent"]; ok {
		threshold = toInt(v, threshold)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "df", Args: []string{"-Pi", mount}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "inode check degraded: " + err.Error(),
			Advice:   "verify inode usage with df -i manually",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics: withDegradedMetrics([]schema.Metric{
				{Label: "mount", Value: mount},
			}, "df_unavailable"),
		}, nil
	}

	used, parseErr := parseInodeUsedPercent(out.Stdout)
	if parseErr != nil {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "inode check degraded: " + parseErr.Error(),
			Advice:   "verify inode usage output format and locale",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics: withDegradedMetrics([]schema.Metric{
				{Label: "mount", Value: mount},
			}, "parse_error"),
		}, nil
	}

	metrics := withHealthyMetrics([]schema.Metric{
		{Label: "mount", Value: mount},
		{Label: "inode_used_percent", Value: strconv.Itoa(used)},
		{Label: "max_iused_percent", Value: strconv.Itoa(threshold)},
	})

	if used <= threshold {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("inode usage %s=%d%% within threshold %d%%", mount, used, threshold),
			Metrics:  metrics,
		}, nil
	}

	msg := fmt.Sprintf("inode usage high %s=%d%% (threshold=%d%%)", mount, used, threshold)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "clean stale files or increase inode capacity"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}

func parseInodeUsedPercent(stdout string) (int, error) {
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df -i output")
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 5 {
		return 0, fmt.Errorf("invalid df -i row")
	}
	usedRaw := strings.TrimSuffix(fields[4], "%")
	used, err := strconv.Atoi(usedRaw)
	if err != nil {
		return 0, fmt.Errorf("parse inode usage: %w", err)
	}
	return used, nil
}
