package executil

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestSystemRunnerSuccess(t *testing.T) {
	r := SystemRunner{}
	res, err := r.Run(context.Background(), Spec{
		Name:    "sh",
		Args:    []string{"-c", "printf ok"},
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
	if strings.TrimSpace(res.Stdout) != "ok" {
		t.Fatalf("unexpected stdout: %q", res.Stdout)
	}
	if res.TimedOut {
		t.Fatalf("expected timedOut=false")
	}
	if res.Duration <= 0 {
		t.Fatalf("expected positive duration")
	}
}

func TestSystemRunnerNonZeroExit(t *testing.T) {
	r := SystemRunner{}
	res, err := r.Run(context.Background(), Spec{
		Name:    "sh",
		Args:    []string{"-c", "exit 7"},
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", res.ExitCode)
	}
	if res.TimedOut {
		t.Fatalf("expected timedOut=false")
	}
}

func TestSystemRunnerTimeout(t *testing.T) {
	r := SystemRunner{}
	res, err := r.Run(context.Background(), Spec{
		Name:    "sh",
		Args:    []string{"-c", "sleep 1"},
		Timeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 124 {
		t.Fatalf("expected exit code 124, got %d", res.ExitCode)
	}
	if !res.TimedOut {
		t.Fatalf("expected timedOut=true")
	}
	if !strings.Contains(res.Stderr, "command timed out") {
		t.Fatalf("unexpected stderr: %q", res.Stderr)
	}
}

func TestSystemRunnerCanceledContext(t *testing.T) {
	r := SystemRunner{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res, err := r.Run(ctx, Spec{
		Name:    "sh",
		Args:    []string{"-c", "sleep 1"},
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 130 {
		t.Fatalf("expected exit code 130, got %d", res.ExitCode)
	}
	if strings.TrimSpace(res.Stderr) == "" {
		t.Fatalf("expected non-empty stderr")
	}
}

func TestSystemRunnerEmptyCommand(t *testing.T) {
	r := SystemRunner{}
	_, err := r.Run(context.Background(), Spec{Name: ""})
	if err == nil {
		t.Fatalf("expected error for empty command")
	}
}

func TestSystemRunnerNilContext(t *testing.T) {
	r := SystemRunner{}
	res, err := r.Run(nil, Spec{
		Name:    "sh",
		Args:    []string{"-c", "printf ok"},
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
}
