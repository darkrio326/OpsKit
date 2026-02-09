package stages

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"opskit/internal/engine"
	evidenceplugin "opskit/internal/plugins/evidence"
	"opskit/internal/schema"
	"opskit/internal/state"
	"opskit/internal/templates"
)

func TestExecuteEvidenceStageWritesConsistencyReportAndSummary(t *testing.T) {
	paths := state.NewPaths(t.TempDir())
	store := state.NewStore(paths)
	if err := store.InitStateIfMissing("demo-template"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	reportFile := filepath.Join(paths.EvidenceDir, "demo-evidence.json")
	reg := evidenceplugin.NewRegistry()
	reg.Register(func() evidenceplugin.Plugin { return &fakeEvidencePlugin{} })

	rt := &engine.Runtime{
		Store:            store,
		EvidenceRegistry: reg,
		Options: engine.RunOptions{
			TemplateID:   "demo-template",
			TemplateMode: "manage",
		},
	}
	stage := templates.StagePlan{
		StageID: "F",
		Evidence: []schema.TemplateStep{
			{ID: "f.fake_evidence", Kind: "fake_evidence", Params: map[string]any{"output": reportFile}},
		},
	}

	res, err := executeEvidenceStage(context.Background(), rt, "F", "Accept Report", stage)
	if err != nil {
		t.Fatalf("execute evidence stage: %v", err)
	}
	if res.Report == "" {
		t.Fatalf("expected report file")
	}
	if len(res.Reports) != 1 || res.Reports[0].ID != "acceptance-consistency" {
		t.Fatalf("expected acceptance consistency report ref, got %+v", res.Reports)
	}
	if len(res.Bundles) != 1 || res.Bundles[0].ID != "acceptance" {
		t.Fatalf("expected acceptance bundle ref, got %+v", res.Bundles)
	}

	consistencyPath := filepath.Join(paths.Root, res.Reports[0].Path)
	b, err := os.ReadFile(consistencyPath)
	if err != nil {
		t.Fatalf("read consistency file: %v", err)
	}
	var consistency map[string]any
	if err := json.Unmarshal(b, &consistency); err != nil {
		t.Fatalf("parse consistency file: %v", err)
	}
	ok, _ := consistency["ok"].(bool)
	if !ok {
		t.Fatalf("expected consistency ok report: %+v", consistency)
	}

	reportBody, err := os.ReadFile(filepath.Join(paths.ReportsDir, res.Report))
	if err != nil {
		t.Fatalf("read report: %v", err)
	}

	bundlePath := filepath.Join(paths.Root, res.Bundles[0].Path)
	bundleReportBody, err := readTarGzFile(bundlePath, filepath.ToSlash(filepath.Join("reports", res.Report)))
	if err != nil {
		t.Fatalf("read report from bundle: %v", err)
	}
	if !bytes.Equal(reportBody, bundleReportBody) {
		t.Fatalf("expected report in bundle equals on-disk report")
	}
}

func readTarGzFile(bundlePath string, target string) ([]byte, error) {
	f, err := os.Open(bundlePath)
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
	for {
		h, err := tr.Next()
		if err != nil {
			return nil, err
		}
		if filepath.ToSlash(h.Name) == target {
			return ioReadAll(tr)
		}
	}
}

func ioReadAll(r *tar.Reader) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type fakeEvidencePlugin struct{}

func (p *fakeEvidencePlugin) Kind() string { return "fake_evidence" }

func (p *fakeEvidencePlugin) Collect(_ context.Context, req evidenceplugin.Request) (evidenceplugin.Result, error) {
	out := req.Params["output"].(string)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return evidenceplugin.Result{}, err
	}
	if err := os.WriteFile(out, []byte(`{"ok":true}`), 0o644); err != nil {
		return evidenceplugin.Result{}, err
	}
	return evidenceplugin.Result{
		EvidenceID: req.ID,
		Status:     schema.StatusPassed,
		Severity:   schema.SeverityInfo,
		Message:    "fake evidence created",
		Path:       out,
	}, nil
}
