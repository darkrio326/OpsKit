package checks

import (
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type fsReadonlyCheck struct{}

func (c *fsReadonlyCheck) Kind() string { return "fs_readonly" }

func (c *fsReadonlyCheck) Run(ctx context.Context, req Request) (Result, error) {
	targets := []string{"/"}
	if v, ok := req.Params["mounts"]; ok {
		targets = toStringSlice(v, targets)
	}
	allowReadonly := []string{}
	if v, ok := req.Params["allow_readonly"]; ok {
		allowReadonly = toStringSlice(v, allowReadonly)
	}
	requirePresent := toBool(req.Params["require_present"], false)
	sev := schema.SeverityWarn
	if v, ok := req.Params["severity"]; ok {
		sev = toSeverity(v, sev)
	}

	allowSet := map[string]bool{}
	for _, p := range allowReadonly {
		allowSet[p] = true
	}

	mountFile := toString(req.Params["mounts_file"], "/proc/mounts")
	mounts, source, degradeReason, err := readMountOptions(ctx, req, mountFile)
	baseMetrics := []schema.Metric{
		{Label: "fs_mount_source", Value: source},
		{Label: "fs_target_mounts", Value: fmt.Sprintf("%d", len(targets))},
	}
	if err != nil {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "filesystem readonly check degraded: " + err.Error(),
			Advice:   "verify mount state manually",
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

	checked := 0
	missing := []string{}
	readonlyViolations := []string{}
	readonlyAllowed := 0
	for _, m := range targets {
		opts, ok := mounts[m]
		if !ok {
			if requirePresent {
				missing = append(missing, m)
			}
			continue
		}
		checked++
		isReadOnly := opts["ro"] && !opts["rw"]
		if !isReadOnly {
			continue
		}
		if allowSet[m] {
			readonlyAllowed++
			continue
		}
		readonlyViolations = append(readonlyViolations, m)
	}

	metrics := withHealthyMetrics([]schema.Metric{
		{Label: "fs_mount_source", Value: source},
		{Label: "fs_checked_mounts", Value: fmt.Sprintf("%d", checked)},
		{Label: "fs_missing_mounts", Value: fmt.Sprintf("%d", len(missing))},
		{Label: "fs_readonly_violations", Value: fmt.Sprintf("%d", len(readonlyViolations))},
		{Label: "fs_readonly_allowed", Value: fmt.Sprintf("%d", readonlyAllowed)},
	})

	if len(readonlyViolations) == 0 && len(missing) == 0 {
		return Result{
			CheckID:  req.ID,
			Status:   schema.StatusPassed,
			Severity: schema.SeverityInfo,
			Message:  "filesystem mount options healthy",
			Metrics:  metrics,
		}, nil
	}

	parts := []string{}
	if len(readonlyViolations) > 0 {
		parts = append(parts, "unexpected readonly mounts: "+strings.Join(readonlyViolations, ","))
	}
	if len(missing) > 0 {
		parts = append(parts, "required mounts missing: "+strings.Join(missing, ","))
	}
	msg := strings.Join(parts, "; ")
	issue := &schema.Issue{
		ID:       req.ID,
		Severity: sev,
		Message:  msg,
		Advice:   "check mount flags and remount rw where appropriate",
	}
	return Result{
		CheckID:  req.ID,
		Status:   statusFromSeverity(sev),
		Severity: sev,
		Message:  msg,
		Issue:    issue,
		Metrics:  metrics,
	}, nil
}

func readMountOptions(ctx context.Context, req Request, mountsFile string) (map[string]map[string]bool, string, string, error) {
	content, err := os.ReadFile(mountsFile)
	if err == nil {
		parsed, parseErr := parseProcMountOptions(string(content))
		if parseErr == nil {
			return parsed, "proc_mounts", "", nil
		}
		return nil, "proc_mounts", "parse_error", parseErr
	}

	out, runErr := req.Exec.Run(ctx, executil.Spec{Name: "mount", Timeout: 8 * time.Second})
	if runErr != nil {
		reason := "mount_probe_failed"
		if errors.Is(runErr, osexec.ErrNotFound) {
			reason = "tool_unavailable"
		}
		return nil, "mount_cmd", reason, runErr
	}
	parsed, parseErr := parseMountCmdOptions(out.Stdout)
	if parseErr != nil {
		return nil, "mount_cmd", "parse_error", parseErr
	}
	return parsed, "mount_cmd", "", nil
}

func parseProcMountOptions(content string) (map[string]map[string]bool, error) {
	mounts := map[string]map[string]bool{}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		target := fields[1]
		opts := parseMountOptionSet(fields[3])
		mounts[target] = opts
	}
	return mounts, nil
}

func parseMountCmdOptions(content string) (map[string]map[string]bool, error) {
	mounts := map[string]map[string]bool{}
	lines := strings.Split(strings.TrimSpace(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " on ")
		if len(parts) < 2 {
			continue
		}
		targetPart := parts[1]
		target := strings.SplitN(targetPart, " type ", 2)[0]
		open := strings.LastIndex(line, "(")
		close := strings.LastIndex(line, ")")
		if open < 0 || close <= open {
			continue
		}
		optsRaw := line[open+1 : close]
		mounts[target] = parseMountOptionSet(optsRaw)
	}
	return mounts, nil
}

func parseMountOptionSet(raw string) map[string]bool {
	out := map[string]bool{}
	for _, opt := range strings.Split(raw, ",") {
		opt = strings.TrimSpace(opt)
		if opt == "" {
			continue
		}
		if idx := strings.Index(opt, "="); idx >= 0 {
			opt = strings.TrimSpace(opt[:idx])
		}
		out[opt] = true
	}
	return out
}
