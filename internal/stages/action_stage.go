package stages

import (
	"context"
	"encoding/json"
	"fmt"

	"opskit/internal/engine"
	actionplugin "opskit/internal/plugins/actions"
	"opskit/internal/schema"
	"opskit/internal/templates"
)

func executeActionStage(ctx context.Context, rt *engine.Runtime, stageID string, reportTitle string, stage templates.StagePlan) (engine.StageResult, error) {
	if len(stage.Actions) == 0 {
		result := engine.StageResult{StageID: stageID, Status: schema.StatusSkipped, Metrics: []schema.Metric{{Label: "actions", Value: "0"}}}
		payload, _ := json.MarshalIndent(map[string]any{"stage": stageID, "status": result.Status, "message": "no actions configured"}, "", "  ")
		report := engine.ReportName(stageID)
		if err := rt.Store.WriteReportStub(report, reportTitle, string(payload)); err != nil {
			return engine.StageResult{}, err
		}
		result.Report = report
		return result, nil
	}

	result := engine.StageResult{StageID: stageID, Status: schema.StatusPassed}
	issues := []schema.Issue{}
	metrics := []schema.Metric{}
	actionRows := []map[string]any{}
	stepStatuses := []schema.Status{}

	for _, step := range stage.Actions {
		plugin, err := rt.ActionRegistry.MustPlugin(step.Kind)
		if err != nil {
			return engine.StageResult{}, err
		}
		res, err := plugin.Run(ctx, actionplugin.Request{ID: step.ID, Params: step.Params, Exec: rt.Exec})
		if err != nil {
			return engine.StageResult{}, err
		}
		stepStatuses = append(stepStatuses, res.Status)
		if len(res.Metrics) > 0 {
			metrics = append(metrics, res.Metrics...)
		}
		if res.Issue != nil {
			issues = append(issues, *res.Issue)
		}
		if len(res.Bundles) > 0 {
			result.Bundles = append(result.Bundles, res.Bundles...)
		}

		actionRows = append(actionRows, map[string]any{
			"actionId": res.ActionID,
			"status":   res.Status,
			"message":  res.Message,
		})

		if res.Status == schema.StatusFailed {
			result.Status = schema.StatusFailed
		} else if res.Status == schema.StatusWarn && result.Status != schema.StatusFailed {
			result.Status = schema.StatusWarn
		}
	}

	result.Metrics = append(metrics, schema.Metric{Label: "actions", Value: fmt.Sprintf("%d", len(stage.Actions))})
	result.Issues = issues
	result.StepStatuses = stepStatuses

	payload, _ := json.MarshalIndent(map[string]any{"stage": stageID, "status": result.Status, "actions": actionRows, "issues": issues}, "", "  ")
	report := engine.ReportName(stageID)
	if err := rt.Store.WriteReportStub(report, reportTitle, string(payload)); err != nil {
		return engine.StageResult{}, err
	}
	result.Report = report
	return result, nil
}
