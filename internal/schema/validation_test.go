package schema

import "testing"

func TestValidateTemplate_UnresolvedVar(t *testing.T) {
	tpl := Template{
		ID:   "t",
		Name: "t",
		Mode: "manage",
		Stages: map[string]TemplateStageSpec{
			"A": {
				Checks: []TemplateStep{
					{
						ID:   "a.system_info",
						Kind: "system_info",
						Params: map[string]any{
							"output": "${MISSING_VAR}",
						},
					},
				},
			},
		},
	}
	if err := ValidateTemplate(tpl); err == nil {
		t.Fatalf("expected unresolved var error")
	}
}

func TestValidateTemplate_SeverityEnum(t *testing.T) {
	tpl := Template{
		ID:   "t",
		Name: "t",
		Mode: "manage",
		Stages: map[string]TemplateStageSpec{
			"A": {
				Checks: []TemplateStep{
					{
						ID:   "a.system_info",
						Kind: "system_info",
						Params: map[string]any{
							"severity": "nope",
						},
					},
				},
			},
		},
	}
	if err := ValidateTemplate(tpl); err == nil {
		t.Fatalf("expected invalid severity error")
	}
}
