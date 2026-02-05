package templates

import "testing"

func TestSingleServiceDeployTemplateHasDeployActionsEnabled(t *testing.T) {
	tpl, _, err := Resolve(ResolveOptions{TemplateRef: "single-service-deploy", BaseDir: "/var/lib/opskit"})
	if err != nil {
		t.Fatalf("resolve template: %v", err)
	}
	plan := BuildPlanWithOptions(tpl, []string{"C"}, false)
	if len(plan.Stages) != 1 {
		t.Fatalf("expected 1 stage in plan, got %d", len(plan.Stages))
	}
	if len(plan.Stages[0].Actions) < 8 {
		t.Fatalf("expected deploy stage to include deploy actions, got %d", len(plan.Stages[0].Actions))
	}
}
