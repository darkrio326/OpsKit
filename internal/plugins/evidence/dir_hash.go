package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"opskit/internal/schema"
)

type dirHashEvidence struct{}

func (p *dirHashEvidence) Kind() string { return "dir_hash" }

func (p *dirHashEvidence) Collect(_ context.Context, req Request) (Result, error) {
	dir := toString(req.Params["dir"], "")
	output := toString(req.Params["output"], "")
	if dir == "" || output == "" {
		return Result{}, fmt.Errorf("dir_hash requires params.dir and params.output")
	}

	files := make([]string, 0)
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, rel)
		return nil
	}); err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "walk dir failed: " + err.Error()}
		return Result{EvidenceID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	sort.Strings(files)

	records := make([]map[string]string, 0, len(files))
	for _, rel := range files {
		p := filepath.Join(dir, rel)
		f, err := os.Open(p)
		if err != nil {
			return Result{}, err
		}
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			_ = f.Close()
			return Result{}, err
		}
		_ = f.Close()
		records = append(records, map[string]string{"file": rel, "sha256": hex.EncodeToString(h.Sum(nil))})
	}

	payload := map[string]any{"evidenceId": req.ID, "kind": "dir_hash", "dir": dir, "files": records}
	if err := ensureParent(output); err != nil {
		return Result{}, err
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(output, b, 0o644); err != nil {
		return Result{}, err
	}

	return Result{EvidenceID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "dir hash collected", Path: output, Metrics: []schema.Metric{{Label: "dir_files", Value: fmt.Sprintf("%d", len(records))}}}, nil
}
