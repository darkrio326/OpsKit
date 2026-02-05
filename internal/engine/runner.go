package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"opskit/internal/core/timeutil"
	"opskit/internal/schema"
	"opskit/internal/state"
	"opskit/internal/templates"
)

type Runner struct {
	executors map[string]StageExecutor
}

func NewRunner(execs []StageExecutor) *Runner {
	m := map[string]StageExecutor{}
	for _, e := range execs {
		m[e.StageID()] = e
	}
	return &Runner{executors: m}
}

func (r *Runner) Execute(ctx context.Context, rt *Runtime) ([]StageResult, error) {
	if err := rt.Store.InitStateIfMissing(rt.Options.TemplateID); err != nil {
		return nil, err
	}
	lifecycle, err := rt.Store.ReadLifecycle()
	if err != nil {
		return nil, err
	}
	services, err := rt.Store.ReadServices()
	if err != nil {
		return nil, err
	}
	artifacts, err := rt.Store.ReadArtifacts()
	if err != nil {
		return nil, err
	}

	results := make([]StageResult, 0, len(rt.Plan.Stages))
	for _, stagePlan := range rt.Plan.Stages {
		exec, ok := r.executors[stagePlan.StageID]
		if !ok {
			return nil, fmt.Errorf("stage executor not found: %s", stagePlan.StageID)
		}
		res, err := exec.Execute(ctx, rt, stagePlan)
		if err != nil {
			return nil, err
		}
		results = append(results, res)

		for i := range lifecycle.Stages {
			if lifecycle.Stages[i].StageID != res.StageID {
				continue
			}
			lifecycle.Stages[i].Status = res.Status
			lifecycle.Stages[i].LastRunTime = timeutil.NowISO8601()
			lifecycle.Stages[i].Metrics = res.Metrics
			lifecycle.Stages[i].Issues = res.Issues
			lifecycle.Stages[i].ReportRef = res.Report
		}

		if len(res.Checks) > 0 {
			service := schema.ServiceState{ServiceID: "host", Unit: "host", Health: serviceHealthFromStatus(res.Status), Checks: res.Checks}
			idx := slices.IndexFunc(services.Services, func(s schema.ServiceState) bool { return s.ServiceID == "host" })
			if idx >= 0 {
				services.Services[idx] = service
			} else {
				services.Services = append(services.Services, service)
			}
		}

		if res.Report != "" {
			rel := filepath.Join("reports", res.Report)
			artifacts.Reports = append(artifacts.Reports, schema.ArtifactRef{ID: stringsLowerStage(res.StageID), Path: rel})
		}
		if len(res.Bundles) > 0 {
			artifacts.Bundles = append(artifacts.Bundles, res.Bundles...)
		}
	}

	overallStatus, issues := state.DeriveOverall(lifecycle)
	overall := schema.OverallState{
		OverallStatus:   overallStatus,
		LastRefreshTime: timeutil.NowISO8601(),
		ActiveTemplates: []string{rt.Options.TemplateID},
		OpenIssuesCount: issues,
		RecoverSummary:  state.DeriveRecoverSummary(rt.Store.Paths(), lifecycle),
	}

	if err := rt.Store.WriteLifecycle(lifecycle); err != nil {
		return nil, err
	}
	if err := rt.Store.WriteServices(services); err != nil {
		return nil, err
	}
	if err := state.ApplyArtifactRetention(rt.Store.Paths(), &artifacts, state.DefaultMaxReports, state.DefaultMaxBundles); err != nil {
		return nil, err
	}
	if err := rt.Store.WriteArtifacts(artifacts); err != nil {
		return nil, err
	}
	if err := rt.Store.WriteOverall(overall); err != nil {
		return nil, err
	}

	return results, nil
}

func stringsLowerStage(stageID string) string {
	switch stageID {
	case "A":
		return "preflight"
	case "B":
		return "baseline"
	case "C":
		return "deploy"
	case "D":
		return "operate"
	case "E":
		return "recover"
	case "F":
		return "accept"
	default:
		return strings.ToLower(stageID)
	}
}

func serviceHealthFromStatus(s schema.Status) string {
	switch s {
	case schema.StatusPassed:
		return "healthy"
	case schema.StatusWarn:
		return "degraded"
	case schema.StatusFailed:
		return "unhealthy"
	default:
		return "unknown"
	}
}

func SelectStages(arg string) ([]string, error) {
	switch arg {
	case "A", "B", "C", "D", "E", "F":
		return []string{arg}, nil
	case "AF", "ALL", "all", "A-F":
		return []string{"A", "B", "C", "D", "E", "F"}, nil
	default:
		return nil, fmt.Errorf("invalid stage selector %q (use A..F or AF)", arg)
	}
}

func BuildPlan(t schema.Template, selected []string, includeDisabled bool) templates.Plan {
	return templates.BuildPlanWithOptions(t, selected, includeDisabled)
}

func ReportName(stageID string) string {
	return fmt.Sprintf("%s-%s.html", stringsLowerStage(stageID), time.Now().Format("20060102-150405"))
}
