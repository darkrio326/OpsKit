package handover

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"

	"opskit/internal/core/timeutil"
	"opskit/internal/reporting"
	"opskit/internal/schema"
	"opskit/internal/state"
)

type Result struct {
	ReportHTML schema.ArtifactRef
	ReportJSON schema.ArtifactRef
	Bundle     schema.ArtifactRef
}

type stageSummary struct {
	StageID     string        `json:"stageId"`
	Name        string        `json:"name"`
	Status      schema.Status `json:"status"`
	LastRunTime string        `json:"lastRunTime,omitempty"`
	IssueCount  int           `json:"issueCount"`
	ReportRef   string        `json:"reportRef,omitempty"`
}

type handoverSummary struct {
	GeneratedAt    string                 `json:"generatedAt"`
	OverallStatus  schema.OverallStatus   `json:"overallStatus"`
	OpenIssues     int                    `json:"openIssues"`
	ActiveTemplate []string               `json:"activeTemplates"`
	RecoverSummary *schema.RecoverSummary `json:"recoverSummary,omitempty"`
	Stages         []stageSummary         `json:"stages"`
	RecentReports  []schema.ArtifactRef   `json:"recentReports"`
	RecentBundles  []schema.ArtifactRef   `json:"recentBundles"`
}

func Generate(store *state.Store) (Result, error) {
	overall, err := store.ReadOverall()
	if err != nil {
		return Result{}, err
	}
	lifecycle, err := store.ReadLifecycle()
	if err != nil {
		return Result{}, err
	}
	artifacts, err := store.ReadArtifacts()
	if err != nil {
		return Result{}, err
	}

	now := timeutil.NowISO8601Compact()
	reportJSONName := fmt.Sprintf("handover-%s.json", now)
	reportHTMLName := fmt.Sprintf("handover-%s.html", now)
	bundleName := fmt.Sprintf("handover-%s.tar.gz", now)

	summary := handoverSummary{
		GeneratedAt:    timeutil.NowISO8601(),
		OverallStatus:  overall.OverallStatus,
		OpenIssues:     overall.OpenIssuesCount,
		ActiveTemplate: overall.ActiveTemplates,
		RecoverSummary: overall.RecoverSummary,
		Stages:         make([]stageSummary, 0, len(lifecycle.Stages)),
		RecentReports:  tailArtifacts(artifacts.Reports, 10),
		RecentBundles:  tailArtifacts(artifacts.Bundles, 10),
	}
	for _, s := range lifecycle.Stages {
		summary.Stages = append(summary.Stages, stageSummary{
			StageID:     s.StageID,
			Name:        s.Name,
			Status:      s.Status,
			LastRunTime: s.LastRunTime,
			IssueCount:  len(s.Issues),
			ReportRef:   s.ReportRef,
		})
	}

	reportJSONAbs := filepath.Join(store.Paths().ReportsDir, reportJSONName)
	if err := writeJSON(reportJSONAbs, summary); err != nil {
		return Result{}, err
	}
	reportHTMLAbs := filepath.Join(store.Paths().ReportsDir, reportHTMLName)
	if err := writeHTML(reportHTMLAbs, summary); err != nil {
		return Result{}, err
	}

	bundleAbs := filepath.Join(store.Paths().BundlesDir, bundleName)
	files := map[string]string{
		reportJSONAbs: filepath.Join("reports", reportJSONName),
		reportHTMLAbs: filepath.Join("reports", reportHTMLName),
		filepath.Join(store.Paths().StateDir, "overall.json"):   filepath.Join("state", "overall.json"),
		filepath.Join(store.Paths().StateDir, "lifecycle.json"): filepath.Join("state", "lifecycle.json"),
		filepath.Join(store.Paths().StateDir, "services.json"):  filepath.Join("state", "services.json"),
		filepath.Join(store.Paths().StateDir, "artifacts.json"): filepath.Join("state", "artifacts.json"),
	}
	if err := reporting.CreateTarGz(bundleAbs, files); err != nil {
		return Result{}, err
	}

	return Result{
		ReportHTML: schema.ArtifactRef{ID: "handover", Path: filepath.Join("reports", reportHTMLName)},
		ReportJSON: schema.ArtifactRef{ID: "handover-json", Path: filepath.Join("reports", reportJSONName)},
		Bundle:     schema.ArtifactRef{ID: "handover", Path: filepath.Join("bundles", bundleName)},
	}, nil
}

func tailArtifacts(in []schema.ArtifactRef, n int) []schema.ArtifactRef {
	if len(in) <= n {
		return append([]schema.ArtifactRef{}, in...)
	}
	return append([]schema.ArtifactRef{}, in[len(in)-n:]...)
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

func writeHTML(path string, summary handoverSummary) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	rows := make([]string, 0, len(summary.Stages))
	for _, s := range summary.Stages {
		rows = append(rows, fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%s</td></tr>",
			html.EscapeString(s.StageID), html.EscapeString(s.Name), html.EscapeString(string(s.Status)), s.IssueCount, html.EscapeString(s.LastRunTime)))
	}
	content := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>Handover Report</title>
<style>body{font-family:sans-serif;padding:16px}table{border-collapse:collapse;width:100%%}th,td{border:1px solid #ccc;padding:6px;text-align:left}</style>
</head><body>
<h1>Handover Report</h1>
<p>Generated: %s</p>
<p>Overall: <strong>%s</strong> | Open Issues: %d | Templates: %s</p>
<p>Recover: %s</p>
<table><thead><tr><th>Stage</th><th>Name</th><th>Status</th><th>Issues</th><th>Last Run</th></tr></thead><tbody>%s</tbody></table>
</body></html>`,
		html.EscapeString(summary.GeneratedAt),
		html.EscapeString(string(summary.OverallStatus)),
		summary.OpenIssues,
		html.EscapeString(strings.Join(summary.ActiveTemplate, ",")),
		html.EscapeString(recoverSummaryText(summary.RecoverSummary)),
		strings.Join(rows, ""),
	)
	return os.WriteFile(path, []byte(content), 0o644)
}

func recoverSummaryText(r *schema.RecoverSummary) string {
	if r == nil {
		return "-"
	}
	status := string(r.LastStatus)
	if strings.TrimSpace(status) == "" {
		status = "UNKNOWN"
	}
	trigger := r.LastTrigger
	if strings.TrimSpace(trigger) == "" {
		trigger = "-"
	}
	base := fmt.Sprintf("last=%s trigger=%s ok=%d fail=%d warn=%d", status, trigger, r.SuccessCount, r.FailureCount, r.WarnCount)
	if r.CircuitOpen {
		return base + " circuit=open until " + r.CooldownUntil
	}
	return base
}
