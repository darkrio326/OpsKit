package checks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type mountCheck struct{}

func (c *mountCheck) Kind() string { return "mount_check" }

func (c *mountCheck) Run(ctx context.Context, req Request) (Result, error) {
	required := []string{"/", "/data", "/opt", "/logs"}
	if v, ok := req.Params["required_mounts"]; ok {
		required = toStringSlice(v, required)
	}

	mounted := map[string]bool{}
	f, err := os.Open("/proc/mounts")
	if err == nil {
		defer f.Close()

		s := bufio.NewScanner(f)
		for s.Scan() {
			parts := strings.Fields(s.Text())
			if len(parts) >= 2 {
				mounted[parts[1]] = true
			}
		}
		if err := s.Err(); err != nil {
			return Result{}, err
		}
	} else {
		out, runErr := req.Exec.Run(ctx, executil.Spec{Name: "mount", Timeout: 8 * time.Second})
		if runErr != nil {
			return Result{}, runErr
		}
		for _, line := range strings.Split(out.Stdout, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// mount output is generally "<src> on <target> type <fstype> ..."
			parts := strings.Split(line, " on ")
			if len(parts) < 2 {
				continue
			}
			right := parts[1]
			target := strings.SplitN(right, " type ", 2)[0]
			if target != "" {
				mounted[target] = true
			}
		}
	}

	missing := []string{}
	for _, m := range required {
		if !mounted[m] {
			missing = append(missing, m)
		}
	}

	if len(missing) > 0 {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "missing required mounts: " + strings.Join(missing, ","), Advice: "mount required filesystems before deploy"}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusFailed,
			Severity: schema.SeverityFail,
			Message:  issue.Message,
			Issue:    issue,
			Metrics:  []schema.Metric{{Label: "missing_mounts", Value: fmt.Sprintf("%d", len(missing))}},
		}, nil
	}

	return Result{
		CheckID:  req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  "required mounts are present",
		Metrics:  []schema.Metric{{Label: "required_mounts", Value: fmt.Sprintf("%d", len(required))}},
	}, nil
}
