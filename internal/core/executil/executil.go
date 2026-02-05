package executil

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
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
}

type Runner interface {
	Run(ctx context.Context, spec Spec) (Result, error)
}

type SystemRunner struct{}

func (r SystemRunner) Run(ctx context.Context, spec Spec) (Result, error) {
	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

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
	}

	if err == nil {
		return res, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		res.ExitCode = exitErr.ExitCode()
		return res, nil
	}
	if errors.Is(cctx.Err(), context.DeadlineExceeded) {
		res.ExitCode = 124
		res.Stderr = "command timed out"
		return res, nil
	}
	return res, err
}
