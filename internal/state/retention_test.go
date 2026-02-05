package state

import (
	"os"
	"path/filepath"
	"testing"

	"opskit/internal/schema"
)

func TestApplyArtifactRetentionKeepsLatestN(t *testing.T) {
	root := t.TempDir()
	paths := NewPaths(root)
	if err := os.MkdirAll(paths.ReportsDir, 0o755); err != nil {
		t.Fatalf("mkdir reports: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "bundles"), 0o755); err != nil {
		t.Fatalf("mkdir bundles: %v", err)
	}

	for _, f := range []string{"r1.html", "r2.html", "r3.html"} {
		if err := os.WriteFile(filepath.Join(paths.ReportsDir, f), []byte("x"), 0o644); err != nil {
			t.Fatalf("write report: %v", err)
		}
	}
	for _, f := range []string{"b1.tgz", "b2.tgz"} {
		if err := os.WriteFile(filepath.Join(root, "bundles", f), []byte("x"), 0o644); err != nil {
			t.Fatalf("write bundle: %v", err)
		}
	}

	artifacts := schema.ArtifactsState{
		Reports: []schema.ArtifactRef{{ID: "r1", Path: "reports/r1.html"}, {ID: "r2", Path: "reports/r2.html"}, {ID: "r3", Path: "reports/r3.html"}},
		Bundles: []schema.ArtifactRef{{ID: "b1", Path: "bundles/b1.tgz"}, {ID: "b2", Path: "bundles/b2.tgz"}},
	}

	if err := ApplyArtifactRetention(paths, &artifacts, 2, 1); err != nil {
		t.Fatalf("apply retention: %v", err)
	}
	if len(artifacts.Reports) != 2 || artifacts.Reports[0].ID != "r2" || artifacts.Reports[1].ID != "r3" {
		t.Fatalf("unexpected reports after retention: %+v", artifacts.Reports)
	}
	if len(artifacts.Bundles) != 1 || artifacts.Bundles[0].ID != "b2" {
		t.Fatalf("unexpected bundles after retention: %+v", artifacts.Bundles)
	}

	if _, err := os.Stat(filepath.Join(paths.ReportsDir, "r1.html")); !os.IsNotExist(err) {
		t.Fatalf("expected old report deleted")
	}
	if _, err := os.Stat(filepath.Join(root, "bundles", "b1.tgz")); !os.IsNotExist(err) {
		t.Fatalf("expected old bundle deleted")
	}
}
