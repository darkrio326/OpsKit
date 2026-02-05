package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type RecoverStage struct{}

func (s *RecoverStage) StageID() string { return "E" }

func (s *RecoverStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeActionStage(ctx, rt, "E", "Recover Report", stage)
}
