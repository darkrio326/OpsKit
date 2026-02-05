package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type PreflightStage struct{}

func (s *PreflightStage) StageID() string { return "A" }

func (s *PreflightStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeCheckStage(ctx, rt, "A", "Preflight Report", stage)
}
