package checks

import (
	"context"
	"errors"
	"fmt"
	"math"
	osexec "os/exec"
	"strconv"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type clockSkewCheck struct{}

func (c *clockSkewCheck) Kind() string { return "clock_skew" }

func (c *clockSkewCheck) Run(ctx context.Context, req Request) (Result, error) {
	maxSkewMS := 2000.0
	if v, ok := req.Params["max_skew_ms"]; ok {
		maxSkewMS = toFloat(v, maxSkewMS)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	skewMS, source, degradeReason, err := detectClockSkewMS(ctx, req)
	baseMetrics := []schema.Metric{
		{Label: "clock_skew_source", Value: source},
		{Label: "max_skew_ms", Value: fmt.Sprintf("%.3f", maxSkewMS)},
	}
	if err != nil {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "clock skew check degraded: " + err.Error(),
			Advice:   "verify clock sync status manually",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  withDegradedMetrics(baseMetrics, degradeReason),
		}, nil
	}

	metrics := withHealthyMetrics([]schema.Metric{
		{Label: "clock_skew_ms", Value: fmt.Sprintf("%.3f", skewMS)},
		{Label: "clock_skew_source", Value: source},
		{Label: "max_skew_ms", Value: fmt.Sprintf("%.3f", maxSkewMS)},
	})
	if skewMS <= maxSkewMS {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("clock skew %.3fms within threshold %.3fms", skewMS, maxSkewMS),
			Metrics:  metrics,
		}, nil
	}

	msg := fmt.Sprintf("clock skew %.3fms exceeds threshold %.3fms", skewMS, maxSkewMS)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "check time sync source and host clock stability"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}

func detectClockSkewMS(ctx context.Context, req Request) (float64, string, string, error) {
	parseFailed := false
	commandFailed := false
	toolUnavailable := false
	attemptDetails := []string{}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "chronyc", Args: []string{"tracking"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		if v, ok := parseChronySystemTimeMS(out.Stdout); ok {
			return v, "chronyc_system_time", "", nil
		}
		if v, ok := parseChronyLastOffsetMS(out.Stdout); ok {
			return v, "chronyc_last_offset", "", nil
		}
		parseFailed = true
		attemptDetails = append(attemptDetails, "chronyc parse failed")
	} else if err != nil {
		if errors.Is(err, osexec.ErrNotFound) {
			toolUnavailable = true
		} else {
			commandFailed = true
		}
		attemptDetails = append(attemptDetails, "chronyc unavailable")
	} else {
		commandFailed = true
		attemptDetails = append(attemptDetails, "chronyc exit non-zero")
	}

	out, err = req.Exec.Run(ctx, executil.Spec{Name: "timedatectl", Args: []string{"timesync-status"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		if v, ok := parseTimedatectlOffsetMS(out.Stdout); ok {
			return v, "timedatectl_offset", "", nil
		}
		parseFailed = true
		attemptDetails = append(attemptDetails, "timedatectl parse failed")
	} else if err != nil {
		if errors.Is(err, osexec.ErrNotFound) {
			toolUnavailable = true
		} else {
			commandFailed = true
		}
		attemptDetails = append(attemptDetails, "timedatectl unavailable")
	} else {
		commandFailed = true
		attemptDetails = append(attemptDetails, "timedatectl exit non-zero")
	}

	reason := "unknown"
	switch {
	case parseFailed:
		reason = "parse_error"
	case commandFailed:
		reason = "command_failed"
	case toolUnavailable:
		reason = "tool_unavailable"
	default:
		reason = "unsupported_platform"
	}
	return 0, "unknown", reason, fmt.Errorf(strings.Join(attemptDetails, "; "))
}

func parseChronySystemTimeMS(out string) (float64, bool) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(strings.ToLower(line), "system time") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		right := strings.TrimSpace(line[idx+1:])
		fields := strings.Fields(right)
		if len(fields) < 2 {
			continue
		}
		num := strings.Trim(fields[0], "±")
		n, err := strconv.ParseFloat(num, 64)
		if err != nil {
			continue
		}
		unit := strings.ToLower(fields[1])
		switch unit {
		case "second", "seconds", "sec", "s":
			return math.Abs(n) * 1000, true
		case "millisecond", "milliseconds", "ms":
			return math.Abs(n), true
		case "microsecond", "microseconds", "us":
			return math.Abs(n) / 1000, true
		default:
			return math.Abs(n) * 1000, true
		}
	}
	return 0, false
}
