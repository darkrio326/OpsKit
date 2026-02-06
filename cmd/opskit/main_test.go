package main

import (
	"testing"

	"opskit/internal/core/exitcode"
	"opskit/internal/core/lock"
	"opskit/internal/schema"
	"opskit/internal/state"
)

func TestRunCLI_NoArgs(t *testing.T) {
	if got := runCLI([]string{}); got != exitcode.Precondition {
		t.Fatalf("expected precondition exit code, got %d", got)
	}
}

func TestRunCLI_UnknownCommand(t *testing.T) {
	if got := runCLI([]string{"nope"}); got != exitcode.Precondition {
		t.Fatalf("expected precondition exit code, got %d", got)
	}
}

func TestCmdStatus_ExitCodes(t *testing.T) {
	tmp := t.TempDir()
	store := state.NewStore(state.NewPaths(tmp))
	if err := store.InitStateIfMissing("demo"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	tests := []struct {
		name   string
		status schema.Status
		want   int
	}{
		{"passed", schema.StatusPassed, exitcode.Success},
		{"warn", schema.StatusWarn, exitcode.PartialSuccess},
		{"failed", schema.StatusFailed, exitcode.Failure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := state.DefaultLifecycle()
			for i := range lifecycle.Stages {
				lifecycle.Stages[i].Status = schema.StatusPassed
			}
			lifecycle.Stages[0].Status = tt.status
			if err := store.WriteLifecycle(lifecycle); err != nil {
				t.Fatalf("write lifecycle: %v", err)
			}

			got := cmdStatus([]string{"--output", tmp})
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestCmdStatus_LockConflict(t *testing.T) {
	tmp := t.TempDir()
	store := state.NewStore(state.NewPaths(tmp))
	if err := store.InitStateIfMissing("demo"); err != nil {
		t.Fatalf("init state: %v", err)
	}
	l, err := lock.Acquire(store.Paths().LockFile)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	defer func() { _ = l.Release() }()

	if got := cmdStatus([]string{"--output", tmp}); got != exitcode.ManualIntervention {
		t.Fatalf("expected manual intervention exit code, got %d", got)
	}
}
