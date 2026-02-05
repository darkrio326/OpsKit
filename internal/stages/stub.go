package stages

import (
	"context"
	"encoding/json"
	"fmt"

	"opskit/internal/engine"
	"opskit/internal/schema"
	"opskit/internal/templates"
)

type StubStage struct {
	ID   string
	Name string
}

func (s *StubStage) StageID() string { return s.ID }

func (s *StubStage) Execute(_ context.Context, rt *engine.Runtime, _ templates.StagePlan) (engine.StageResult, error) {
	result := engine.StageResult{
		StageID: s.ID,
		Status:  schema.StatusSkipped,
		Metrics: []schema.Metric{{Label: "implementation", Value: "stub"}},
		Issues:  []schema.Issue{},
	}

	payload, _ := json.MarshalIndent(map[string]any{
		"stage":   s.ID,
		"name":    s.Name,
		"status":  result.Status,
		"message": "stub stage for Milestone 1",
	}, "", "  ")
	report := engine.ReportName(s.ID)
	if err := rt.Store.WriteReportStub(report, fmt.Sprintf("%s Stage (Stub)", s.Name), string(payload)); err != nil {
		return engine.StageResult{}, err
	}
	result.Report = report
	return result, nil
}

func DefaultExecutors() []engine.StageExecutor {
	return []engine.StageExecutor{
		&PreflightStage{},
		&BaselineStage{},
		&DeployStage{},
		&OperateStage{},
		&RecoverStage{},
		&AcceptStage{},
	}
}
