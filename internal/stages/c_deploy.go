package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type DeployStage struct{}

func (s *DeployStage) StageID() string { return "C" }

func (s *DeployStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeActionStage(ctx, rt, "C", "Deploy Report", stage)
}
