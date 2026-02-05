package fsx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func EnsureDir(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func AtomicWriteJSON(path string, value any) error {
	if err := EnsureDir(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".tmp-*.json")
	if err != nil {
		return err
	}
	tmp := file.Name()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return err
	}

	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("atomic rename failed: %w", err)
	}
	return nil
}
