package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"opskit/internal/schema"
)

type fileHashEvidence struct{}

func (p *fileHashEvidence) Kind() string { return "file_hash" }

func (p *fileHashEvidence) Collect(_ context.Context, req Request) (Result, error) {
	filePath := toString(req.Params["file"], "")
	output := toString(req.Params["output"], "")
	if filePath == "" || output == "" {
		return Result{}, fmt.Errorf("file_hash requires params.file and params.output")
	}

	f, err := os.Open(filePath)
	if err != nil {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "read file failed: " + err.Error()}
		return Result{EvidenceID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return Result{}, err
	}
	sum := fmt.Sprintf("%x", h.Sum(nil))
	payload := map[string]any{"evidenceId": req.ID, "kind": "file_hash", "file": filePath, "sha256": sum}
	if err := ensureParent(output); err != nil {
		return Result{}, err
	}
	b, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(output, b, 0o644); err != nil {
		return Result{}, err
	}

	return Result{
		EvidenceID: req.ID,
		Status:     schema.StatusPassed,
		Severity:   schema.SeverityInfo,
		Message:    "file hash collected",
		Path:       output,
		Metrics:    []schema.Metric{{Label: "sha256", Value: sum}},
	}, nil
}
