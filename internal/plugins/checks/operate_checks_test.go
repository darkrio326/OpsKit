package checks

import (
	"context"
	"errors"
	osexec "os/exec"
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
	if !hasMetric(res.Metrics, "check_degraded", "false") {
		t.Fatalf("expected check_degraded=false")
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
	if !hasMetric(res.Metrics, "check_degraded", "true") {
		t.Fatalf("expected check_degraded=true")
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
			return executil.Result{}, osexec.ErrNotFound
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "check_degraded_reason", "tool_unavailable") {
		t.Fatalf("expected tool_unavailable reason")
	}
}

func TestDNSResolveSkipNetworkQuery(t *testing.T) {
	plugin := &dnsResolveCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.dns_resolve",
		Params: map[string]any{
			"host":               "localhost",
			"skip_network_query": true,
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			if spec.Name == "nslookup" {
				t.Fatalf("nslookup should not run when skip_network_query=true")
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
	if !hasMetric(res.Metrics, "dns_skip_network_query", "true") {
		t.Fatalf("expected dns_skip_network_query=true")
	}
}

func TestLoadAverageUsesUptimeFallback(t *testing.T) {
	plugin := &loadAverageCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.load_average",
		Params: map[string]any{
			"loadavg_file": "/path/not-found",
			"max_load1":    2,
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			if spec.Name != "uptime" {
				t.Fatalf("unexpected command: %s", spec.Name)
			}
			return executil.Result{ExitCode: 0, Stdout: " 12:00  up 12 days,  2 users,  load averages: 1.24 0.92 0.70\n"}, nil
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "load_source", "uptime") {
		t.Fatalf("expected load_source=uptime")
	}
}

func TestLoadAverageUsesSysctlFallback(t *testing.T) {
	plugin := &loadAverageCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.load_average",
		Params: map[string]any{
			"loadavg_file": "/path/not-found",
			"max_load1":    3,
		},
		Exec: fakeRunner{run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
			switch spec.Name {
			case "uptime":
				return executil.Result{}, errors.New("uptime unavailable")
			case "sysctl":
				return executil.Result{ExitCode: 0, Stdout: "{ 2.11 1.80 1.70 }\n"}, nil
			default:
				t.Fatalf("unexpected command: %s", spec.Name)
				return executil.Result{}, nil
			}
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "load_source", "sysctl_vm_loadavg") {
		t.Fatalf("expected load_source=sysctl_vm_loadavg")
	}
}

func TestLoadAverageDegradedReasonMetric(t *testing.T) {
	plugin := &loadAverageCheck{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "d.load_average",
		Params: map[string]any{
			"loadavg_file": "/path/not-found",
		},
		Exec: fakeRunner{run: func(_ context.Context, _ executil.Spec) (executil.Result, error) {
			return executil.Result{}, osexec.ErrNotFound
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusWarn {
		t.Fatalf("expected warn, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "check_degraded", "true") {
		t.Fatalf("expected check_degraded=true")
	}
	if !hasMetric(res.Metrics, "check_degraded_reason", "tool_unavailable") {
		t.Fatalf("expected tool_unavailable reason")
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
	if !hasMetric(res.Metrics, "check_degraded_reason", "systemctl_query_failed") {
		t.Fatalf("expected systemctl_query_failed reason")
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
