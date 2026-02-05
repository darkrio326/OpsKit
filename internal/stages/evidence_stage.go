package stages

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"opskit/internal/engine"
	evidenceplugin "opskit/internal/plugins/evidence"
	"opskit/internal/reporting"
	"opskit/internal/schema"
	"opskit/internal/templates"
)

func executeEvidenceStage(ctx context.Context, rt *engine.Runtime, stageID string, reportTitle string, stage templates.StagePlan) (engine.StageResult, error) {
	result := engine.StageResult{StageID: stageID, Status: schema.StatusPassed}
	issues := []schema.Issue{}
	metrics := []schema.Metric{}
	evidenceRows := []map[string]any{}
	bundleFiles := map[string]string{}

	for _, step := range stage.Evidence {
		plugin, err := rt.EvidenceRegistry.MustPlugin(step.Kind)
		if err != nil {
			return engine.StageResult{}, err
		}
		res, err := plugin.Collect(ctx, evidenceplugin.Request{ID: step.ID, Params: step.Params, Exec: rt.Exec})
		if err != nil {
			return engine.StageResult{}, err
		}
		if res.Issue != nil {
			issues = append(issues, *res.Issue)
		}
		if len(res.Metrics) > 0 {
			metrics = append(metrics, res.Metrics...)
		}
		if res.Path != "" {
			bundleFiles[res.Path] = filepath.Join("evidence", filepath.Base(res.Path))
		}

		evidenceRows = append(evidenceRows, map[string]any{
			"evidenceId": res.EvidenceID,
			"status":     res.Status,
			"message":    res.Message,
			"path":       res.Path,
		})

		if res.Status == schema.StatusFailed {
			result.Status = schema.StatusFailed
		} else if res.Status == schema.StatusWarn && result.Status != schema.StatusFailed {
			result.Status = schema.StatusWarn
		}
	}

	reportPayload, _ := json.MarshalIndent(map[string]any{
		"stage":    stageID,
		"status":   result.Status,
		"evidence": evidenceRows,
		"issues":   issues,
	}, "", "  ")
	reportName := engine.ReportName(stageID)
	if err := rt.Store.WriteReportStub(reportName, reportTitle, string(reportPayload)); err != nil {
		return engine.StageResult{}, err
	}
	reportAbs := filepath.Join(rt.Store.Paths().ReportsDir, reportName)
	bundleFiles[reportAbs] = filepath.Join("reports", reportName)

	snapshots := []string{}
	for _, stateName := range []string{"overall.json", "lifecycle.json", "services.json", "artifacts.json"} {
		abs := filepath.Join(rt.Store.Paths().StateDir, stateName)
		if fileExists(abs) {
			rel := filepath.Join("state", stateName)
			bundleFiles[abs] = rel
			snapshots = append(snapshots, rel)
		}
	}

	now := time.Now().Format("20060102-150405")
	summaryPath := filepath.Join(rt.Store.Paths().EvidenceDir, "acceptance-"+now+".json")
	summary := map[string]any{
		"stage":     stageID,
		"status":    result.Status,
		"generated": time.Now().Format(time.RFC3339),
		"evidence":  evidenceRows,
		"issues":    issues,
		"report":    filepath.Join("reports", reportName),
		"snapshots": snapshots,
	}
	if err := writeJSON(summaryPath, summary); err != nil {
		return engine.StageResult{}, err
	}
	bundleFiles[summaryPath] = filepath.Join("evidence", filepath.Base(summaryPath))

	bundleName := "acceptance-" + now + ".tar.gz"
	bundlePath := filepath.Join(rt.Store.Paths().BundlesDir, bundleName)
	if err := reporting.CreateTarGz(bundlePath, bundleFiles); err != nil {
		return engine.StageResult{}, err
	}
	result.Bundles = append(result.Bundles, schema.ArtifactRef{ID: "acceptance", Path: filepath.Join("bundles", bundleName)})

	result.Metrics = append(metrics,
		schema.Metric{Label: "evidence_items", Value: fmt.Sprintf("%d", len(evidenceRows))},
		schema.Metric{Label: "snapshots", Value: fmt.Sprintf("%d", len(snapshots))},
		schema.Metric{Label: "bundle", Value: bundleName},
	)
	result.Issues = issues
	result.Report = reportName
	return result, nil
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
