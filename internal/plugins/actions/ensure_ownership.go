package actions

import (
	"context"
	"fmt"

	"opskit/internal/schema"
)

type ensureOwnershipAction struct{}

func (a *ensureOwnershipAction) Kind() string { return "ensure_ownership" }

func (a *ensureOwnershipAction) Run(ctx context.Context, req Request) (Result, error) {
	path := toString(req.Params["path"], "")
	user := toString(req.Params["user"], "")
	group := toString(req.Params["group"], "")
	recursive := toBool(req.Params["recursive"], false)
	if path == "" || user == "" || group == "" {
		return Result{}, fmt.Errorf("ensure_ownership requires params.path/user/group")
	}
	owner := user + ":" + group
	args := []string{}
	if recursive {
		args = append(args, "-R")
	}
	args = append(args, owner, path)
	res, err := runCmd(ctx, req, "chown", args...)
	if err != nil {
		return Result{}, err
	}
	if res.ExitCode != 0 {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "chown failed for " + path, Advice: res.Stderr}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("ownership ensured for %s", path)}, nil
}
