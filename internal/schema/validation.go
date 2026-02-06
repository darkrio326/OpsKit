package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var templateVarRef = regexp.MustCompile(`\$\{[A-Za-z0-9_]+\}`)

func ValidateTemplate(t Template) error {
	if t.ID == "" {
		return fmt.Errorf("template.id is required")
	}
	if t.Name == "" {
		return fmt.Errorf("template.name is required")
	}
	if t.Mode != "manage" && t.Mode != "deploy" {
		return fmt.Errorf("template.mode must be manage or deploy")
	}
	if len(t.Stages) == 0 {
		return fmt.Errorf("template.stages is required")
	}
	if err := validateVarSpecs(t.Vars); err != nil {
		return err
	}
	for stageID, stage := range t.Stages {
		if !isStageID(stageID) {
			return fmt.Errorf("invalid stage id: %s (expect A-F)", stageID)
		}
		for _, c := range stage.Checks {
			if c.ID == "" || c.Kind == "" {
				return fmt.Errorf("stage %s check id/kind is required", stageID)
			}
			if err := validateStepParams(stageID, "check", c); err != nil {
				return err
			}
		}
		for _, a := range stage.Actions {
			if a.ID == "" || a.Kind == "" {
				return fmt.Errorf("stage %s action id/kind is required", stageID)
			}
			if err := validateStepParams(stageID, "action", a); err != nil {
				return err
			}
		}
		for _, e := range stage.Evidence {
			if e.ID == "" || e.Kind == "" {
				return fmt.Errorf("stage %s evidence id/kind is required", stageID)
			}
			if err := validateStepParams(stageID, "evidence", e); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateVars(specs map[string]VarSpec, vars map[string]string) error {
	if err := validateVarSpecs(specs); err != nil {
		return err
	}
	if len(specs) == 0 {
		return nil
	}
	for name, spec := range specs {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("template.vars has empty variable name")
		}
		val, ok := vars[name]
		if !ok || strings.TrimSpace(val) == "" {
			if spec.Required {
				return fmt.Errorf("template.vars.%s is required", name)
			}
			continue
		}
		if err := validateVarValue(name, spec, val); err != nil {
			return err
		}
	}
	return nil
}

func validateVarSpecs(specs map[string]VarSpec) error {
	if len(specs) == 0 {
		return nil
	}
	for name, spec := range specs {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("template.vars has empty variable name")
		}
		if err := validateVarType(name, spec.Type); err != nil {
			return err
		}
		if len(spec.Enum) > 0 {
			seen := map[string]struct{}{}
			for _, v := range spec.Enum {
				v = strings.TrimSpace(v)
				if v == "" {
					return fmt.Errorf("template.vars.%s enum contains empty value", name)
				}
				if _, ok := seen[v]; ok {
					return fmt.Errorf("template.vars.%s enum contains duplicate value: %s", name, v)
				}
				seen[v] = struct{}{}
			}
			if spec.Default != "" {
				if !containsString(spec.Enum, spec.Default) {
					return fmt.Errorf("template.vars.%s default not in enum: %s", name, spec.Default)
				}
			}
		}
		if spec.Default != "" {
			if err := validateVarValue(name, spec, spec.Default); err != nil {
				return fmt.Errorf("template.vars.%s default invalid: %w", name, err)
			}
		}
	}
	return nil
}

func validateVarType(name, typ string) error {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "", "string":
		return nil
	case "int", "integer", "number", "float", "bool", "boolean":
		return nil
	default:
		return fmt.Errorf("template.vars.%s invalid type: %s", name, typ)
	}
}

func validateVarValue(name string, spec VarSpec, val string) error {
	if len(spec.Enum) > 0 && !containsString(spec.Enum, val) {
		return fmt.Errorf("template.vars.%s invalid value: %s", name, val)
	}
	switch strings.ToLower(strings.TrimSpace(spec.Type)) {
	case "", "string":
		return nil
	case "int", "integer":
		if _, err := strconv.Atoi(strings.TrimSpace(val)); err != nil {
			return fmt.Errorf("template.vars.%s expects int", name)
		}
	case "number", "float":
		if _, err := strconv.ParseFloat(strings.TrimSpace(val), 64); err != nil {
			return fmt.Errorf("template.vars.%s expects number", name)
		}
	case "bool", "boolean":
		if _, err := strconv.ParseBool(strings.TrimSpace(val)); err != nil {
			return fmt.Errorf("template.vars.%s expects bool", name)
		}
	default:
		return fmt.Errorf("template.vars.%s invalid type: %s", name, spec.Type)
	}
	return nil
}

func containsString(xs []string, v string) bool {
	for _, s := range xs {
		if s == v {
			return true
		}
	}
	return false
}

func validateStepParams(stageID, stepType string, step TemplateStep) error {
	if step.Params == nil {
		return nil
	}
	if raw, ok := step.Params["severity"]; ok {
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("stage %s %s %s severity must be string", stageID, stepType, step.ID)
		}
		if !IsValidSeverity(Severity(s)) {
			return fmt.Errorf("stage %s %s %s invalid severity: %s", stageID, stepType, step.ID, s)
		}
	}
	if path, token, ok := findUnresolvedVar(step.Params, "params"); ok {
		return fmt.Errorf("stage %s %s %s unresolved var %s at %s", stageID, stepType, step.ID, token, path)
	}
	return nil
}

func findUnresolvedVar(value any, path string) (string, string, bool) {
	switch v := value.(type) {
	case string:
		if token := templateVarRef.FindString(v); token != "" {
			return path, token, true
		}
	case []any:
		for i, item := range v {
			if p, token, ok := findUnresolvedVar(item, fmt.Sprintf("%s[%d]", path, i)); ok {
				return p, token, ok
			}
		}
	case map[string]any:
		for k, item := range v {
			if p, token, ok := findUnresolvedVar(item, path+"."+k); ok {
				return p, token, ok
			}
		}
	}
	return "", "", false
}

func IsValidStatus(v Status) bool {
	switch v {
	case StatusNotStarted, StatusRunning, StatusPassed, StatusWarn, StatusFailed, StatusSkipped:
		return true
	default:
		return false
	}
}

func IsValidSeverity(v Severity) bool {
	switch v {
	case SeverityInfo, SeverityWarn, SeverityFail:
		return true
	default:
		return false
	}
}

func IsValidOverallStatus(v OverallStatus) bool {
	switch v {
	case OverallHealthy, OverallDegraded, OverallUnhealthy, OverallUnknown:
		return true
	default:
		return false
	}
}

func ValidateLifecycleState(l LifecycleState) error {
	for _, s := range l.Stages {
		if !isStageID(s.StageID) {
			return fmt.Errorf("invalid stage id: %s (expect A-F)", s.StageID)
		}
		if !IsValidStatus(s.Status) {
			return fmt.Errorf("invalid stage status: %s for %s", s.Status, s.StageID)
		}
		for _, i := range s.Issues {
			if !IsValidSeverity(i.Severity) {
				return fmt.Errorf("invalid issue severity: %s in stage %s", i.Severity, s.StageID)
			}
		}
	}
	return nil
}

func isStageID(id string) bool {
	switch id {
	case "A", "B", "C", "D", "E", "F":
		return true
	default:
		return false
	}
}
