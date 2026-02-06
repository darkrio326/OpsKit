package actions

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

func TestCreateCollectBundle(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	evidenceDir := filepath.Join(root, "evidence")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	collectFile := filepath.Join(evidenceDir, "recover-collect.json")
	circuitFile := filepath.Join(stateDir, "recover_circuit.json")
	for _, f := range []string{
		collectFile,
		circuitFile,
		filepath.Join(stateDir, "overall.json"),
		filepath.Join(stateDir, "lifecycle.json"),
		filepath.Join(stateDir, "services.json"),
		filepath.Join(stateDir, "artifacts.json"),
	} {
		if err := os.WriteFile(f, []byte(`{"ok":true}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ref, err := createCollectBundle(collectFile, circuitFile, "E", "")
	if err != nil {
		t.Fatalf("createCollectBundle error: %v", err)
	}
	if ref.ID != "collect-e" {
		t.Fatalf("unexpected artifact id: %s", ref.ID)
	}
	if filepath.Ext(ref.Path) != ".gz" {
		t.Fatalf("unexpected artifact path: %s", ref.Path)
	}

	names, err := tarEntries(filepath.Join(root, ref.Path))
	if err != nil {
		t.Fatalf("read bundle error: %v", err)
	}
	for _, want := range []string{
		"evidence/recover-collect.json",
		"state/recover_circuit.json",
		"state/overall.json",
		"state/lifecycle.json",
		"state/services.json",
		"state/artifacts.json",
		"hashes.txt",
	} {
		if !slices.Contains(names, want) {
			t.Fatalf("missing bundle entry: %s", want)
		}
	}
}

func TestRecoverSequencePreconditionGeneratesCollectBundle(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	evidenceDir := filepath.Join(root, "evidence")
	bundlesDir := filepath.Join(root, "bundles")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{
		filepath.Join(stateDir, "overall.json"),
		filepath.Join(stateDir, "lifecycle.json"),
		filepath.Join(stateDir, "services.json"),
		filepath.Join(stateDir, "artifacts.json"),
	} {
		if err := os.WriteFile(f, []byte(`{"ok":true}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	plugin := &recoverSequenceAction{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "e.recover_sequence",
		Params: map[string]any{
			"units":              []any{"demo.service"},
			"max_attempts":       1,
			"readiness_network":  true,
			"readiness_mounts":   []any{"/"},
			"trigger_source":     "onboot",
			"stage":              "E",
			"circuit_file":       filepath.Join(stateDir, "recover_circuit.json"),
			"collect_file":       filepath.Join(evidenceDir, "recover-collect.json"),
			"collect_bundle_dir": bundlesDir,
			"collect_paths":      []any{stateDir},
		},
		Exec: fakeExecRunner{
			run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
				if spec.Name == "ip" {
					return executil.Result{}, fmtPrecondition(spec.Name)
				}
				return executil.Result{ExitCode: 0}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusFailed {
		t.Fatalf("expected failed status, got %s", res.Status)
	}
	if len(res.Bundles) == 0 {
		t.Fatalf("expected collect bundle artifact")
	}
	if !hasMetric(res.Metrics, "recover_trigger", "onboot") {
		t.Fatalf("expected recover_trigger metric onboot")
	}
	if !strings.Contains(res.Message, "readiness check failed") {
		t.Fatalf("unexpected message: %s", res.Message)
	}
	if _, statErr := os.Stat(filepath.Join(root, res.Bundles[0].Path)); statErr != nil {
		t.Fatalf("expected bundle file exists: %v", statErr)
	}

	collectPath := filepath.Join(evidenceDir, "recover-collect.json")
	payload, err := readJSONMap(collectPath)
	if err != nil {
		t.Fatalf("read collect payload failed: %v", err)
	}
	if payload["triggerSource"] != "onboot" {
		t.Fatalf("unexpected triggerSource: %v", payload["triggerSource"])
	}
	if _, ok := payload["paths"]; !ok {
		t.Fatalf("expected collect paths summary")
	}
	if _, ok := payload["journals"]; !ok {
		t.Fatalf("expected collect journals section")
	}
}

func TestRecoverSequenceAllowEmptyUnits(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	evidenceDir := filepath.Join(root, "evidence")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}

	plugin := &recoverSequenceAction{}
	res, err := plugin.Run(context.Background(), Request{
		ID: "e.recover_host",
		Params: map[string]any{
			"units":             []any{},
			"allow_empty_units": true,
			"max_attempts":      1,
			"readiness_network": false,
			"readiness_mounts":  []any{"/"},
			"trigger_source":    "manual",
			"stage":             "E",
			"circuit_file":      filepath.Join(stateDir, "recover_circuit.json"),
			"collect_file":      filepath.Join(evidenceDir, "recover-collect.json"),
		},
		Exec: fakeExecRunner{
			run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
				return executil.Result{ExitCode: 0}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != schema.StatusPassed {
		t.Fatalf("expected passed status, got %s", res.Status)
	}
	if !hasMetric(res.Metrics, "recovered_units", "0") {
		t.Fatalf("expected recovered_units metric 0")
	}
}

func TestWriteRecoverCollectRedactionAndLimit(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	evidenceDir := filepath.Join(root, "evidence")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{
		filepath.Join(stateDir, "overall.json"),
		filepath.Join(stateDir, "lifecycle.json"),
		filepath.Join(stateDir, "services.json"),
		filepath.Join(stateDir, "artifacts.json"),
		filepath.Join(stateDir, "recover_circuit.json"),
	} {
		if err := os.WriteFile(f, []byte(`{"ok":true}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	req := Request{
		ID: "e.recover_sequence",
		Exec: fakeExecRunner{
			run: func(_ context.Context, spec executil.Spec) (executil.Result, error) {
				switch spec.Name {
				case "journalctl":
					return executil.Result{ExitCode: 0, Stdout: "password=abc " + strings.Repeat("J", 200)}, nil
				default:
					return executil.Result{ExitCode: 0, Stdout: "token=xyz " + strings.Repeat("S", 200)}, nil
				}
			},
		},
	}
	collectFile := filepath.Join(evidenceDir, "recover-collect.json")
	_, err := writeRecoverCollect(
		context.Background(),
		req,
		collectFile,
		[]string{"demo.service"},
		"recover failed token=xyz",
		filepath.Join(stateDir, "recover_circuit.json"),
		"E",
		"",
		"manual",
		[]string{stateDir},
		[]string{"token", "password"},
		120,
	)
	if err != nil {
		t.Fatalf("writeRecoverCollect: %v", err)
	}

	payload, err := readJSONMap(collectFile)
	if err != nil {
		t.Fatalf("read collect payload failed: %v", err)
	}
	commands, ok := payload["commands"].(map[string]any)
	if !ok {
		t.Fatalf("commands section missing")
	}
	ss, _ := commands["ss_ltn"].(string)
	if strings.Contains(ss, "xyz") {
		t.Fatalf("expected token redaction in command output")
	}
	if !strings.Contains(ss, "...(truncated)") {
		t.Fatalf("expected command output truncation marker")
	}
	journals, ok := payload["journals"].(map[string]any)
	if !ok {
		t.Fatalf("journals section missing")
	}
	j, _ := journals["journal_demo_service"].(string)
	if strings.Contains(j, "abc") {
		t.Fatalf("expected password redaction in journal output")
	}
	limits, ok := payload["limits"].(map[string]any)
	if !ok {
		t.Fatalf("limits section missing")
	}
	maxChars, _ := limits["maxChars"].(float64)
	if int(maxChars) != 120 {
		t.Fatalf("unexpected maxChars: %v", limits["maxChars"])
	}
}

func fmtPrecondition(cmd string) error {
	return errors.Join(coreerr.ErrPreconditionFailed, errors.New("required command not found: "+cmd))
}

type fakeExecRunner struct {
	run func(context.Context, executil.Spec) (executil.Result, error)
}

func (f fakeExecRunner) Run(ctx context.Context, spec executil.Spec) (executil.Result, error) {
	if f.run == nil {
		return executil.Result{}, nil
	}
	return f.run(ctx, spec)
}

func tarEntries(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	var out []string
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		out = append(out, h.Name)
	}
	return out, nil
}

func hasMetric(metrics []schema.Metric, label, value string) bool {
	for _, m := range metrics {
		if m.Label == label && m.Value == value {
			return true
		}
	}
	return false
}

func readJSONMap(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}
