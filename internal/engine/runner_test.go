package engine

import (
	"context"
	"slices"
	"testing"

	"opskit/internal/schema"
	"opskit/internal/state"
	"opskit/internal/templates"
)

func TestSummarizeStageSteps(t *testing.T) {
	summary := summarizeStageSteps([]schema.Status{
		schema.StatusPassed,
		schema.StatusWarn,
		schema.StatusFailed,
		schema.StatusSkipped,
		schema.StatusNotStarted,
		schema.StatusRunning,
	})
	if summary.Total != 6 {
		t.Fatalf("expected total 6, got %d", summary.Total)
	}
	if summary.Pass != 1 || summary.Warn != 1 || summary.Fail != 1 || summary.Skip != 3 {
		t.Fatalf("unexpected summary: %+v", *summary)
	}
}

func TestRunnerPersistsAdditionalReports(t *testing.T) {
	paths := state.NewPaths(t.TempDir())
	store := state.NewStore(paths)
	rt := &Runtime{
		Store: store,
		Plan: templates.Plan{
			TemplateID: "demo",
			Stages: []templates.StagePlan{
				{StageID: "F"},
			},
		},
		Options: RunOptions{
			TemplateID:   "demo",
			TemplateMode: "manage",
		},
	}
	r := NewRunner([]StageExecutor{
		fakeStageExecutor{
			id: "F",
			result: StageResult{
				StageID: "F",
				Status:  schema.StatusPassed,
				Report:  "accept-20260207-000000.html",
				Reports: []schema.ArtifactRef{
					{ID: "acceptance-consistency", Path: "evidence/acceptance-consistency-20260207-000000.json"},
				},
			},
		},
	})

	if _, err := r.Execute(context.Background(), rt); err != nil {
		t.Fatalf("execute runner: %v", err)
	}
	artifacts, err := store.ReadArtifacts()
	if err != nil {
		t.Fatalf("read artifacts: %v", err)
	}
	if !slices.ContainsFunc(artifacts.Reports, func(a schema.ArtifactRef) bool {
		return a.ID == "accept" && a.Path == "reports/accept-20260207-000000.html"
	}) {
		t.Fatalf("expected stage report indexed in artifacts: %+v", artifacts.Reports)
	}
	if !slices.ContainsFunc(artifacts.Reports, func(a schema.ArtifactRef) bool {
		return a.ID == "acceptance-consistency" && a.Path == "evidence/acceptance-consistency-20260207-000000.json"
	}) {
		t.Fatalf("expected additional report indexed in artifacts: %+v", artifacts.Reports)
	}
}

type fakeStageExecutor struct {
	id     string
	result StageResult
}

func (f fakeStageExecutor) StageID() string { return f.id }

func (f fakeStageExecutor) Execute(_ context.Context, _ *Runtime, _ templates.StagePlan) (StageResult, error) {
	return f.result, nil
}
