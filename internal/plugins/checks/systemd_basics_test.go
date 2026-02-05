package checks

import (
	"context"
	"testing"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

func TestSystemdBasicsNoDiscoveredUnits(t *testing.T) {
	plugin := &systemdBasicsCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.systemd_basics",
		Params: map[string]any{
			"candidate_units": []any{"a.service", "b.service"},
			"require_any":     false,
		},
		Exec: fakeRunner{
			run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
				_ = spec
				return executil.Result{Stdout: "not-found\ninactive\n", ExitCode: 0}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
}

func TestSystemdBasicsInactiveUnitWarn(t *testing.T) {
	plugin := &systemdBasicsCheck{}
	call := 0
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.systemd_basics",
		Params: map[string]any{
			"candidate_units": []any{"good.service", "bad.service"},
			"severity":        "warn",
		},
		Exec: fakeRunner{
			run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
				call++
				switch call {
				case 1:
					return executil.Result{Stdout: "loaded\nactive\n", ExitCode: 0}, nil
				case 2:
					return executil.Result{Stdout: "loaded\ninactive\n", ExitCode: 0}, nil
				default:
					t.Fatalf("unexpected call: %v", spec.Args)
					return executil.Result{}, nil
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
	if res.Issue == nil {
		t.Fatalf("expected issue for inactive unit")
	}
}

type fakeRunner struct {
	run func(context.Context, executil.Spec) (executil.Result, error)
}

func (f fakeRunner) Run(ctx context.Context, spec executil.Spec) (executil.Result, error) {
	return f.run(ctx, spec)
}
