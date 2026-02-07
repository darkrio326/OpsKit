package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/exitcode"
	"opskit/internal/core/lock"
	"opskit/internal/engine"
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

			gotJSON := cmdStatus([]string{"--output", tmp, "--json"})
			if gotJSON != tt.want {
				t.Fatalf("expected %d with --json, got %d", tt.want, gotJSON)
			}
		})
	}
}

func TestCmdStatus_JSONOutputContract(t *testing.T) {
	tmp := t.TempDir()
	store := state.NewStore(state.NewPaths(tmp))
	if err := store.InitStateIfMissing("demo"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	lifecycle := state.DefaultLifecycle()
	for i := range lifecycle.Stages {
		lifecycle.Stages[i].Status = schema.StatusPassed
	}
	if err := store.WriteLifecycle(lifecycle); err != nil {
		t.Fatalf("write lifecycle: %v", err)
	}

	exit, stdout := captureStdout(t, func() int {
		return cmdStatus([]string{"--output", tmp, "--json"})
	})
	if exit != exitcode.Success {
		t.Fatalf("expected success, got %d", exit)
	}

	var payload statusJSONPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &payload); err != nil {
		t.Fatalf("json unmarshal failed: %v, output=%q", err, stdout)
	}
	if payload.SchemaVersion != statusJSONSchemaVersion {
		t.Fatalf("unexpected schemaVersion: %s", payload.SchemaVersion)
	}
	if payload.Command != statusJSONCommand {
		t.Fatalf("unexpected command: %s", payload.Command)
	}
	if payload.ExitCode != exit {
		t.Fatalf("unexpected exitCode: %d", payload.ExitCode)
	}
	if payload.Health != "ok" {
		t.Fatalf("unexpected health: %s", payload.Health)
	}
	if strings.TrimSpace(payload.GeneratedAt) == "" {
		t.Fatalf("generatedAt should not be empty")
	}
	if payload.Overall.LastRefreshTime != payload.GeneratedAt {
		t.Fatalf("generatedAt should match overall.lastRefreshTime, got %s vs %s", payload.GeneratedAt, payload.Overall.LastRefreshTime)
	}
	if len(payload.Lifecycle.Stages) != 6 {
		t.Fatalf("expected 6 lifecycle stages, got %d", len(payload.Lifecycle.Stages))
	}
}

