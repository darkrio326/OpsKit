package stages

import (
	"context"

	"opskit/internal/engine"
	"opskit/internal/templates"
)

type OperateStage struct{}

func (s *OperateStage) StageID() string { return "D" }

func (s *OperateStage) Execute(ctx context.Context, rt *engine.Runtime, stage templates.StagePlan) (engine.StageResult, error) {
	return executeCheckStage(ctx, rt, "D", "Operate Report", stage)
}
