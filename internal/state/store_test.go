package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitStateIfMissingWritesFourJSONFiles(t *testing.T) {
	tmp := t.TempDir()
	store := NewStore(NewPaths(tmp))

	if err := store.InitStateIfMissing("generic-manage-v1"); err != nil {
		t.Fatalf("init state: %v", err)
	}

	for _, name := range []string{"overall.json", "lifecycle.json", "services.json", "artifacts.json"} {
		path := filepath.Join(store.Paths().StateDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected state file %s: %v", name, err)
		}
	}
}
