package schema

import (
	"encoding/json"
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
			return fmt.Errorf("template.stages.%s: invalid stage id (expect A-F)", stageID)
		}
		for i, c := range stage.Checks {
			if c.ID == "" {
				return fmt.Errorf("template.stages.%s.checks[%d].id is required", stageID, i)
			}
			if c.Kind == "" {
				return fmt.Errorf("template.stages.%s.checks[%d].kind is required", stageID, i)
			}
			if err := validateStepParams(stageID, "checks", i, c); err != nil {
				return err
			}
		}
		for i, a := range stage.Actions {
			if a.ID == "" {
				return fmt.Errorf("template.stages.%s.actions[%d].id is required", stageID, i)
			}
			if a.Kind == "" {
				return fmt.Errorf("template.stages.%s.actions[%d].kind is required", stageID, i)
			}
			if err := validateStepParams(stageID, "actions", i, a); err != nil {
				return err
			}
		}
		for i, e := range stage.Evidence {
			if e.ID == "" {
				return fmt.Errorf("template.stages.%s.evidence[%d].id is required", stageID, i)
			}
			if e.Kind == "" {
				return fmt.Errorf("template.stages.%s.evidence[%d].kind is required", stageID, i)
			}
			if err := validateStepParams(stageID, "evidence", i, e); err != nil {
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
	case "array", "object", "map":
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
	case "array":
		var out []any
		if err := json.Unmarshal([]byte(strings.TrimSpace(val)), &out); err != nil {
			return fmt.Errorf("template.vars.%s expects json array", name)
		}
	case "object", "map":
		var out map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(val)), &out); err != nil {
			return fmt.Errorf("template.vars.%s expects json object", name)
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

func validateStepParams(stageID, stepType string, index int, step TemplateStep) error {
	if step.Params == nil {
		return nil
	}
	if raw, ok := step.Params["severity"]; ok {
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("template.stages.%s.%s[%d].params.severity must be string", stageID, stepType, index)
		}
		if !IsValidSeverity(Severity(s)) {
			return fmt.Errorf("template.stages.%s.%s[%d].params.severity invalid value: %s", stageID, stepType, index, s)
		}
	}
	if path, token, ok := findUnresolvedVar(step.Params, fmt.Sprintf("template.stages.%s.%s[%d].params", stageID, stepType, index)); ok {
		return fmt.Errorf("%s: unresolved var %s", path, token)
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
		if s.Summary != nil {
			if s.Summary.Total < 0 || s.Summary.Pass < 0 || s.Summary.Warn < 0 || s.Summary.Fail < 0 || s.Summary.Skip < 0 {
				return fmt.Errorf("invalid stage summary count for %s", s.StageID)
			}
			sum := s.Summary.Pass + s.Summary.Warn + s.Summary.Fail + s.Summary.Skip
			if s.Summary.Total != sum {
				return fmt.Errorf("invalid stage summary totals for %s: total=%d sum=%d", s.StageID, s.Summary.Total, sum)
			}
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
