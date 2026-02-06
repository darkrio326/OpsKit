package recover

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCircuitOpenAndClose(t *testing.T) {
	p := filepath.Join(t.TempDir(), "recover_circuit.json")
	now := time.Now().UTC()
	if err := OpenWithTriggerCode(p, now, 10*time.Second, "boom", "readiness_failed", "onboot"); err != nil {
		t.Fatalf("open circuit: %v", err)
	}
	state, err := Load(p)
	if err != nil {
		t.Fatalf("load circuit: %v", err)
	}
	open, _ := IsOpen(state, now.Add(5*time.Second))
	if !open {
		t.Fatalf("expected circuit open")
	}
	if state.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", state.FailureCount)
	}
	if state.LastErrorCode != "readiness_failed" {
		t.Fatalf("expected last error code readiness_failed, got %q", state.LastErrorCode)
	}
	if err := Close(p); err != nil {
		t.Fatalf("close circuit: %v", err)
	}
	state, err = Load(p)
	if err != nil {
		t.Fatalf("reload circuit: %v", err)
	}
	open, _ = IsOpen(state, now.Add(5*time.Second))
	if open {
		t.Fatalf("expected circuit closed")
	}
	if state.SuccessCount != 1 {
		t.Fatalf("expected success count 1, got %d", state.SuccessCount)
	}
	if state.LastErrorCode != "" {
		t.Fatalf("expected last error code cleared, got %q", state.LastErrorCode)
	}
}

func TestCircuitWarn(t *testing.T) {
	p := filepath.Join(t.TempDir(), "recover_circuit.json")
	now := time.Now().UTC()
	if err := MarkWarnWithCode(p, now, "cooldown", "circuit_open", "timer"); err != nil {
		t.Fatalf("mark warn: %v", err)
	}
	state, err := Load(p)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if state.WarnCount != 1 {
		t.Fatalf("expected warn count 1, got %d", state.WarnCount)
	}
	if state.LastStatus != "WARN" {
		t.Fatalf("expected WARN status, got %s", state.LastStatus)
	}
	if state.LastErrorCode != "circuit_open" {
		t.Fatalf("expected WARN reason code circuit_open, got %q", state.LastErrorCode)
	}
}
