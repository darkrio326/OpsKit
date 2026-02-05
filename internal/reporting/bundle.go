package reporting

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func CreateTarGz(bundlePath string, files map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(bundlePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	entries := normalizeEntries(files)
	hashes := make([]string, 0, len(entries))
	for _, e := range entries {
		sum, err := fileSHA256(e.Abs)
		if err != nil {
			return err
		}
		hashes = append(hashes, fmt.Sprintf("%s  %s", sum, filepath.ToSlash(e.Rel)))
		if err := addFile(tw, e.Abs, e.Rel); err != nil {
			return err
		}
	}
	if err := addBytes(tw, []byte(strings.Join(hashes, "\n")+"\n"), "hashes.txt", 0o644); err != nil {
		return err
	}
	return nil
}

type bundleEntry struct {
	Abs string
	Rel string
}

func normalizeEntries(files map[string]string) []bundleEntry {
	entries := make([]bundleEntry, 0, len(files))
	for abs, rel := range files {
		if rel == "" {
			rel = filepath.Base(abs)
		}
		entries = append(entries, bundleEntry{Abs: abs, Rel: rel})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Rel == entries[j].Rel {
			return entries[i].Abs < entries[j].Abs
		}
		return entries[i].Rel < entries[j].Rel
	})
	return entries
}

func addFile(tw *tar.Writer, abs string, rel string) error {
	st, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("directories not supported in bundle directly: %s", abs)
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()

	hdr, err := tar.FileInfoHeader(st, "")
	if err != nil {
		return err
	}
	hdr.Name = filepath.ToSlash(rel)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}

func addBytes(tw *tar.Writer, b []byte, rel string, mode os.FileMode) error {
	hdr := &tar.Header{
		Name: filepath.ToSlash(rel),
		Mode: int64(mode),
		Size: int64(len(b)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(b)
	return err
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
