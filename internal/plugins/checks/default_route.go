package checks

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type defaultRouteCheck struct{}

func (c *defaultRouteCheck) Kind() string { return "default_route" }

func (c *defaultRouteCheck) Run(ctx context.Context, req Request) (Result, error) {
	sev := schema.SeverityFail
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	ok, source, err := hasDefaultRoute(ctx, req)
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "default route check degraded: " + err.Error(), Advice: "install iproute tools or verify route manually"}
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  issue.Message,
			Issue:    issue,
		}, nil
	}
	if ok {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "default route present (" + source + ")",
			Metrics:  []schema.Metric{{Label: "default_route", Value: "present"}},
		}, nil
	}

	msg := "default route missing"
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "configure default gateway before production traffic"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  []schema.Metric{{Label: "default_route", Value: "missing"}},
	}, nil
}

func hasDefaultRoute(ctx context.Context, req Request) (bool, string, error) {
	f, err := os.Open("/proc/net/route")
	if err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" || strings.HasPrefix(line, "Iface") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == "00000000" {
				return true, "/proc/net/route", nil
			}
		}
		if scanErr := s.Err(); scanErr != nil {
			return false, "", scanErr
		}
		return false, "/proc/net/route", nil
	}

	out, runErr := req.Exec.Run(ctx, executil.Spec{Name: "ip", Args: []string{"route", "show", "default"}, Timeout: 8 * time.Second})
	if runErr != nil {
		return false, "", runErr
	}
	return strings.TrimSpace(out.Stdout) != "", "ip route", nil
}
