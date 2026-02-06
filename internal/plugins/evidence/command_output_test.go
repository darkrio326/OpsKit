package evidence

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"opskit/internal/core/executil"
	"opskit/internal/core/redaction"
)

type fakeRunner struct {
	result executil.Result
	err    error
}

func (f *fakeRunner) Run(_ context.Context, _ executil.Spec) (executil.Result, error) {
	return f.result, f.err
}

func TestCommandOutputEvidence_Redaction(t *testing.T) {
	tmp := t.TempDir()
	output := filepath.Join(tmp, "out.json")
	runner := &fakeRunner{
		result: executil.Result{
			Stdout:   "password=abc token=xyz secret=zzz",
			Stderr:   "--password aaa --token bbb",
			ExitCode: 0,
		},
	}
	plugin := &commandOutputEvidence{}
	_, err := plugin.Collect(context.Background(), Request{
		ID: "e1",
		Params: map[string]any{
			"name":        "dummy",
			"args":        []any{"--password", "abc", "token=xyz"},
			"output":      output,
			"redact_keys": []any{"password", "token", "secret"},
		},
		Exec: runner,
	})
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	text := string(b)
	for _, plain := range []string{"abc", "xyz", "zzz", "aaa", "bbb"} {
		if strings.Contains(text, plain) {
			t.Fatalf("expected redaction to remove %q", plain)
		}
	}
	if !strings.Contains(text, redaction.Mask) {
		t.Fatalf("expected redaction mask in output")
	}
}
