package actions

import (
	"context"
	"fmt"

	"opskit/internal/core/timeutil"
	"opskit/internal/schema"
)

type declareStackAction struct{}

func (a *declareStackAction) Kind() string { return "declare_stack" }

func (a *declareStackAction) Run(_ context.Context, req Request) (Result, error) {
	output := toString(req.Params["output"], "")
	stackID := toString(req.Params["stack_id"], "")
	mode := toString(req.Params["mode"], "")
	if output == "" || stackID == "" || mode == "" {
		return Result{}, fmt.Errorf("declare_stack requires params.output/stack_id/mode")
	}

	entry := map[string]any{
		"stackId":    stackID,
		"mode":       mode,
		"declaredAt": timeutil.NowISO8601(),
		"components": toStringSlice(req.Params["components"]),
	}
	if err := writeJSON(output, entry); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "write stack declaration failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "stack declared: " + stackID}, nil
}
