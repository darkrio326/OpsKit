package templates

import (
	"testing"

	"opskit/internal/schema"
)

func TestBuildPlanWithOptionsIncludeDisabled(t *testing.T) {
	disabled := false
	tpl := schema.Template{
		ID:   "t",
		Name: "t",
		Mode: "deploy",
		Stages: map[string]schema.TemplateStageSpec{
			"C": {
				Actions: []schema.TemplateStep{
					{ID: "a1", Kind: "ensure_paths", Enabled: &disabled},
				},
			},
		},
	}

	planDefault := BuildPlanWithOptions(tpl, []string{"C"}, false)
	if len(planDefault.Stages) != 1 || len(planDefault.Stages[0].Actions) != 0 {
		t.Fatalf("expected disabled action excluded, got %+v", planDefault.Stages)
	}

	planFix := BuildPlanWithOptions(tpl, []string{"C"}, true)
	if len(planFix.Stages) != 1 || len(planFix.Stages[0].Actions) != 1 {
		t.Fatalf("expected disabled action included with includeDisabled=true, got %+v", planFix.Stages)
	}
}
