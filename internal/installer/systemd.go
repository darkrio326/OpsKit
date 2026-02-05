package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/executil"
	"opskit/internal/core/fsx"
	"opskit/internal/state"
)

type SystemdOptions struct {
	SystemdDir string
	BinaryPath string
	ListenAddr string
}

type SystemdResult struct {
	Actions []string
	Units   []string
}

type SystemdActivationResult struct {
	Actions []string
}

func InstallSystemdUnits(paths state.Paths, opt SystemdOptions, dryRun bool) (SystemdResult, error) {
	if opt.ListenAddr == "" {
		opt.ListenAddr = ":18080"
	}
	if opt.SystemdDir == "" {
		return SystemdResult{}, fmt.Errorf("systemd dir is required")
	}
	if opt.BinaryPath == "" {
		return SystemdResult{}, fmt.Errorf("binary path is required")
	}

	units := renderUnitFiles(paths, opt)
	actions := make([]string, 0, len(units))
	unitPaths := make([]string, 0, len(units))
	for name := range units {
		actions = append(actions, "write "+filepath.Join(opt.SystemdDir, name))
		unitPaths = append(unitPaths, filepath.Join(opt.SystemdDir, name))
	}

	if dryRun {
		return SystemdResult{Actions: actions, Units: unitPaths}, nil
	}

	if err := fsx.EnsureDir(opt.SystemdDir, 0o755); err != nil {
		return SystemdResult{}, err
	}
	for name, content := range units {
		path := filepath.Join(opt.SystemdDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return SystemdResult{}, fmt.Errorf("write systemd unit %s: %w", path, err)
		}
	}
	return SystemdResult{Actions: actions, Units: unitPaths}, nil
}

func renderUnitFiles(paths state.Paths, opt SystemdOptions) map[string]string {
	webService := fmt.Sprintf(`[Unit]
Description=OpsKit Static Web
After=network.target

[Service]
Type=simple
ExecStart=%s web --output %s --listen %s
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
`, opt.BinaryPath, paths.Root, opt.ListenAddr)

	patrolService := fmt.Sprintf(`[Unit]
Description=OpsKit Operate Patrol
After=network.target

[Service]
Type=oneshot
ExecStart=%s run D --output %s
`, opt.BinaryPath, paths.Root)

	patrolTimer := `[Unit]
Description=OpsKit Operate Patrol Timer

[Timer]
OnBootSec=2min
OnUnitActiveSec=10min
Unit=opskit-patrol.service

[Install]
WantedBy=timers.target
`

	recoverService := fmt.Sprintf(`[Unit]
Description=OpsKit Recovery Trigger
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=%s run E --output %s --vars RECOVER_TRIGGER=onboot
`, opt.BinaryPath, paths.Root)

	recoverTimer := `[Unit]
Description=OpsKit Recovery On-Boot Timer

[Timer]
OnBootSec=2min
OnUnitActiveSec=10min
Unit=opskit-recover.service
Persistent=true

[Install]
WantedBy=timers.target
`

	return map[string]string{
		"opskit-web.service":     webService,
		"opskit-patrol.service":  patrolService,
		"opskit-patrol.timer":    patrolTimer,
		"opskit-recover.service": recoverService,
		"opskit-recover.timer":   recoverTimer,
	}
}

func ActivationActions() []string {
	return []string{
		"run systemctl daemon-reload",
		"run systemctl enable --now opskit-web.service",
		"run systemctl enable --now opskit-patrol.timer",
		"run systemctl enable --now opskit-recover.timer",
	}
}

func ActivateSystemdUnits(ctx context.Context, runner executil.Runner, dryRun bool) (SystemdActivationResult, error) {
	actions := ActivationActions()
	if dryRun {
		return SystemdActivationResult{Actions: actions}, nil
	}
	if runner == nil {
		return SystemdActivationResult{}, fmt.Errorf("exec runner is required for systemd activation")
	}

	commands := [][]string{
		{"daemon-reload"},
		{"enable", "--now", "opskit-web.service"},
		{"enable", "--now", "opskit-patrol.timer"},
		{"enable", "--now", "opskit-recover.timer"},
	}
	for _, args := range commands {
		res, err := runner.Run(ctx, executil.Spec{
			Name:    "systemctl",
			Args:    args,
			Timeout: 30 * time.Second,
		})
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				return SystemdActivationResult{}, fmt.Errorf("%w: required command not found: systemctl", coreerr.ErrPreconditionFailed)
			}
			return SystemdActivationResult{}, err
		}
		if res.ExitCode != 0 {
			msg := res.Stderr
			if msg == "" {
				msg = res.Stdout
			}
			return SystemdActivationResult{}, fmt.Errorf("systemctl %v failed: %s", args, msg)
		}
	}
	return SystemdActivationResult{Actions: actions}, nil
}
