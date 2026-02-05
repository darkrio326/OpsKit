package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/executil"
	"opskit/internal/schema"
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

func toInt(v any, fallback int) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func toFileMode(v any, fallback os.FileMode) os.FileMode {
	if v == nil {
		return fallback
	}
	switch n := v.(type) {
	case int:
		return os.FileMode(n)
	case float64:
		return os.FileMode(int(n))
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return fallback
		}
		if parsed, err := strconv.ParseInt(s, 8, 32); err == nil {
			return os.FileMode(parsed)
		}
		if parsed, err := strconv.Atoi(s); err == nil {
			return os.FileMode(parsed)
		}
	}
	return fallback
}

func toBool(v any, fallback bool) bool {
	switch b := v.(type) {
	case bool:
		return b
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(b))
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if s, ok := v.([]string); ok {
		return s
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		s := strings.TrimSpace(fmt.Sprintf("%v", item))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func toIntSlice(v any) []int {
	if v == nil {
		return nil
	}
	if ints, ok := v.([]int); ok {
		return ints
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(arr))
	for _, item := range arr {
		out = append(out, toInt(item, 0))
	}
	return out
}

func ensureParent(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func writeJSON(path string, v any) error {
	if err := ensureParent(path); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func runCmd(ctx context.Context, req Request, name string, args ...string) (executil.Result, error) {
	if req.Exec == nil {
		return executil.Result{}, fmt.Errorf("exec runner not configured for action plugin")
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

func statusFromSeverity(sev schema.Severity) schema.Status {
	switch sev {
	case schema.SeverityWarn:
		return schema.StatusWarn
	case schema.SeverityInfo:
		return schema.StatusPassed
	default:
		return schema.StatusFailed
	}
}
