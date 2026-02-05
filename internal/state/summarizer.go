package state

import (
	"path/filepath"
	"time"

	"opskit/internal/recover"
	"opskit/internal/schema"
)

func DeriveOverall(lifecycle schema.LifecycleState) (schema.OverallStatus, int) {
	issues := 0
	hasWarn := false
	for _, s := range lifecycle.Stages {
		switch s.Status {
		case schema.StatusFailed:
			issues++
		case schema.StatusWarn:
			issues++
			hasWarn = true
		}
	}
	if issues == 0 {
		return schema.OverallHealthy, 0
	}
	if hasWarn {
		return schema.OverallDegraded, issues
	}
	return schema.OverallUnhealthy, issues
}

func DeriveRecoverSummary(paths Paths, lifecycle schema.LifecycleState) *schema.RecoverSummary {
	circuit, err := recover.Load(recoverCircuitPath(paths))
	if err != nil {
		return nil
	}
	stage, ok := findStage(lifecycle, "E")
	stageHasData := ok && (stage.Status != schema.StatusNotStarted || stage.LastRunTime != "" || metricValue(stage.Metrics, "recover_trigger") != "")
	if !stageHasData && circuit.SuccessCount == 0 && circuit.FailureCount == 0 && circuit.WarnCount == 0 && circuit.LastRunTime == "" {
		return nil
	}

	s := &schema.RecoverSummary{
		SuccessCount:  circuit.SuccessCount,
		FailureCount:  circuit.FailureCount,
		WarnCount:     circuit.WarnCount,
		CooldownUntil: circuit.CooldownUntil,
	}
	open, _ := recover.IsOpen(circuit, time.Now())
	s.CircuitOpen = open

	if stageHasData {
		s.LastStatus = stage.Status
		s.LastRunTime = stage.LastRunTime
		if trigger := metricValue(stage.Metrics, "recover_trigger"); trigger != "" {
			s.LastTrigger = trigger
		}
	}
	if s.LastStatus == "" && circuit.LastStatus != "" {
		s.LastStatus = schema.Status(circuit.LastStatus)
	}
	if s.LastRunTime == "" {
		s.LastRunTime = circuit.LastRunTime
	}
	if s.LastTrigger == "" {
		s.LastTrigger = circuit.LastTrigger
	}
	return s
}

func recoverCircuitPath(paths Paths) string {
	return filepath.Join(paths.StateDir, "recover_circuit.json")
}

func findStage(lifecycle schema.LifecycleState, id string) (schema.StageState, bool) {
	for _, s := range lifecycle.Stages {
		if s.StageID == id {
			return s, true
		}
	}
	return schema.StageState{}, false
}

func metricValue(metrics []schema.Metric, label string) string {
	for _, m := range metrics {
		if m.Label == label {
			return m.Value
		}
	}
	return ""
}
