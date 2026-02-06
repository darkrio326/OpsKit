package engine

import (
	"testing"

	"opskit/internal/schema"
)

func TestSummarizeStageSteps(t *testing.T) {
	summary := summarizeStageSteps([]schema.Status{
		schema.StatusPassed,
		schema.StatusWarn,
		schema.StatusFailed,
		schema.StatusSkipped,
		schema.StatusNotStarted,
		schema.StatusRunning,
	})
	if summary.Total != 6 {
		t.Fatalf("expected total 6, got %d", summary.Total)
	}
	if summary.Pass != 1 || summary.Warn != 1 || summary.Fail != 1 || summary.Skip != 3 {
		t.Fatalf("unexpected summary: %+v", *summary)
	}
}
