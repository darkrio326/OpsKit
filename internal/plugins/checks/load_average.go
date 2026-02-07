package checks

import (
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"regexp"
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

	load1, source, degradeReason, err := readLoad1(ctx, req)
	baseMetrics := []schema.Metric{
		{Label: "max_load1", Value: fmt.Sprintf("%.2f", maxLoad1)},
		{Label: "load_source", Value: source},
	}
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "load average check degraded: " + err.Error(), Advice: "verify host load manually"}
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
		{Label: "load1", Value: fmt.Sprintf("%.2f", load1)},
		{Label: "max_load1", Value: fmt.Sprintf("%.2f", maxLoad1)},
		{Label: "load_source", Value: source},
	})
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

func readLoad1(ctx context.Context, req Request) (float64, string, string, error) {
	loadavgPath := toString(req.Params["loadavg_file"], "/proc/loadavg")
	parseFailed := false
	commandFailed := false
	toolUnavailable := false
	attemptDetails := []string{}

	if b, err := os.ReadFile(loadavgPath); err == nil {
		v, parseErr := parseFirstFloat(string(b))
		if parseErr == nil {
			return v, "proc_loadavg", "", nil
		}
		parseFailed = true
		attemptDetails = append(attemptDetails, "proc_loadavg parse failed")
	} else {
		attemptDetails = append(attemptDetails, "proc_loadavg unavailable")
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "uptime", Timeout: 8 * time.Second})
	if err == nil {
		v, parseErr := parseLoadFromUptime(out.Stdout)
		if parseErr == nil {
			return v, "uptime", "", nil
		}
		parseFailed = true
		attemptDetails = append(attemptDetails, "uptime parse failed")
	} else {
		if errors.Is(err, osexec.ErrNotFound) {
			toolUnavailable = true
		} else {
			commandFailed = true
		}
		attemptDetails = append(attemptDetails, "uptime unavailable")
	}

	out, err = req.Exec.Run(ctx, executil.Spec{Name: "sysctl", Args: []string{"-n", "vm.loadavg"}, Timeout: 8 * time.Second})
	if err == nil {
		v, parseErr := parseFirstFloat(out.Stdout)
		if parseErr == nil {
			return v, "sysctl_vm_loadavg", "", nil
		}
		parseFailed = true
		attemptDetails = append(attemptDetails, "sysctl parse failed")
	} else {
		if errors.Is(err, osexec.ErrNotFound) {
			toolUnavailable = true
		} else {
			commandFailed = true
		}
		attemptDetails = append(attemptDetails, "sysctl unavailable")
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

var numericPattern = regexp.MustCompile(`[+-]?\d+(?:[.,]\d+)?`)

func parseLoadFromUptime(out string) (float64, error) {
	text := strings.TrimSpace(out)
	if text == "" {
		return 0, fmt.Errorf("empty output")
	}
	lower := strings.ToLower(text)
	if idx := strings.Index(lower, "load average"); idx >= 0 {
		text = text[idx:]
	} else if idx := strings.Index(lower, "load averages"); idx >= 0 {
		text = text[idx:]
	}
	return parseFirstFloat(text)
}

func parseFirstFloat(text string) (float64, error) {
	nums := numericPattern.FindAllString(text, -1)
	if len(nums) == 0 {
		return 0, fmt.Errorf("no numeric values")
	}
	raw := strings.ReplaceAll(nums[0], ",", ".")
	return strconv.ParseFloat(raw, 64)
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
