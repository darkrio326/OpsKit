package reporting

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyBundleConsistencyOK(t *testing.T) {
	root := t.TempDir()
	stateDir := filepath.Join(root, "state")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	overall := filepath.Join(stateDir, "overall.json")
	if err := os.WriteFile(overall, []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("write overall: %v", err)
	}
	report := filepath.Join(root, "report.html")
	if err := os.WriteFile(report, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}

	bundle := filepath.Join(root, "acceptance.tar.gz")
	if err := CreateTarGzWithManifest(bundle, map[string]string{
		overall: "state/overall.json",
		report:  "reports/report.html",
	}, map[string]string{"stage": "F"}); err != nil {
		t.Fatalf("create bundle: %v", err)
	}

	r, err := VerifyBundleConsistency(bundle, []string{"state/overall.json"})
	if err != nil {
		t.Fatalf("verify consistency: %v", err)
	}
	if !r.OK {
		t.Fatalf("expected consistency ok, got %+v", r)
	}
}

func TestVerifyBundleConsistencyMissingRequiredState(t *testing.T) {
	root := t.TempDir()
	report := filepath.Join(root, "report.html")
	if err := os.WriteFile(report, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}

	bundle := filepath.Join(root, "acceptance.tar.gz")
	if err := CreateTarGzWithManifest(bundle, map[string]string{
		report: "reports/report.html",
	}, nil); err != nil {
		t.Fatalf("create bundle: %v", err)
	}

	r, err := VerifyBundleConsistency(bundle, []string{"state/overall.json"})
	if err != nil {
		t.Fatalf("verify consistency: %v", err)
	}
	if r.OK {
		t.Fatalf("expected consistency fail")
	}
	if len(r.MissingRequiredState) != 1 || r.MissingRequiredState[0] != "state/overall.json" {
		t.Fatalf("unexpected missing required states: %+v", r.MissingRequiredState)
	}
}

func TestVerifyBundleConsistencyHashMismatch(t *testing.T) {
	root := t.TempDir()
	bundle := filepath.Join(root, "bad.tar.gz")

	f, err := os.Create(bundle)
	if err != nil {
		t.Fatalf("create bundle: %v", err)
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	manifest := Manifest{
		Version:     "v1",
		GeneratedAt: "2026-01-01T00:00:00Z",
		Files: []ManifestFile{
			{Path: "state/overall.json", SHA256: "111"},
		},
	}
	mb, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := addBytes(tw, []byte("222  state/overall.json\n"), "hashes.txt", 0o644); err != nil {
		t.Fatalf("add hashes: %v", err)
	}
	if err := addBytes(tw, mb, "manifest.json", 0o644); err != nil {
		t.Fatalf("add manifest: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	r, err := VerifyBundleConsistency(bundle, []string{"state/overall.json"})
	if err != nil {
		t.Fatalf("verify consistency: %v", err)
	}
	if r.OK {
		t.Fatalf("expected consistency fail")
	}
	if len(r.HashMismatch) != 1 || r.HashMismatch[0] != "state/overall.json" {
		t.Fatalf("unexpected hash mismatch: %+v", r.HashMismatch)
	}
}
