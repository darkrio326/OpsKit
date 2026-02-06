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

func TestValidateVars_RequiredAndEnum(t *testing.T) {
	specs := map[string]VarSpec{
		"ENV":   {Type: "string", Required: true, Enum: []string{"dev", "prod"}},
		"PORT":  {Type: "int", Required: true},
		"PORTS": {Type: "array"},
		"META":  {Type: "object"},
	}
	if err := ValidateVars(specs, map[string]string{"ENV": "dev", "PORT": "18080", "PORTS": "[80,443]", "META": "{\"env\":\"test\"}"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateVars(specs, map[string]string{"ENV": "test", "PORT": "18080"}); err == nil {
		t.Fatalf("expected enum validation error")
	}
	if err := ValidateVars(specs, map[string]string{"ENV": "dev"}); err == nil {
		t.Fatalf("expected required validation error")
	}
	if err := ValidateVars(specs, map[string]string{"ENV": "dev", "PORT": "18080", "PORTS": "80,443"}); err == nil {
		t.Fatalf("expected array validation error")
	}
}

func TestValidateLifecycleState_Summary(t *testing.T) {
	lifecycle := LifecycleState{
		Stages: []StageState{
			{
				StageID: "A",
				Status:  StatusPassed,
				Summary: &StageSummary{
					Total: 3,
					Pass:  1,
					Warn:  1,
					Fail:  0,
					Skip:  1,
				},
			},
		},
	}
	if err := ValidateLifecycleState(lifecycle); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lifecycle.Stages[0].Summary.Total = 2
	if err := ValidateLifecycleState(lifecycle); err == nil {
		t.Fatalf("expected summary total mismatch error")
	}
}
