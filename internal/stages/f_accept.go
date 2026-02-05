package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type AcceptStage struct{}

func (s *AcceptStage) StageID() string { return "F" }

func (s *AcceptStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeEvidenceStage(ctx, rt, "F", "Accept Report", stage)
}
