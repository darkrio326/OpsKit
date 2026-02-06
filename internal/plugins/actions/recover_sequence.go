package actions

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/redaction"
	"opskit/internal/recover"
	"opskit/internal/reporting"
	"opskit/internal/schema"
)

type recoverSequenceAction struct{}

func (a *recoverSequenceAction) Kind() string { return "recover_sequence" }

func (a *recoverSequenceAction) Run(ctx context.Context, req Request) (Result, error) {
	units := toStringSlice(req.Params["units"])
	allowEmptyUnits := toBool(req.Params["allow_empty_units"], false)
	if len(units) == 0 && !allowEmptyUnits {
		return Result{ActionID: req.ID, Status: schema.StatusSkipped, Severity: schema.SeverityInfo, Message: "no units configured for recovery"}, nil
	}

	maxAttempts := toInt(req.Params["max_attempts"], 1)
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	cooldownSeconds := toInt(req.Params["cooldown_seconds"], 600)
	if cooldownSeconds < 1 {
		cooldownSeconds = 600
	}
	requireNetwork := toBool(req.Params["readiness_network"], true)
	requiredMounts := toStringSlice(req.Params["readiness_mounts"])
	if len(requiredMounts) == 0 {
		requiredMounts = []string{"/"}
	}
	circuitFile := toString(req.Params["circuit_file"], "/var/lib/opskit/state/recover_circuit.json")
	collectFile := toString(req.Params["collect_file"], "")
	collectBundleDir := toString(req.Params["collect_bundle_dir"], "")
	collectOutputLimit := toInt(req.Params["collect_output_limit"], 16384)
	if collectOutputLimit < 256 {
		collectOutputLimit = 256
	}
	stageID := strings.ToUpper(toString(req.Params["stage"], "E"))
	triggerSource := toString(req.Params["trigger_source"], "manual")
	redactKeys := toStringSlice(req.Params["redact_keys"])
	collectPaths := toStringSlice(req.Params["collect_paths"])

	now := time.Now()
	circuit, err := recover.Load(circuitFile)
	if err != nil {
		return Result{}, err
	}
	open, until := recover.IsOpen(circuit, now)
	if open {
		msg := fmt.Sprintf("recovery circuit open until %s", until.Format(time.RFC3339))
		_ = recover.MarkWarn(circuitFile, now, msg, triggerSource)
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityWarn, Message: msg, Advice: "wait for cooldown or manual intervention"}
		return Result{
			ActionID: req.ID,
			Status:   schema.StatusWarn,
			Severity: schema.SeverityWarn,
			Message:  msg,
			Issue:    issue,
			Metrics:  []schema.Metric{{Label: "recover_trigger", Value: triggerSource}},
		}, nil
	}

	if err := checkReadiness(ctx, req, requireNetwork, requiredMounts); err != nil {
		_ = recover.OpenWithTrigger(circuitFile, now, time.Duration(cooldownSeconds)*time.Second, err.Error(), triggerSource)
		bundles := []schema.ArtifactRef{}
		if collectFile != "" {
			if b, collectErr := writeRecoverCollect(ctx, req, collectFile, units, err.Error(), circuitFile, stageID, collectBundleDir, triggerSource, collectPaths, redactKeys, collectOutputLimit); collectErr == nil && b.Path != "" {
				bundles = append(bundles, b)
			}
		}
		msg := "readiness check failed: " + err.Error()
		advice := "fix readiness issues then retry"
		if errors.Is(err, coreerr.ErrPreconditionFailed) {
			advice = "install required system command or adjust recover readiness checks"
		}
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: msg, Advice: advice}
		return Result{
			ActionID: req.ID,
			Status:   schema.StatusFailed,
			Severity: schema.SeverityFail,
			Message:  issue.Message,
			Issue:    issue,
			Bundles:  bundles,
			Metrics: []schema.Metric{
				{Label: "recover_trigger", Value: triggerSource},
				{Label: "recover_precondition", Value: boolMetric(errors.Is(err, coreerr.ErrPreconditionFailed))},
			},
		}, nil
	}

	lastErr := ""
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		allActive := true
		for _, unit := range units {
			startRes, runErr := runCmd(ctx, req, "systemctl", "start", unit)
			if runErr != nil {
				lastErr = runErr.Error()
				allActive = false
				break
			}
			if startRes.ExitCode != 0 {
				lastErr = strings.TrimSpace(startRes.Stderr)
				if lastErr == "" {
					lastErr = fmt.Sprintf("failed to start %s", unit)
				}
				allActive = false
				break
			}
			activeRes, activeErr := runCmd(ctx, req, "systemctl", "is-active", unit)
			if activeErr != nil {
				lastErr = activeErr.Error()
				allActive = false
				break
			}
			if strings.TrimSpace(activeRes.Stdout) != "active" {
				lastErr = fmt.Sprintf("unit not active after start: %s", unit)
				allActive = false
				break
			}
		}
		if allActive {
			_ = recover.CloseWithTrigger(circuitFile, now, triggerSource)
			recoveredUnits := len(units)
			return Result{
				ActionID: req.ID,
				Status:   schema.StatusPassed,
				Severity: schema.SeverityInfo,
				Message:  fmt.Sprintf("recovery checks passed (%d units)", recoveredUnits),
				Metrics: []schema.Metric{
					{Label: "recover_trigger", Value: triggerSource},
					{Label: "recover_attempts", Value: fmt.Sprintf("%d", attempt)},
					{Label: "recovered_units", Value: fmt.Sprintf("%d", recoveredUnits)},
				},
			}, nil
		}
	}

	if lastErr == "" {
		lastErr = "unknown recovery failure"
	}
	_ = recover.OpenWithTrigger(circuitFile, now, time.Duration(cooldownSeconds)*time.Second, lastErr, triggerSource)
	bundles := []schema.ArtifactRef{}
	if collectFile != "" {
		if b, collectErr := writeRecoverCollect(ctx, req, collectFile, units, lastErr, circuitFile, stageID, collectBundleDir, triggerSource, collectPaths, redactKeys, collectOutputLimit); collectErr == nil && b.Path != "" {
			bundles = append(bundles, b)
		}
	}
	issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "recovery failed: " + lastErr, Advice: "inspect collect bundle and systemd logs"}
	return Result{
		ActionID: req.ID,
		Status:   schema.StatusFailed,
		Severity: schema.SeverityFail,
		Message:  issue.Message,
		Issue:    issue,
		Metrics: []schema.Metric{
			{Label: "recover_trigger", Value: triggerSource},
			{Label: "recover_attempts", Value: fmt.Sprintf("%d", maxAttempts)},
		},
		Bundles: bundles,
	}, nil
}

