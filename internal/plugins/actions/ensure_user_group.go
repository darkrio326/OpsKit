package actions

import (
	"context"
	"fmt"

	"opskit/internal/schema"
)

type ensureUserGroupAction struct{}

func (a *ensureUserGroupAction) Kind() string { return "ensure_user_group" }

func (a *ensureUserGroupAction) Run(ctx context.Context, req Request) (Result, error) {
	user := toString(req.Params["user"], "")
	group := toString(req.Params["group"], "")
	if user == "" || group == "" {
		return Result{}, fmt.Errorf("ensure_user_group requires params.user and params.group")
	}

	grp, err := runCmd(ctx, req, "getent", "group", group)
	if err != nil {
		return Result{}, err
	}
	if grp.ExitCode != 0 {
		addGrp, addErr := runCmd(ctx, req, "groupadd", "--system", group)
		if addErr != nil {
			return Result{}, addErr
		}
		if addGrp.ExitCode != 0 {
			issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "failed to create group: " + group, Advice: addGrp.Stderr}
			return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
		}
	}

	uid, err := runCmd(ctx, req, "id", "-u", user)
	if err != nil {
		return Result{}, err
	}
	if uid.ExitCode != 0 {
		addUser, addErr := runCmd(ctx, req, "useradd", "--system", "-g", group, "-M", "-s", "/usr/sbin/nologin", user)
		if addErr != nil {
			return Result{}, addErr
		}
		if addUser.ExitCode != 0 {
			issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "failed to create user: " + user, Advice: addUser.Stderr}
			return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
		}
	}

	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("ensured user/group %s:%s", user, group)}, nil
}
