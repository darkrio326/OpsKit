package engine

import (
	"strings"
	"testing"

	actionplugin "opskit/internal/plugins/actions"
	checkplugin "opskit/internal/plugins/checks"
	evidenceplugin "opskit/internal/plugins/evidence"
	"opskit/internal/schema"
	"opskit/internal/templates"
)

func TestValidatePlanPluginsPass(t *testing.T) {
	checkReg := checkplugin.NewRegistry()
	checkplugin.RegisterBuiltins(checkReg)
	actionReg := actionplugin.NewRegistry()
	actionplugin.RegisterBuiltins(actionReg)
	evidenceReg := evidenceplugin.NewRegistry()
	evidenceplugin.RegisterBuiltins(evidenceReg)

	plan := templates.Plan{
		TemplateID: "demo",
		Stages: []templates.StagePlan{
			{
				StageID:  "A",
				Checks:   []schema.TemplateStep{{ID: "a.system_info", Kind: "system_info"}},
				Actions:  []schema.TemplateStep{{ID: "a.ensure_paths", Kind: "ensure_paths"}},
				Evidence: []schema.TemplateStep{{ID: "a.command_output", Kind: "command_output"}},
			},
		},
	}
	if err := ValidatePlanPlugins(plan, checkReg, actionReg, evidenceReg); err != nil {
		t.Fatalf("validate plan plugins: %v", err)
	}
}

func TestValidatePlanPluginsFail(t *testing.T) {
	checkReg := checkplugin.NewRegistry()
	checkplugin.RegisterBuiltins(checkReg)
	actionReg := actionplugin.NewRegistry()
	actionplugin.RegisterBuiltins(actionReg)
	evidenceReg := evidenceplugin.NewRegistry()
	evidenceplugin.RegisterBuiltins(evidenceReg)

	plan := templates.Plan{
		TemplateID: "demo",
		Stages: []templates.StagePlan{
			{
				StageID:  "A",
				Checks:   []schema.TemplateStep{{ID: "a.bad_check", Kind: "bad_check"}},
				Actions:  []schema.TemplateStep{{ID: "a.bad_action", Kind: "bad_action"}},
				Evidence: []schema.TemplateStep{{ID: "a.bad_evidence", Kind: "bad_evidence"}},
			},
		},
	}
	err := ValidatePlanPlugins(plan, checkReg, actionReg, evidenceReg)
	if err == nil {
		t.Fatalf("expected error")
	}
	msg := err.Error()
	for _, want := range []string{
		"template.stages.A.checks[0].kind unsupported",
		"template.stages.A.actions[0].kind unsupported",
		"template.stages.A.evidence[0].kind unsupported",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("expected message contains %q, got %q", want, msg)
		}
	}
}
