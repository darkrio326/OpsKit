package checks

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type timeDriftCheck struct{}

func (c *timeDriftCheck) Kind() string { return "time_drift" }

func (c *timeDriftCheck) Run(ctx context.Context, req Request) (Result, error) {
	maxOffsetMS := 500.0
	if v, ok := req.Params["max_offset_ms"]; ok {
		maxOffsetMS = toFloat(v, maxOffsetMS)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	offsetMS, source, detail := detectTimeOffsetMS(ctx, req)
	metrics := []schema.Metric{
		{Label: "time_offset_ms", Value: fmt.Sprintf("%.3f", offsetMS)},
		{Label: "time_offset_source", Value: source},
	}

	if source == "unknown" {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "time drift unknown: " + detail,
			Advice:   "install chrony/ntp tooling or verify time sync manually",
		}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue, Metrics: metrics}, nil
	}

	if offsetMS <= maxOffsetMS {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("time drift %.3fms within threshold %.3fms", offsetMS, maxOffsetMS),
			Metrics:  metrics,
		}, nil
	}

	msg := fmt.Sprintf("time drift %.3fms exceeds threshold %.3fms", offsetMS, maxOffsetMS)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "check NTP sync and host clock source"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue, Metrics: metrics}, nil
}

func detectTimeOffsetMS(ctx context.Context, req Request) (float64, string, string) {
	out, err := req.Exec.Run(ctx, executil.Spec{Name: "chronyc", Args: []string{"tracking"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		offsetMS, ok := parseChronyLastOffsetMS(out.Stdout)
		if ok {
			return offsetMS, "chronyc", "last offset"
		}
		return 0, "chronyc", "unable to parse Last offset"
	}

	out, err = req.Exec.Run(ctx, executil.Spec{Name: "timedatectl", Args: []string{"timesync-status"}, Timeout: 8 * time.Second})
	if err == nil && out.ExitCode == 0 {
		offsetMS, ok := parseTimedatectlOffsetMS(out.Stdout)
		if ok {
			return offsetMS, "timedatectl", "offset"
		}
	}
	return 0, "unknown", "chronyc/timedatectl offset unavailable"
}

func parseChronyLastOffsetMS(out string) (float64, bool) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(strings.ToLower(line), "last offset") {
			continue
		}
		if idx := strings.Index(line, ":"); idx >= 0 {
			right := strings.TrimSpace(line[idx+1:])
			fields := strings.Fields(right)
			if len(fields) == 0 {
				return 0, false
			}
			n, err := strconv.ParseFloat(fields[0], 64)
			if err != nil {
				return 0, false
			}
			unit := "seconds"
			if len(fields) > 1 {
				unit = strings.ToLower(fields[1])
			}
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
	}
	return 0, false
}

func parseTimedatectlOffsetMS(out string) (float64, bool) {
	for _, line := range strings.Split(out, "\n") {
		lower := strings.ToLower(strings.TrimSpace(line))
		if !strings.Contains(lower, "offset") {
			continue
		}
		if idx := strings.Index(line, ":"); idx >= 0 {
			right := strings.TrimSpace(line[idx+1:])
			fields := strings.Fields(right)
			if len(fields) == 0 {
				continue
			}
			num := strings.Trim(fields[0], "±")
			n, err := strconv.ParseFloat(num, 64)
			if err != nil {
				continue
			}
			unit := "ms"
			if len(fields) > 1 {
				unit = strings.ToLower(fields[1])
			}
			switch unit {
			case "s", "sec", "second", "seconds":
				return math.Abs(n) * 1000, true
			case "ms", "millisecond", "milliseconds":
				return math.Abs(n), true
			case "us", "microsecond", "microseconds":
				return math.Abs(n) / 1000, true
			default:
				return math.Abs(n), true
			}
		}
	}
	return 0, false
}
