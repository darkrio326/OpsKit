package evidence

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/executil"
)

func toString(v any, fallback string) string {
	if v == nil {
		return fallback
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return fallback
	}
	return s
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if xs, ok := v.([]string); ok {
		return xs
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s := toString(item, "")
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func toInt(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return parsed
		}
	}
	return fallback
}

func ensureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func runCmd(ctx context.Context, req Request, name string, args ...string) (executil.Result, error) {
	if req.Exec == nil {
		return executil.Result{}, fmt.Errorf("exec runner not configured for evidence plugin")
	}
	res, err := req.Exec.Run(ctx, executil.Spec{Name: name, Args: args})
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return executil.Result{}, fmt.Errorf("%w: required command not found: %s", coreerr.ErrPreconditionFailed, name)
		}
		return executil.Result{}, err
	}
	return res, nil
}
