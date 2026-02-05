package actions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"opskit/internal/schema"
)

type baselineSnapshotAction struct{}

func (a *baselineSnapshotAction) Kind() string { return "baseline_snapshot" }

func (a *baselineSnapshotAction) Run(_ context.Context, req Request) (Result, error) {
	paths := toStringSlice(req.Params["paths"])
	output := toString(req.Params["output"], "")
	if output == "" {
		return Result{}, fmt.Errorf("baseline_snapshot requires params.output")
	}
	if len(paths) == 0 {
		return Result{ActionID: req.ID, Status: schema.StatusSkipped, Severity: schema.SeverityInfo, Message: "no baseline paths configured"}, nil
	}

	entries := make([]map[string]any, 0, len(paths))
	hashLines := make([]string, 0, len(paths))
	for _, p := range paths {
		path := strings.TrimSpace(p)
		if path == "" {
			continue
		}
		item, line := snapshotEntry(path)
		entries = append(entries, item)
		hashLines = append(hashLines, line)
	}
	sort.Strings(hashLines)
	sum := sha256.Sum256([]byte(strings.Join(hashLines, "\n")))
	baselineHash := hex.EncodeToString(sum[:])

	payload := map[string]any{
		"baselineId": req.ID,
		"paths":      entries,
		"hash":       baselineHash,
	}
	if err := writeJSON(output, payload); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "write baseline snapshot failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}

	return Result{
		ActionID: req.ID,
		Status:   schema.StatusPassed,
		Severity: schema.SeverityInfo,
		Message:  "baseline snapshot written: " + output,
		Metrics: []schema.Metric{
			{Label: "baseline_paths", Value: fmt.Sprintf("%d", len(entries))},
			{Label: "baseline_hash", Value: baselineHash[:12]},
		},
	}, nil
}

func snapshotEntry(path string) (map[string]any, string) {
	item := map[string]any{
		"path":   path,
		"exists": false,
	}
	st, err := os.Stat(path)
	if err != nil {
		item["error"] = err.Error()
		return item, path + "|missing"
	}
	item["exists"] = true
	item["isDir"] = st.IsDir()
	item["mode"] = st.Mode().String()
	item["size"] = st.Size()

	if st.IsDir() {
		item["entryCount"] = countTopEntries(path, 1024)
		return item, fmt.Sprintf("%s|dir|%s|%d|%d", path, st.Mode().String(), st.Size(), item["entryCount"])
	}
	return item, fmt.Sprintf("%s|file|%s|%d", path, st.Mode().String(), st.Size())
}

func countTopEntries(path string, max int) int {
	f, err := os.Open(filepath.Clean(path))
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
