package reporting

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTarGzIncludesHashes(t *testing.T) {
	root := t.TempDir()
	f1 := filepath.Join(root, "a.txt")
	f2 := filepath.Join(root, "b.txt")
	if err := os.WriteFile(f1, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	if err := os.WriteFile(f2, []byte("beta"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}

	bundle := filepath.Join(root, "bundle.tar.gz")
	if err := CreateTarGz(bundle, map[string]string{f1: "evidence/a.txt", f2: "evidence/b.txt"}); err != nil {
		t.Fatalf("create bundle: %v", err)
	}

	containsHashes := false
	containsA := false
	containsB := false
	containsHashLine := false

	fp, err := os.Open(bundle)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer fp.Close()
	gz, err := gzip.NewReader(fp)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		switch hdr.Name {
		case "hashes.txt":
			containsHashes = true
			buf := make([]byte, hdr.Size)
			_, _ = tr.Read(buf)
			if strings.Contains(string(buf), "evidence/a.txt") {
				containsHashLine = true
			}
		case "evidence/a.txt":
			containsA = true
		case "evidence/b.txt":
			containsB = true
		}
	}

	if !containsA || !containsB || !containsHashes || !containsHashLine {
		t.Fatalf("bundle entries missing: a=%v b=%v hashes=%v hashline=%v", containsA, containsB, containsHashes, containsHashLine)
	}
}

func TestCreateTarGzHashesSorted(t *testing.T) {
	root := t.TempDir()
	f1 := filepath.Join(root, "z.txt")
	f2 := filepath.Join(root, "a.txt")
	if err := os.WriteFile(f1, []byte("zeta"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	if err := os.WriteFile(f2, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}

	bundle := filepath.Join(root, "bundle.tar.gz")
	if err := CreateTarGz(bundle, map[string]string{f1: "evidence/z.txt", f2: "evidence/a.txt"}); err != nil {
		t.Fatalf("create bundle: %v", err)
	}

	var hashLines []string
	fp, err := os.Open(bundle)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer fp.Close()
	gz, err := gzip.NewReader(fp)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if hdr.Name == "hashes.txt" {
			buf := make([]byte, hdr.Size)
			_, _ = tr.Read(buf)
			for _, line := range strings.Split(string(buf), "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					hashLines = append(hashLines, line)
				}
			}
		}
	}

	if len(hashLines) != 2 {
		t.Fatalf("expected 2 hash lines, got %d", len(hashLines))
	}

	paths := []string{}
	for _, line := range hashLines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			t.Fatalf("unexpected hash line: %q", line)
		}
		paths = append(paths, fields[1])
	}
	if paths[0] != "evidence/a.txt" || paths[1] != "evidence/z.txt" {
		t.Fatalf("hashes not sorted by path: %v", paths)
	}
}

func TestCreateTarGzManifestMatchesHashes(t *testing.T) {
	root := t.TempDir()
	f1 := filepath.Join(root, "a.txt")
	f2 := filepath.Join(root, "b.txt")
	if err := os.WriteFile(f1, []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write file1: %v", err)
	}
	if err := os.WriteFile(f2, []byte("beta"), 0o644); err != nil {
		t.Fatalf("write file2: %v", err)
	}

	bundle := filepath.Join(root, "bundle.tar.gz")
	meta := map[string]string{"stage": "F", "template": "demo"}
	if err := CreateTarGzWithManifest(bundle, map[string]string{f1: "evidence/a.txt", f2: "reports/b.txt"}, meta); err != nil {
		t.Fatalf("create bundle: %v", err)
	}

	hashes := map[string]string{}
	var manifest Manifest

	fp, err := os.Open(bundle)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer fp.Close()
	gz, err := gzip.NewReader(fp)
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		switch hdr.Name {
		case "hashes.txt":
			buf := make([]byte, hdr.Size)
			_, _ = tr.Read(buf)
			for _, line := range strings.Split(string(buf), "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					hashes[fields[1]] = fields[0]
				}
			}
		case "manifest.json":
			buf := make([]byte, hdr.Size)
			_, _ = tr.Read(buf)
			if err := json.Unmarshal(buf, &manifest); err != nil {
				t.Fatalf("parse manifest: %v", err)
			}
		}
	}

	if manifest.Version == "" || len(manifest.Files) == 0 {
		t.Fatalf("manifest missing files")
	}
	if manifest.Meta["stage"] != "F" || manifest.Meta["template"] != "demo" {
		t.Fatalf("manifest meta mismatch: %+v", manifest.Meta)
	}
	for _, f := range manifest.Files {
		if hashes[f.Path] == "" {
			t.Fatalf("hashes missing entry for %s", f.Path)
		}
		if hashes[f.Path] != f.SHA256 {
			t.Fatalf("hash mismatch for %s", f.Path)
		}
	}
}
