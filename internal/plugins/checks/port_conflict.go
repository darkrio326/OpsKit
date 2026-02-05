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

type portConflictCheck struct{}

func (c *portConflictCheck) Kind() string { return "port_conflict" }

func (c *portConflictCheck) Run(ctx context.Context, req Request) (Result, error) {
	reserved := []int{80, 443, 9200, 5601}
	if v, ok := req.Params["reserved_ports"]; ok {
		reserved = toIntSlice(v, reserved)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "ss", Args: []string{"-ltnH"}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "port scan command unavailable: " + err.Error(),
			Advice:   "install iproute2/ss or provide alternative scanner in later milestone",
		}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
		}, nil
	}

	listening := map[int]bool{}
	for _, line := range strings.Split(strings.TrimSpace(out.Stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		local := fields[3]
		idx := strings.LastIndex(local, ":")
		if idx < 0 || idx+1 >= len(local) {
			continue
		}
		p, err := strconv.Atoi(local[idx+1:])
		if err == nil {
			listening[p] = true
		}
	}

	conflicts := []int{}
	for _, p := range reserved {
		if listening[p] {
			conflicts = append(conflicts, p)
		}
	}

	if len(conflicts) > 0 {
		vals := make([]string, 0, len(conflicts))
		for _, c := range conflicts {
			vals = append(vals, strconv.Itoa(c))
		}
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "reserved ports already in use: " + strings.Join(vals, ","), Advice: "stop conflicting service or update template reserved_ports"}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusFailed,
			Severity: schema.SeverityFail,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  []schema.Metric{{Label: "port_conflicts", Value: fmt.Sprintf("%d", len(conflicts))}},
		}, nil
	}

	return Result{
		CheckID:  req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  "no reserved port conflicts",
		Metrics:  []schema.Metric{{Label: "reserved_ports", Value: fmt.Sprintf("%d", len(reserved))}},
	}, nil
}
