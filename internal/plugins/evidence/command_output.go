package evidence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/redaction"
	"opskit/internal/schema"
)

type commandOutputEvidence struct{}

func (p *commandOutputEvidence) Kind() string { return "command_output" }

func (p *commandOutputEvidence) Collect(ctx context.Context, req Request) (Result, error) {
	name := toString(req.Params["name"], "")
	args := toStringSlice(req.Params["args"])
	output := toString(req.Params["output"], "")
	redactKeys := toStringSlice(req.Params["redact_keys"])
	if name == "" || output == "" {
		return Result{}, fmt.Errorf("command_output requires params.name and params.output")
	}
	res, err := runCmd(ctx, req, name, args...)
	redactedArgs := redaction.RedactArgs(args, redactKeys...)
	commandLine := strings.Join(append([]string{name}, redactedArgs...), " ")
	if err != nil {
		if errors.Is(err, coreerr.ErrPreconditionFailed) {
			payload := map[string]any{
				"evidenceId": req.ID,
				"kind":       "command_output",
				"command":    commandLine,
				"error":      err.Error(),
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
				Message:  "command unavailable for evidence: " + name,
				Advice:   "install command or remove this evidence item",
			}
			return Result{
				EvidenceID: req.ID,
				Status:     schema.StatusWarn,
				Severity:   schema.SeverityWarn,
				Message:    issue.Message,
				Path:       output,
				Issue:      issue,
			}, nil
		}
		return Result{}, err
	}

	payload := map[string]any{
		"evidenceId": req.ID,
		"kind":       "command_output",
		"command":    commandLine,
		"exitCode":   res.ExitCode,
		"stdout":     redaction.RedactText(res.Stdout, redactKeys...),
		"stderr":     redaction.RedactText(res.Stderr, redactKeys...),
	}
	if err := ensureParent(output); err != nil {
		return Result{}, err
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(output, b, 0o644); err != nil {
		return Result{}, err
	}

	status := schema.StatusPassed
	sev := schema.SeverityInfo
	msg := "command output collected"
	if res.ExitCode != 0 {
		status = schema.StatusWarn
		sev = schema.SeverityWarn
		msg = fmt.Sprintf("command returned non-zero exit: %d", res.ExitCode)
	}
	return Result{EvidenceID: req.ID, Status: status, Severity: sev, Message: msg, Path: output}, nil
}
