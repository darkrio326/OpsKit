package stages

import (
	"context"
	"encoding/json"

	"opskit/internal/engine"
	checkplugin "opskit/internal/plugins/checks"
	"opskit/internal/schema"
	"opskit/internal/templates"
)

func executeCheckStage(ctx context.Context, rt *engine.Runtime, stageID string, reportTitle string, stage templates.StagePlan) (engine.StageResult, error) {
	result := engine.StageResult{StageID: stageID, Status: schema.StatusPassed}
	allMetrics := []schema.Metric{}
	issues := []schema.Issue{}
	checks := []schema.CheckState{}

	for _, step := range stage.Checks {
		plugin, err := rt.CheckRegistry.MustPlugin(step.Kind)
		if err != nil {
			return engine.StageResult{}, err
		}
		res, err := plugin.Run(ctx, checkplugin.Request{ID: step.ID, Params: step.Params, Exec: rt.Exec})
		if err != nil {
			return engine.StageResult{}, err
		}
		allMetrics = append(allMetrics, res.Metrics...)
		checks = append(checks, schema.CheckState{
			CheckID:  res.CheckID,
			Result:   toCheckResult(res.Status),
			Severity: res.Severity,
			Message:  res.Message,
		})
		if res.Issue != nil {
			issues = append(issues, *res.Issue)
		}

		if res.Status == schema.StatusFailed {
			result.Status = schema.StatusFailed
		} else if res.Status == schema.StatusWarn && result.Status != schema.StatusFailed {
			result.Status = schema.StatusWarn
		}
	}

	result.Metrics = allMetrics
	result.Issues = issues
	result.Checks = checks

	body, _ := json.MarshalIndent(map[string]any{"stage": stageID, "status": result.Status, "issues": issues, "checks": checks}, "", "  ")
	reportName := engine.ReportName(stageID)
	if err := rt.Store.WriteReportStub(reportName, reportTitle, string(body)); err != nil {
		return engine.StageResult{}, err
	}
	result.Report = reportName
	return result, nil
}

func toCheckResult(status schema.Status) string {
	switch status {
	case schema.StatusPassed:
		return "PASS"
	case schema.StatusWarn:
		return "WARN"
	case schema.StatusFailed:
		return "FAIL"
	default:
		return "UNKNOWN"
	}
}
