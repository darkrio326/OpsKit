package engine

import (
	"context"

	"opskit/internal/core/executil"
	"opskit/internal/plugins/actions"
	"opskit/internal/plugins/checks"
	"opskit/internal/plugins/evidence"
	"opskit/internal/schema"
	"opskit/internal/state"
	"opskit/internal/templates"
)

type RunOptions struct {
	SelectedStages []string
	TemplateID     string
	TemplateMode   string
	DryRun         bool
}

type Runtime struct {
	Store            *state.Store
	CheckRegistry    *checks.Registry
	ActionRegistry   *actions.Registry
	EvidenceRegistry *evidence.Registry
	Exec             executil.Runner
	Plan             templates.Plan
	Options          RunOptions
}

type StageResult struct {
	StageID      string
	Status       schema.Status
	Metrics      []schema.Metric
	Issues       []schema.Issue
	Checks       []schema.CheckState
	StepStatuses []schema.Status
	Report       string
	Bundles      []schema.ArtifactRef
}

type StageExecutor interface {
	StageID() string
	Execute(ctx context.Context, rt *Runtime, stage templates.StagePlan) (StageResult, error)
}
