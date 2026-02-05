package installer

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"opskit/internal/core/executil"
	"opskit/internal/state"
)

func TestInstallSystemdUnitsDryRun(t *testing.T) {
	root := t.TempDir()
	paths := state.NewPaths(root)
	res, err := InstallSystemdUnits(paths, SystemdOptions{
		SystemdDir: filepath.Join(root, "systemd"),
		BinaryPath: "/usr/local/bin/opskit",
		ListenAddr: ":18080",
	}, true)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	if len(res.Actions) != 5 {
		t.Fatalf("expected 5 systemd actions, got %d", len(res.Actions))
	}

	units := renderUnitFiles(paths, SystemdOptions{
		SystemdDir: filepath.Join(root, "systemd"),
		BinaryPath: "/usr/local/bin/opskit",
		ListenAddr: ":18080",
	})
	recoverService := units["opskit-recover.service"]
	if !strings.Contains(recoverService, "RECOVER_TRIGGER=onboot") {
		t.Fatalf("recover service should set RECOVER_TRIGGER=onboot")
	}
	recoverTimer := units["opskit-recover.timer"]
	if !strings.Contains(recoverTimer, "OnUnitActiveSec=10min") {
		t.Fatalf("recover timer should include periodic recovery interval")
	}
}

func TestActivateSystemdUnitsDryRun(t *testing.T) {
	res, err := ActivateSystemdUnits(context.Background(), nil, true)
	if err != nil {
		t.Fatalf("dry-run activate failed: %v", err)
	}
	if len(res.Actions) != 4 {
		t.Fatalf("expected 4 activation actions, got %d", len(res.Actions))
	}
	if !slices.Contains(res.Actions, "run systemctl enable --now opskit-recover.timer") {
		t.Fatalf("missing recover timer activation action")
	}
}

func TestActivateSystemdUnitsExecutesCommands(t *testing.T) {
	var calls []string
	runner := fakeRunner{
		run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			calls = append(calls, spec.Name+" "+strings.Join(spec.Args, " "))
			return executil.Result{ExitCode: 0}, nil
		},
	}

	_, err := ActivateSystemdUnits(context.Background(), runner, false)
	if err != nil {
		t.Fatalf("activate failed: %v", err)
	}
	want := []string{
		"systemctl daemon-reload",
		"systemctl enable --now opskit-web.service",
		"systemctl enable --now opskit-patrol.timer",
		"systemctl enable --now opskit-recover.timer",
	}
	if len(calls) != len(want) {
		t.Fatalf("unexpected call count: %d", len(calls))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("call[%d] expected %q got %q", i, want[i], calls[i])
		}
	}
}

type fakeRunner struct {
	run func(context.Context, executil.Spec) (executil.Result, error)
}

func (f fakeRunner) Run(ctx context.Context, spec executil.Spec) (executil.Result, error) {
	return f.run(ctx, spec)
}
