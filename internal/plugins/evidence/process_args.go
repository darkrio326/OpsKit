package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"opskit/internal/core/redaction"
	"opskit/internal/schema"
)

type processArgsEvidence struct{}

func (p *processArgsEvidence) Kind() string { return "process_args" }

func (p *processArgsEvidence) Collect(_ context.Context, req Request) (Result, error) {
	match := toString(req.Params["match"], "")
	if match == "" {
		match = toString(req.Params["process_match"], "")
	}
	output := toString(req.Params["output"], "")
	if match == "" || output == "" {
		return Result{}, fmt.Errorf("process_args requires params.match (or process_match) and params.output")
	}

	limit := toInt(req.Params["limit"], toInt(req.Params["max_results"], 10))
	if limit < 1 {
		limit = 10
	}
	redactKeys := toStringSlice(req.Params["redact_keys"])
	records, err := scanProcessArgs(match, limit, redactKeys)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			payload := map[string]any{
				"evidenceId": req.ID,
				"kind":       "process_args",
				"match":      match,
				"generated":  time.Now().Format(time.RFC3339),
				"error":      err.Error(),
				"processes":  []any{},
			}
			if err := ensureParent(output); err != nil {
				return Result{}, err
			}
			b, _ := json.MarshalIndent(payload, "", "  ")
			if err := os.WriteFile(output, b, 0o644); err != nil {
				return Result{}, err
			}
			issue := &schema.Issue{
				ID:       req.ID,
				Severity: schema.SeverityWarn,
				Message:  "process list unavailable for evidence",
				Advice:   "run on Linux /proc environment or disable process_args step",
			}
			return Result{
				EvidenceID: req.ID,
				Status:     schema.StatusWarn,
				Severity:   schema.SeverityWarn,
				Message:    issue.Message,
				Path:       output,
				Metrics:    []schema.Metric{{Label: "process_matches", Value: "0"}},
				Issue:      issue,
			}, nil
		}
		return Result{}, err
	}

	payload := map[string]any{
		"evidenceId": req.ID,
		"kind":       "process_args",
		"match":      match,
		"generated":  time.Now().Format(time.RFC3339),
		"processes":  records,
	}
	if err := ensureParent(output); err != nil {
		return Result{}, err
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(output, b, 0o644); err != nil {
		return Result{}, err
	}

	if len(records) == 0 {
		issue := &schema.Issue{
			ID:       req.ID,
			Severity: schema.SeverityWarn,
			Message:  "no matching process found",
			Advice:   "check process name and runtime state",
		}
		return Result{
			EvidenceID: req.ID,
			Status:     schema.StatusWarn,
			Severity:   schema.SeverityWarn,
			Message:    issue.Message,
			Path:       output,
			Metrics:    []schema.Metric{{Label: "process_matches", Value: "0"}},
			Issue:      issue,
		}, nil
	}

	return Result{
		EvidenceID: req.ID,
		Status:     schema.StatusPassed,
		Severity:   schema.SeverityInfo,
		Message:    "process args collected",
		Path:       output,
		Metrics:    []schema.Metric{{Label: "process_matches", Value: fmt.Sprintf("%d", len(records))}},
	}, nil
}

type processRecord struct {
	PID     int      `json:"pid"`
	Exe     string   `json:"exe,omitempty"`
	Args    []string `json:"args"`
	Command string   `json:"command"`
}

func scanProcessArgs(match string, limit int, redactKeys []string) ([]processRecord, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	matchLower := strings.ToLower(strings.TrimSpace(match))
	out := make([]processRecord, 0, limit)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		cmdline, err := os.ReadFile("/proc/" + e.Name() + "/cmdline")
		if err != nil || len(cmdline) == 0 {
			continue
		}
		args := parseCmdline(cmdline)
		if len(args) == 0 {
			continue
		}
		command := strings.Join(args, " ")
		if !strings.Contains(strings.ToLower(command), matchLower) {
			continue
		}

		exe, _ := os.Readlink("/proc/" + e.Name() + "/exe")
		redactedArgs := redaction.RedactArgs(args, redactKeys...)
		out = append(out, processRecord{
			PID:     pid,
			Exe:     exe,
			Args:    redactedArgs,
			Command: strings.Join(redactedArgs, " "),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].PID < out[j].PID })
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func parseCmdline(cmdline []byte) []string {
	parts := strings.Split(string(cmdline), "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
