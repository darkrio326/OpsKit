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

func TestRunnerBlocksAfterPreflightFailure(t *testing.T) {
	paths := state.NewPaths(t.TempDir())
	store := state.NewStore(paths)
	calls := map[string]int{}
	rt := &Runtime{
		Store: store,
		Plan: templates.Plan{
			TemplateID: "demo",
			Stages: []templates.StagePlan{
				{StageID: "A"},
				{StageID: "B"},
				{StageID: "C"},
			},
		},
		Options: RunOptions{TemplateID: "demo", TemplateMode: "manage"},
	}
	r := NewRunner([]StageExecutor{
		recordingStageExecutor{id: "A", status: schema.StatusFailed, calls: calls},
		recordingStageExecutor{id: "B", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "C", status: schema.StatusPassed, calls: calls},
	})

	results, err := r.Execute(context.Background(), rt)
	if err != nil {
		t.Fatalf("execute runner: %v", err)
	}
	if got, want := len(results), 3; got != want {
		t.Fatalf("unexpected result count: %d", got)
	}
	if results[0].Status != schema.StatusFailed {
		t.Fatalf("expected stage A failed, got %s", results[0].Status)
	}
	if results[1].Status != schema.StatusSkipped || results[2].Status != schema.StatusSkipped {
		t.Fatalf("expected B/C skipped, got B=%s C=%s", results[1].Status, results[2].Status)
	}
	if calls["A"] != 1 || calls["B"] != 0 || calls["C"] != 0 {
		t.Fatalf("unexpected executor calls: %+v", calls)
	}
}

func TestRunnerBlocksDeployPathAfterDeployFailure(t *testing.T) {
	paths := state.NewPaths(t.TempDir())
	store := state.NewStore(paths)
	calls := map[string]int{}
	rt := &Runtime{
		Store: store,
		Plan: templates.Plan{
			TemplateID: "demo",
			Stages: []templates.StagePlan{
				{StageID: "A"},
				{StageID: "B"},
				{StageID: "C"},
				{StageID: "D"},
				{StageID: "E"},
				{StageID: "F"},
			},
		},
		Options: RunOptions{TemplateID: "demo", TemplateMode: "deploy"},
	}
	r := NewRunner([]StageExecutor{
		recordingStageExecutor{id: "A", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "B", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "C", status: schema.StatusFailed, calls: calls},
		recordingStageExecutor{id: "D", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "E", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "F", status: schema.StatusPassed, calls: calls},
	})

	results, err := r.Execute(context.Background(), rt)
	if err != nil {
		t.Fatalf("execute runner: %v", err)
	}
	if got, want := len(results), 6; got != want {
		t.Fatalf("unexpected result count: %d", got)
	}
	if results[2].Status != schema.StatusFailed {
		t.Fatalf("expected stage C failed, got %s", results[2].Status)
	}
	for _, idx := range []int{3, 4, 5} {
		if results[idx].Status != schema.StatusSkipped {
			t.Fatalf("expected stage %d skipped, got %s", idx, results[idx].Status)
		}
	}
	if calls["D"] != 0 || calls["E"] != 0 || calls["F"] != 0 {
		t.Fatalf("expected D/E/F not executed, calls=%+v", calls)
	}
}

func TestRunnerDoesNotBlockOperatePathForManageTemplate(t *testing.T) {
	paths := state.NewPaths(t.TempDir())
	store := state.NewStore(paths)
	calls := map[string]int{}
	rt := &Runtime{
		Store: store,
		Plan: templates.Plan{
			TemplateID: "demo",
			Stages: []templates.StagePlan{
				{StageID: "A"},
				{StageID: "B"},
				{StageID: "C"},
				{StageID: "D"},
			},
		},
		Options: RunOptions{TemplateID: "demo", TemplateMode: "manage"},
	}
	r := NewRunner([]StageExecutor{
		recordingStageExecutor{id: "A", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "B", status: schema.StatusPassed, calls: calls},
		recordingStageExecutor{id: "C", status: schema.StatusFailed, calls: calls},
		recordingStageExecutor{id: "D", status: schema.StatusPassed, calls: calls},
	})

	results, err := r.Execute(context.Background(), rt)
	if err != nil {
		t.Fatalf("execute runner: %v", err)
	}
	if got, want := len(results), 4; got != want {
		t.Fatalf("unexpected result count: %d", got)
	}
	if results[3].Status != schema.StatusPassed {
		t.Fatalf("expected stage D executed in manage mode, got %s", results[3].Status)
	}
	if calls["D"] != 1 {
		t.Fatalf("expected stage D executed once, calls=%+v", calls)
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

type recordingStageExecutor struct {
	id     string
	status schema.Status
	calls  map[string]int
}

func (r recordingStageExecutor) StageID() string { return r.id }

func (r recordingStageExecutor) Execute(_ context.Context, _ *Runtime, _ templates.StagePlan) (StageResult, error) {
	r.calls[r.id]++
	return StageResult{
		StageID: r.id,
		Status:  r.status,
	}, nil
}
