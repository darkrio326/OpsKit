package reporting

import (
	"archive/tar"
	"compress/gzip"
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
