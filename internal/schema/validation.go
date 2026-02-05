package schema

import "fmt"

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
	for stageID, stage := range t.Stages {
		if !isStageID(stageID) {
			return fmt.Errorf("invalid stage id: %s (expect A-F)", stageID)
		}
		for _, c := range stage.Checks {
			if c.ID == "" || c.Kind == "" {
				return fmt.Errorf("stage %s check id/kind is required", stageID)
			}
		}
		for _, a := range stage.Actions {
			if a.ID == "" || a.Kind == "" {
				return fmt.Errorf("stage %s action id/kind is required", stageID)
			}
		}
		for _, e := range stage.Evidence {
			if e.ID == "" || e.Kind == "" {
				return fmt.Errorf("stage %s evidence id/kind is required", stageID)
			}
		}
	}
	return nil
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
