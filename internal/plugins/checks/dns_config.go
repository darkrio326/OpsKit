package checks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"opskit/internal/schema"
)

type dnsConfigCheck struct{}

func (c *dnsConfigCheck) Kind() string { return "dns_config" }

func (c *dnsConfigCheck) Run(_ context.Context, req Request) (Result, error) {
	sev := schema.SeverityFail
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}
	path := toString(req.Params["resolv_conf"], "/etc/resolv.conf")
	f, err := os.Open(path)
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "read resolv.conf failed: " + err.Error(), Advice: "verify DNS configuration manually"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}
	defer f.Close()

	nameservers := 0
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "nameserver ") {
			nameservers++
		}
	}
	if err := s.Err(); err != nil {
		return Result{}, err
	}

	metrics := []schema.Metric{{Label: "dns_nameservers", Value: fmt.Sprintf("%d", nameservers)}}
	if nameservers > 0 {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "dns nameserver entries present",
			Metrics:  metrics,
		}, nil
	}
	msg := "no dns nameserver entry found"
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "configure at least one resolver in resolv.conf"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}