func TestCmdStatus_JSONHealthByLifecycle(t *testing.T) {
	tmp := t.TempDir()
	store := state.NewStore(state.NewPaths(tmp))
	if err := store.InitStateIfMissing("demo"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	tests := []struct {
		name         string
		stageStatus  schema.Status
		wantExitCode int
		wantHealth   string
	}{
		{name: "ok", stageStatus: schema.StatusPassed, wantExitCode: exitcode.Success, wantHealth: "ok"},
		{name: "warn", stageStatus: schema.StatusWarn, wantExitCode: exitcode.PartialSuccess, wantHealth: "warn"},
		{name: "fail", stageStatus: schema.StatusFailed, wantExitCode: exitcode.Failure, wantHealth: "fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := state.DefaultLifecycle()
			for i := range lifecycle.Stages {
				lifecycle.Stages[i].Status = schema.StatusPassed
			}
			lifecycle.Stages[0].Status = tt.stageStatus
			if err := store.WriteLifecycle(lifecycle); err != nil {
				t.Fatalf("write lifecycle: %v", err)
			}

			exit, stdout := captureStdout(t, func() int {
				return cmdStatus([]string{"--output", tmp, "--json"})
			})
			if exit != tt.wantExitCode {
				t.Fatalf("expected exit=%d, got %d", tt.wantExitCode, exit)
			}

			var payload statusJSONPayload
			if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &payload); err != nil {
				t.Fatalf("json unmarshal failed: %v, output=%q", err, stdout)
			}
			if payload.ExitCode != tt.wantExitCode {
				t.Fatalf("expected payload exitCode=%d, got %d", tt.wantExitCode, payload.ExitCode)
			}
			if payload.Health != tt.wantHealth {
				t.Fatalf("expected health=%s, got %s", tt.wantHealth, payload.Health)
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

func TestCmdRun_DryRunSuccess(t *testing.T) {
	tmp := t.TempDir()
	got := cmdRun([]string{"A", "--template", "generic-manage-v1", "--dry-run", "--output", tmp})
	if got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
}

func TestCmdRun_InvalidStage(t *testing.T) {
	got := cmdRun([]string{"Z"})
	if got != exitcode.Precondition {
		t.Fatalf("expected precondition exit code, got %d", got)
	}
}

func TestCmdAccept_DryRunSuccess(t *testing.T) {
	tmp := t.TempDir()
	got := cmdAccept([]string{"--template", "generic-manage-v1", "--dry-run", "--output", tmp})
	if got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
}

func TestCmdTemplate_InvalidFile(t *testing.T) {
	got := cmdTemplate([]string{"validate", "/no/such/template.json"})
	if got != exitcode.Precondition {
		t.Fatalf("expected precondition exit code, got %d", got)
	}
}

func TestCmdTemplate_VarsFile(t *testing.T) {
	tmp := t.TempDir()
	tplPath := filepath.Join(tmp, "t.json")
	varsPath := filepath.Join(tmp, "vars.json")

	tpl := `{
  "id": "t",
  "name": "t",
  "mode": "manage",
  "vars": {
    "ENV": { "type": "string", "required": true },
    "PORTS": { "type": "array", "required": true }
  },
  "stages": {
    "A": {
      "checks": [
        { "id": "a.system_info", "kind": "system_info", "params": { "ports": "${PORTS}", "env": "${ENV}" } }
      ]
    }
  }
}`
	vars := `{"ENV":"dev","PORTS":[80,443]}`
	if err := os.WriteFile(tplPath, []byte(tpl), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	if err := os.WriteFile(varsPath, []byte(vars), 0o644); err != nil {
		t.Fatalf("write vars: %v", err)
	}

	got := cmdTemplate([]string{"validate", "--vars-file", varsPath, tplPath})
	if got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
}

func TestCmdInstall_DryRun(t *testing.T) {
	tmp := t.TempDir()
	got := cmdInstall([]string{"--template", "generic-manage-v1", "--dry-run", "--no-systemd", "--output", tmp})
	if got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
}

func TestMapErrorToExit(t *testing.T) {
	if got := mapErrorToExit(coreerr.ErrPreconditionFailed); got != exitcode.Precondition {
		t.Fatalf("expected precondition, got %d", got)
	}
	if got := mapErrorToExit(coreerr.ErrPartialSuccess); got != exitcode.PartialSuccess {
		t.Fatalf("expected partial success, got %d", got)
	}
	if got := mapErrorToExit(coreerr.ErrLocked); got != exitcode.ManualIntervention {
		t.Fatalf("expected manual intervention, got %d", got)
	}
	if got := mapErrorToExit(errDummy{}); got != exitcode.Failure {
		t.Fatalf("expected failure, got %d", got)
	}
}

func TestExitForLifecycle(t *testing.T) {
	lifecycle := state.DefaultLifecycle()
	for i := range lifecycle.Stages {
		lifecycle.Stages[i].Status = schema.StatusPassed
	}
	if got := exitForLifecycle(lifecycle); got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
	lifecycle.Stages[0].Status = schema.StatusWarn
	if got := exitForLifecycle(lifecycle); got != exitcode.PartialSuccess {
		t.Fatalf("expected partial success, got %d", got)
	}
	lifecycle.Stages[1].Status = schema.StatusFailed
	if got := exitForLifecycle(lifecycle); got != exitcode.Failure {
		t.Fatalf("expected failure, got %d", got)
	}
}

func TestStageResultsExit(t *testing.T) {
	results := []engine.StageResult{
		{StageID: "A", Status: schema.StatusPassed},
	}
	if got := stageResultsExit(results); got != exitcode.Success {
		t.Fatalf("expected success, got %d", got)
	}
	results[0].Status = schema.StatusWarn
	if got := stageResultsExit(results); got != exitcode.PartialSuccess {
		t.Fatalf("expected partial success, got %d", got)
	}
	results = append(results, engine.StageResult{StageID: "B", Status: schema.StatusFailed})
	if got := stageResultsExit(results); got != exitcode.Failure {
		t.Fatalf("expected failure, got %d", got)
	}
}

type errDummy struct{}

func (errDummy) Error() string { return "dummy" }

func captureStdout(t *testing.T, fn func() int) (int, string) {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w

	exit := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = oldStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = r.Close()
	return exit, string(out)
}
