package actions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"opskit/internal/schema"
)

type renderTemplatesAction struct{}

func (a *renderTemplatesAction) Kind() string { return "render_templates" }

func (a *renderTemplatesAction) Run(_ context.Context, req Request) (Result, error) {
	tpl := toString(req.Params["template"], "")
	output := toString(req.Params["output"], "")
	if tpl == "" || output == "" {
		return Result{}, fmt.Errorf("render_templates requires params.template and params.output")
	}

	vars, _ := req.Params["vars"].(map[string]any)
	for k, v := range vars {
		key := "${" + strings.TrimSpace(k) + "}"
		tpl = strings.ReplaceAll(tpl, key, toString(v, ""))
	}

	if err := ensureParent(output); err != nil {
		return Result{}, err
	}
	perm := toFileMode(req.Params["perm"], 0o644)
	if err := os.WriteFile(output, []byte(tpl), perm); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "write rendered template failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "template rendered: " + output}, nil
}
