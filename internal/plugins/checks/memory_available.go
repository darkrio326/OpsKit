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

type memoryAvailableCheck struct{}

func (c *memoryAvailableCheck) Kind() string { return "memory_available" }

func (c *memoryAvailableCheck) Run(ctx context.Context, req Request) (Result, error) {
	minMB := 512
	if v, ok := req.Params["min_available_mb"]; ok {
		minMB = toInt(v, minMB)
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "free", Args: []string{"-m"}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "free command unavailable: " + err.Error(), Advice: "ensure procps available"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	available := -1
	for _, line := range strings.Split(strings.TrimSpace(out.Stdout), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !strings.HasPrefix(fields[0], "Mem:") {
			continue
		}
		if len(fields) >= 7 {
			v, convErr := strconv.Atoi(fields[6])
			if convErr == nil {
				available = v
			}
		}
		break
	}
	if available < 0 {
		return Result{}, fmt.Errorf("failed to parse memory availability")
	}

	metrics := []schema.Metric{{Label: "memory_available_mb", Value: strconv.Itoa(available)}}
	if available >= minMB {
		return Result{CheckID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("memory available %dMB (threshold=%dMB)", available, minMB), Metrics: metrics}, nil
	}

	msg := fmt.Sprintf("memory available low %dMB (threshold=%dMB)", available, minMB)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "release memory or scale host resources"}
	return Result{CheckID: req.ID, Status: statusFromSeverity(sev), Severity: sev, Message: msg, Issue: issue, Metrics: metrics}, nil
}
