package checks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type dnsResolveCheck struct{}

func (c *dnsResolveCheck) Kind() string { return "dns_resolve" }

func (c *dnsResolveCheck) Run(ctx context.Context, req Request) (Result, error) {
	host := toString(req.Params["host"], "localhost")
	minRecords := toInt(req.Params["min_records"], 1)
	if minRecords < 1 {
		minRecords = 1
	}
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	source, records, detail, err := resolveDNS(ctx, req, host)
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: "dns resolve degraded: " + err.Error(), Advice: "install getent/nslookup or verify resolver manually"}
		return Result{CheckID: req.ID, Status: schema.StatusWarn, Severity: schema.SeverityWarn, Message: issue.Message, Issue: issue}, nil
	}
	if records >= minRecords {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  fmt.Sprintf("dns resolve ok: host=%s records=%d source=%s", host, records, source),
			Metrics: []schema.Metric{
				{Label: "dns_host", Value: host},
				{Label: "dns_records", Value: fmt.Sprintf("%d", records)},
				{Label: "dns_source", Value: source},
			},
		}, nil
	}

	msg := fmt.Sprintf("dns resolve failed: host=%s (%s)", host, detail)
	issue := &schema.Issue{ID: req.ID, Severity: sev, Message: msg, Advice: "check DNS server/reachability and /etc/resolv.conf"}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics: []schema.Metric{
			{Label: "dns_host", Value: host},
			{Label: "dns_records", Value: fmt.Sprintf("%d", records)},
			{Label: "dns_source", Value: source},
		},
	}, nil
}

func resolveDNS(ctx context.Context, req Request, host string) (string, int, string, error) {
	specs := []struct {
		source string
		spec   executil.Spec
	}{
		{source: "getent", spec: executil.Spec{Name: "getent", Args: []string{"hosts", host}, Timeout: 8 * time.Second}},
		{source: "nslookup", spec: executil.Spec{Name: "nslookup", Args: []string{host}, Timeout: 8 * time.Second}},
	}

	anyExecuted := false
	lastSource := "unknown"
	details := make([]string, 0, len(specs))
	runErrs := make([]error, 0, len(specs))
	for _, item := range specs {
		res, err := req.Exec.Run(ctx, item.spec)
		if err != nil {
			runErrs = append(runErrs, fmt.Errorf("%s unavailable: %w", item.source, err))
			continue
		}
		anyExecuted = true
		lastSource = item.source
		count := countDNSRecords(item.source, res.Stdout)
		if res.ExitCode == 0 && count > 0 {
			return item.source, count, "resolved", nil
		}
		detail := strings.TrimSpace(res.Stderr)
		if detail == "" {
			detail = strings.TrimSpace(res.Stdout)
		}
		if detail == "" {
			detail = fmt.Sprintf("exit=%d no records", res.ExitCode)
		}
		details = append(details, fmt.Sprintf("%s: %s", item.source, detail))
	}

	if !anyExecuted {
		return "unknown", 0, "", errors.Join(runErrs...)
	}
	if len(details) == 0 {
		details = append(details, "resolver returned no records")
	}
	return lastSource, 0, strings.Join(details, "; "), nil
}

func countDNSRecords(source string, output string) int {
	lines := strings.Split(output, "\n")
	count := 0
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		switch source {
		case "getent":
			count++
		case "nslookup":
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "address:") || strings.Contains(lower, " has address ") {
				count++
			}
		default:
			count++
		}
	}
	return count
}
