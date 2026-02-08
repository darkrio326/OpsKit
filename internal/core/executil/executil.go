package executil

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Spec struct {
	Name    string
	Args    []string
	Timeout time.Duration
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	TimedOut bool
}

type Runner interface {
	Run(ctx context.Context, spec Spec) (Result, error)
}

type SystemRunner struct{}

func (r SystemRunner) Run(ctx context.Context, spec Spec) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(spec.Name) == "" {
		return Result{}, fmt.Errorf("command name is required")
	}

	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	start := time.Now()
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, spec.Name, spec.Args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	res := Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Duration: time.Since(start),
	}

	if err == nil {
		return res, nil
	}

	if errors.Is(cctx.Err(), context.DeadlineExceeded) {
		res.ExitCode = 124
		res.TimedOut = true
		if strings.TrimSpace(res.Stderr) == "" {
			res.Stderr = "command timed out"
		} else {
			res.Stderr = strings.TrimSpace(res.Stderr) + "\ncommand timed out"
		}
		return res, nil
	}
	if errors.Is(cctx.Err(), context.Canceled) {
		res.ExitCode = 130
		if strings.TrimSpace(res.Stderr) == "" {
			res.Stderr = "command canceled"
		}
		return res, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	return res, err
}
