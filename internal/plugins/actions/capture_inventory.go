package actions

import (
	"context"
	"fmt"
	"os"
	"strings"

	"opskit/internal/schema"
)

type captureInventoryAction struct{}

func (a *captureInventoryAction) Kind() string { return "capture_inventory" }

func (a *captureInventoryAction) Run(ctx context.Context, req Request) (Result, error) {
	units := toStringSlice(req.Params["units"])
	ports := toIntSlice(req.Params["ports"])
	paths := toStringSlice(req.Params["paths"])
	output := toString(req.Params["output"], "")
	if output == "" {
		return Result{}, fmt.Errorf("capture_inventory requires params.output")
	}

	inv := map[string]any{
		"units": map[string]string{},
		"ports": map[string]bool{},
		"paths": map[string]bool{},
	}

	unitStates := inv["units"].(map[string]string)
	for _, unit := range units {
		res, err := runCmd(ctx, req, "systemctl", "is-active", unit)
		if err != nil {
			unitStates[unit] = "unknown"
			continue
		}
		state := strings.TrimSpace(res.Stdout)
		if state == "" {
			state = "unknown"
		}
		unitStates[unit] = state
	}

	portStates := inv["ports"].(map[string]bool)
	if len(ports) > 0 {
		res, err := runCmd(ctx, req, "ss", "-ltnH")
		if err == nil {
			listening := map[string]bool{}
			for _, line := range strings.Split(strings.TrimSpace(res.Stdout), "\n") {
				fields := strings.Fields(strings.TrimSpace(line))
				if len(fields) < 4 {
					continue
				}
				local := fields[3]
				idx := strings.LastIndex(local, ":")
				if idx < 0 || idx+1 >= len(local) {
					continue
				}
				listening[local[idx+1:]] = true
			}
			for _, p := range ports {
				portStates[fmt.Sprintf("%d", p)] = listening[fmt.Sprintf("%d", p)]
			}
		}
	}

	pathStates := inv["paths"].(map[string]bool)
	for _, p := range paths {
		_, err := os.Stat(p)
		pathStates[p] = err == nil
	}

	if err := writeJSON(output, inv); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "write inventory failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}

	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "inventory captured: " + output}, nil
}
