package checks

import (
	"context"
	"errors"
	"testing"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

func TestDNSResolvePass(t *testing.T) {
	plugin := &dnsResolveCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.dns_resolve",
		Params: map[string]any{
			"host": "localhost",
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			if spec.Name != "getent" {
				t.Fatalf("unexpected command: %s", spec.Name)
			}
			return executil.Result{ExitCode: 0, Stdout: "127.0.0.1 localhost\n"}, nil
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "dns_records", "1") {
		t.Fatalf("expected dns_records metric")
	}
}

func TestNTPSyncPass(t *testing.T) {
	plugin := &ntpSyncCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.ntp_sync",
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			if spec.Name != "timedatectl" {
				t.Fatalf("unexpected command: %s", spec.Name)
			}
			return executil.Result{ExitCode: 0, Stdout: "yes\n"}, nil
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "ntp_sync", "synced") {
		t.Fatalf("expected ntp_sync metric")
	}
}

func TestNTPSyncFailWithSeverity(t *testing.T) {
	plugin := &ntpSyncCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.ntp_sync",
		Params: map[string]any{
			"severity": "fail",
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			switch spec.Name {
			case "timedatectl":
				return executil.Result{ExitCode: 0, Stdout: "no\n"}, nil
			case "chronyc":
				return executil.Result{ExitCode: 0, Stdout: "Leap status     : Not synchronised\n"}, nil
			default:
				t.Fatalf("unexpected command: %s", spec.Name)
				return executil.Result{}, nil
			}
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusFailed {
		t.Fatalf("expected failed, got %s", res.Status)
	}
	if res.Issue == nil {
		t.Fatalf("expected issue")
	}
}

func TestNTPSyncDegradedWhenToolingUnavailable(t *testing.T) {
	plugin := &ntpSyncCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.ntp_sync",
		Exec: fakeRunner{run: func(_ context.Context, _ executil.Spec) (executil.Result, error) {
			return executil.Result{}, errors.New("tool unavailable")
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
}

func TestDNSResolveFailWithSeverity(t *testing.T) {
	plugin := &dnsResolveCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.dns_resolve",
		Params: map[string]any{
			"host":     "example.invalid",
			"severity": "fail",
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			switch spec.Name {
			case "getent":
				return executil.Result{ExitCode: 2, Stderr: "not found"}, nil
			case "nslookup":
				return executil.Result{ExitCode: 1, Stderr: "server can't find"}, nil
			default:
				t.Fatalf("unexpected command: %s", spec.Name)
				return executil.Result{}, nil
			}
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusFailed {
		t.Fatalf("expected failed, got %s", res.Status)
	}
	if res.Issue == nil {
		t.Fatalf("expected issue")
	}
}

func TestDNSResolveDegradedWhenResolverUnavailable(t *testing.T) {
	plugin := &dnsResolveCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.dns_resolve",
		Exec: fakeRunner{run: func(_ context.Context, _ executil.Spec) (executil.Result, error) {
			return executil.Result{}, errors.New("command not found")
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
}

func TestSystemdRestartCountPass(t *testing.T) {
	plugin := &systemdRestartCountCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.systemd_restart_count",
		Params: map[string]any{
			"unit":         "demo.service",
			"max_restarts": 3,
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			if spec.Name != "systemctl" {
				t.Fatalf("unexpected command: %s", spec.Name)
			}
			return executil.Result{ExitCode: 0, Stdout: "2\n"}, nil
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "restart_count", "2") {
		t.Fatalf("expected restart_count metric")
	}
}

func TestSystemdRestartCountWarn(t *testing.T) {
	plugin := &systemdRestartCountCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.systemd_restart_count",
		Params: map[string]any{
			"unit":         "demo.service",
			"max_restarts": 1,
			"severity":     "warn",
		},
		Exec: fakeRunner{run: func(_ context.Context, _ executil.Spec) (executil.Result, error) {
			return executil.Result{ExitCode: 0, Stdout: "4\n"}, nil
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
	if res.Issue == nil {
		t.Fatalf("expected issue")
	}
}

func TestSystemdRestartCountDegraded(t *testing.T) {
	plugin := &systemdRestartCountCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.systemd_restart_count",
		Params: map[string]any{
			"unit": "demo.service",
		},
		Exec: fakeRunner{run: func(_ context.Context, _ executil.Spec) (executil.Result, error) {
			return executil.Result{}, errors.New("systemctl missing")
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
}

func hasMetric(metrics []schema.Metric, label, value string) bool {
	for _, m := range metrics {
		if m.Label == label && m.Value == value {
			return true
		}
	}
	return false
}
