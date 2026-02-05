package state

import (
	"path/filepath"
	"testing"
	"time"

	"opskit/internal/recover"
	"opskit/internal/schema"
)

func TestDeriveRecoverSummaryEmpty(t *testing.T) {
	paths := NewPaths(t.TempDir())
	lifecycle := DefaultLifecycle()
	if s := DeriveRecoverSummary(paths, lifecycle); s != nil {
		t.Fatalf("expected nil summary when no recover data")
	}
}

func TestDeriveRecoverSummaryFromLifecycleAndCircuit(t *testing.T) {
	root := t.TempDir()
	paths := NewPaths(root)
	if err := NewStore(paths).EnsureLayout(); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	now := time.Now().UTC()
	circuit := filepath.Join(paths.StateDir, "recover_circuit.json")
	if err := recover.OpenWithTrigger(circuit, now, 5*time.Minute, "boom", "onboot"); err != nil {
		t.Fatalf("open circuit: %v", err)
	}

	lifecycle := DefaultLifecycle()
	for i := range lifecycle.Stages {
		if lifecycle.Stages[i].StageID == "E" {
			lifecycle.Stages[i].Status = schema.StatusFailed
			lifecycle.Stages[i].LastRunTime = now.Format(time.RFC3339)
			lifecycle.Stages[i].Metrics = []schema.Metric{{Label: "recover_trigger", Value: "timer"}}
		}
	}
	s := DeriveRecoverSummary(paths, lifecycle)
	if s == nil {
		t.Fatalf("expected recover summary")
	}
	if s.LastTrigger != "timer" {
		t.Fatalf("expected trigger timer, got %q", s.LastTrigger)
	}
	if s.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", s.FailureCount)
	}
	if !s.CircuitOpen {
		t.Fatalf("expected circuit open")
	}
}