func checkReadiness(ctx context.Context, req Request, requireNetwork bool, requiredMounts []string) error {
	if requireNetwork {
		res, err := runCmd(ctx, req, "ip", "route", "show", "default")
		if err != nil {
			return err
		}
		if res.ExitCode != 0 || strings.TrimSpace(res.Stdout) == "" {
			return fmt.Errorf("network default route not ready")
		}
	}
	mounted := map[string]bool{}
	f, err := os.Open("/proc/mounts")
	if err == nil {
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			parts := strings.Fields(s.Text())
			if len(parts) >= 2 {
				mounted[parts[1]] = true
			}
		}
	}
	for _, m := range requiredMounts {
		if len(mounted) > 0 {
			if !mounted[m] {
				return fmt.Errorf("required mount not ready: %s", m)
			}
			continue
		}
		if _, statErr := os.Stat(m); statErr != nil {
			return fmt.Errorf("required mount not accessible: %s", m)
		}
	}
	return nil
}

func writeRecoverCollect(ctx context.Context, req Request, path string, units []string, reason string, circuitFile string, stageID string, collectBundleDir string, triggerSource string, collectPaths []string, redactKeys []string, collectOutputLimit int) (schema.ArtifactRef, error) {
	payload := map[string]any{
		"timestamp":     time.Now().Format(time.RFC3339),
		"triggerSource": triggerSource,
		"reason":        redaction.RedactText(reason, redactKeys...),
		"units":         units,
		"commands":      map[string]string{},
		"journals":      map[string]string{},
		"paths":         summarizePaths(collectPaths),
		"limits": map[string]any{
			"maxChars": collectOutputLimit,
		},
	}
	cmds := payload["commands"].(map[string]string)
	for _, spec := range []struct {
		name string
		args []string
		key  string
	}{
		{name: "ss", args: []string{"-ltn"}, key: "ss_ltn"},
		{name: "df", args: []string{"-h"}, key: "df_h"},
		{name: "free", args: []string{"-m"}, key: "free_m"},
	} {
		res, err := runCmd(ctx, req, spec.name, spec.args...)
		if err != nil {
			cmds[spec.key] = limitText("error: "+err.Error(), collectOutputLimit)
			continue
		}
		out := formatCommandResult(spec.name, res.ExitCode, res.Stdout, res.Stderr)
		cmds[spec.key] = limitText(redaction.RedactText(out, redactKeys...), collectOutputLimit)
	}
	journals := payload["journals"].(map[string]string)
	for _, unit := range units {
		key := "journal_" + strings.ReplaceAll(strings.ReplaceAll(unit, ".", "_"), "-", "_")
		res, err := runCmd(ctx, req, "journalctl", "-u", unit, "--no-pager", "-n", "200")
		if err != nil {
			journals[key] = limitText("error: "+err.Error(), collectOutputLimit)
			continue
		}
		out := formatCommandResult("journalctl", res.ExitCode, res.Stdout, res.Stderr)
		journals[key] = limitText(redaction.RedactText(out, redactKeys...), collectOutputLimit)
	}
	if err := writeJSON(path, payload); err != nil {
		return schema.ArtifactRef{}, err
	}
	return createCollectBundle(path, circuitFile, stageID, collectBundleDir)
}

