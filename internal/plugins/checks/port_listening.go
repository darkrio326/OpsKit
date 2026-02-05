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

type portListeningCheck struct{}

func (c *portListeningCheck) Kind() string { return "port_listening" }

func (c *portListeningCheck) Run(ctx context.Context, req Request) (Result, error) {
	port := 0
	if v, ok := req.Params["port"]; ok {
		port = toInt(v, 0)
	}
	if port <= 0 {
		return Result{}, fmt.Errorf("port_listening requires params.port")
	}
	sev := schema.SeverityFail
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	out, err := req.Exec.Run(ctx, executil.Spec{Name: "ss", Args: []string{"-ltnH"}, Timeout: 8 * time.Second})
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "ss command unavailable: " + err.Error(), Advice: "install iproute2 for port checks"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}

	listening := false
	for _, line := range strings.Split(strings.TrimSpace(out.Stdout), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 4 {
			continue
		}
		local := fields[3]
		idx := strings.LastIndex(local, ":")
		if idx < 0 || idx+1 >= len(local) {
			continue
		}
		p, pErr := strconv.Atoi(local[idx+1:])
		if pErr == nil && p == port {
			listening = true
			break
		}
	}

	if listening {
		return Result{CheckID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("port %d is listening", port), Metrics: []schema.Metric{{Label: "port", Value: strconv.Itoa(port)}}}, nil
	}

	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: fmt.Sprintf("port %d is not listening", port), Advice: "start service or fix service port config"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  issue.Message,
		Issue:    issue,
		Metrics:  []schema.Metric{{Label: "port", Value: strconv.Itoa(port)}},
	}, nil
}
