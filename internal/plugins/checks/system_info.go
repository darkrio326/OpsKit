package checks

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type systemInfoCheck struct{}

func (c *systemInfoCheck) Kind() string { return "system_info" }

func (c *systemInfoCheck) Run(ctx context.Context, req Request) (Result, error) {
	uname, err := req.Exec.Run(ctx, executil.Spec{Name: "uname", Args: []string{"-sr"}, Timeout: 5 * time.Second})
	if err != nil {
		return Result{}, err
	}
	tz, err := req.Exec.Run(ctx, executil.Spec{Name: "date", Args: []string{"+%Z"}, Timeout: 5 * time.Second})
	if err != nil {
		return Result{}, err
	}
	msg := fmt.Sprintf("os=%s arch=%s kernel=%s timezone=%s", runtime.GOOS, runtime.GOARCH, strings.TrimSpace(uname.Stdout), strings.TrimSpace(tz.Stdout))
	return Result{
		CheckID:  req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  msg,
		Metrics: []schema.Metric{
			{Label: "OS", Value: runtime.GOOS},
			{Label: "Arch", Value: runtime.GOARCH},
		},
	}, nil
}
