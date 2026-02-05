package handover

import (
	"os"
	"path/filepath"
	"testing"

	"opskit/internal/state"
)

func TestGenerateWritesReportsAndBundle(t *testing.T) {
	root := t.TempDir()
	store := state.NewStore(state.NewPaths(root))
	if err := store.InitStateIfMissing("test"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	res, err := Generate(store)
	if err != nil {
		t.Fatalf("generate handover: %v", err)
	}

	for _, rel := range []string{res.ReportHTML.Path, res.ReportJSON.Path, res.Bundle.Path} {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("missing generated file %s: %v", rel, err)
		}
	}
}
