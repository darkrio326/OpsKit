package actions

import (
	"context"
	"fmt"

	"opskit/internal/core/fsx"
	"opskit/internal/schema"
)

type ensurePathsAction struct{}

func (a *ensurePathsAction) Kind() string { return "ensure_paths" }

func (a *ensurePathsAction) Run(_ context.Context, req Request) (Result, error) {
	paths := toStringSlice(req.Params["paths"])
	if len(paths) == 0 {
		for _, key := range []string{"install_root", "conf_dir", "data_dir", "logs_dir"} {
			if v, ok := req.Params[key]; ok {
				if p := toString(v, ""); p != "" {
					paths = append(paths, p)
				}
			}
		}
	}
	if len(paths) == 0 {
		return Result{}, fmt.Errorf("ensure_paths requires params.paths or install_root/conf_dir/data_dir/logs_dir")
	}

	perm := toFileMode(req.Params["perm"], 0o755)
	specs := make([]fsx.PathSpec, 0, len(paths))
	for _, p := range paths {
		specs = append(specs, fsx.PathSpec{Path: p, Perm: perm})
	}
	if err := fsx.EnsurePaths(specs); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "ensure_paths failed: " + err.Error(), Advice: "verify path permissions"}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}

	return Result{
		ActionID: req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  fmt.Sprintf("ensured %d paths", len(paths)),
		Metrics:  []schema.Metric{{Label: "ensured_paths", Value: fmt.Sprintf("%d", len(paths))}},
	}, nil
}
