package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"opskit/internal/schema"
)

type systemdInstallUnitAction struct{}

func (a *systemdInstallUnitAction) Kind() string { return "systemd_install_unit" }

func (a *systemdInstallUnitAction) Run(_ context.Context, req Request) (Result, error) {
	unitName := toString(req.Params["unit_name"], "")
	content := toString(req.Params["unit_content"], "")
	dir := toString(req.Params["unit_dir"], "/etc/systemd/system")
	if unitName == "" || content == "" {
		return Result{}, fmt.Errorf("systemd_install_unit requires params.unit_name and params.unit_content")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Result{}, err
	}
	target := filepath.Join(dir, unitName)
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "write unit failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "unit installed: " + target}, nil
}
