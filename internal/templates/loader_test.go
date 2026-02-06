package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplate_UnknownField(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.json")
	content := `{"id":"t","name":"t","mode":"manage","unknown":true,"stages":{"A":{"checks":[{"id":"a.system_info","kind":"system_info"}]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatalf("expected unknown field error")
	}
}