func formatCommandResult(name string, exitCode int, stdout string, stderr string) string {
	out := strings.TrimSpace(stdout)
	errText := strings.TrimSpace(stderr)
	if out == "" {
		out = errText
	}
	if out == "" {
		out = "(empty output)"
	}
	return fmt.Sprintf("[%s exit=%d]\n%s", name, exitCode, out)
}

func limitText(input string, maxChars int) string {
	if maxChars <= 0 || len(input) <= maxChars {
		return input
	}
	const marker = "\n...(truncated)"
	if maxChars <= len(marker)+1 {
		return input[:maxChars]
	}
	return input[:maxChars-len(marker)] + marker
}

func createCollectBundle(collectFile string, circuitFile string, stageID string, collectBundleDir string) (schema.ArtifactRef, error) {
	evidenceDir := filepath.Dir(collectFile)
	rootDir := filepath.Dir(evidenceDir)
	bundlesDir := collectBundleDir
	if strings.TrimSpace(bundlesDir) == "" {
		bundlesDir = filepath.Join(rootDir, "bundles")
	}
	if strings.TrimSpace(stageID) == "" {
		stageID = "E"
	}
	name := fmt.Sprintf("collect-%s-%s.tar.gz", strings.ToUpper(stageID), time.Now().Format("20060102-150405"))
	absBundle := filepath.Join(bundlesDir, name)

	files := map[string]string{
		collectFile: filepath.Join("evidence", filepath.Base(collectFile)),
	}
	if fileExists(circuitFile) {
		files[circuitFile] = filepath.Join("state", filepath.Base(circuitFile))
	}
	for _, stateName := range []string{"overall.json", "lifecycle.json", "services.json", "artifacts.json"} {
		abs := filepath.Join(rootDir, "state", stateName)
		if fileExists(abs) {
			files[abs] = filepath.Join("state", stateName)
		}
	}
	manifestMeta := map[string]string{
		"bundle": "collect",
		"stage":  strings.ToUpper(stageID),
	}
	if err := reporting.CreateTarGzWithManifest(absBundle, files, manifestMeta); err != nil {
		return schema.ArtifactRef{}, err
	}

	rel, err := filepath.Rel(rootDir, absBundle)
	if err != nil || strings.HasPrefix(rel, "..") {
		rel = filepath.Join("bundles", name)
	}
	return schema.ArtifactRef{
		ID:   "collect-" + strings.ToLower(stageID),
		Path: filepath.ToSlash(rel),
	}, nil
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func boolMetric(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func summarizePaths(paths []string) []map[string]any {
	out := make([]map[string]any, 0, len(paths))
	for _, p := range paths {
		path := strings.TrimSpace(p)
		if path == "" {
			continue
		}
		item := map[string]any{
			"path":   path,
			"exists": false,
		}
		st, err := os.Stat(path)
		if err != nil {
			item["error"] = err.Error()
			out = append(out, item)
			continue
		}
		item["exists"] = true
		item["isDir"] = st.IsDir()
		item["size"] = st.Size()
		if st.IsDir() {
			item["entryCount"] = countDirEntries(path, 512)
		}
		out = append(out, item)
	}
	return out
}

func countDirEntries(path string, max int) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	entries, err := f.ReadDir(max)
	if err != nil {
		return 0
	}
	return len(entries)
}
