package installer

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"opskit/internal/core/fsx"
	"opskit/internal/state"
)

//go:embed assets/*
var assets embed.FS

type Result struct {
	Actions []string
}

func InstallLayout(store *state.Store, dryRun bool) (Result, error) {
	paths := store.Paths()
	actions := []string{
		"ensure " + paths.StateDir,
		"ensure " + paths.ReportsDir,
		"ensure " + paths.EvidenceDir,
		"ensure " + paths.BundlesDir,
		"ensure " + paths.CacheDir,
		"write " + filepath.Join(paths.UIDir, "index.html"),
		"write " + filepath.Join(paths.UIDir, "style.css"),
		"write " + filepath.Join(paths.UIDir, "app.js"),
	}
	if dryRun {
		return Result{Actions: actions}, nil
	}

	if err := store.EnsureLayout(); err != nil {
		return Result{}, err
	}
	if err := writeUI(paths.UIDir); err != nil {
		return Result{}, err
	}

	return Result{Actions: actions}, nil
}

func writeUI(dir string) error {
	if err := fsx.EnsureDir(dir, 0o755); err != nil {
		return err
	}
	return fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, err := assets.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.Base(path)
		target := filepath.Join(dir, name)
		if err := os.WriteFile(target, b, 0o644); err != nil {
			return fmt.Errorf("write ui %s: %w", name, err)
		}
		return nil
	})
}
