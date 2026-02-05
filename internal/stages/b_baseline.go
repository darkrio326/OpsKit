package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type BaselineStage struct{}

func (s *BaselineStage) StageID() string { return "B" }

func (s *BaselineStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeActionStage(ctx, rt, "B", "Baseline Report", stage)
}
