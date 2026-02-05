package checks

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type loadAverageCheck struct{}

func (c *loadAverageCheck) Kind() string { return "load_average" }

func (c *loadAverageCheck) Run(ctx context.Context, req Request) (Result, error) {
	maxLoad1 := 4.0
	if v, ok := req.Params["max_load1"]; ok {
		maxLoad1 = toFloat(v, maxLoad1)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	load1, err := readLoad1(ctx, req)
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "load average check degraded: " + err.Error(), Advice: "verify host load manually"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	metrics := []schema.Metric{
		{Label: "load1", Value: fmt.Sprintf("%.2f", load1)},
		{Label: "max_load1", Value: fmt.Sprintf("%.2f", maxLoad1)},
	}
	if load1 <= maxLoad1 {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("load1 %.2f within threshold %.2f", load1, maxLoad1),
			Metrics:  metrics,
		}, nil
	}
	msg := fmt.Sprintf("load1 %.2f exceeds threshold %.2f", load1, maxLoad1)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "investigate hot processes or scale resources"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue, Metrics: metrics}, nil
}

func readLoad1(ctx context.Context, req Request) (float64, error) {
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(strings.TrimSpace(string(b)))
		if len(fields) > 0 {
			return strconv.ParseFloat(fields[0], 64)
		}
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "uptime", Timeout: 8 * time.Second})
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(out.Stdout)
	if idx := strings.Index(s, "load average"); idx >= 0 {
		rest := s[idx:]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			nums := strings.Split(strings.TrimSpace(rest[colon+1:]), ",")
			if len(nums) > 0 {
				return strconv.ParseFloat(strings.TrimSpace(nums[0]), 64)
			}
		}
	}
	return 0, fmt.Errorf("unable to parse load average")
}

func toFloat(v any, fallback float64) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return parsed
		}
	}
	return fallback
}
