package reporting

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type BundleConsistency struct {
	Bundle               string   `json:"bundle"`
	CheckedAt            string   `json:"checkedAt"`
	ManifestPresent      bool     `json:"manifestPresent"`
	HashesPresent        bool     `json:"hashesPresent"`
	ManifestCount        int      `json:"manifestCount"`
	HashesCount          int      `json:"hashesCount"`
	ManifestHashesMatch  bool     `json:"manifestHashesMatch"`
	MissingInManifest    []string `json:"missingInManifest,omitempty"`
	MissingInHashes      []string `json:"missingInHashes,omitempty"`
	HashMismatch         []string `json:"hashMismatch,omitempty"`
	RequiredStatePaths   []string `json:"requiredStatePaths,omitempty"`
	MissingRequiredState []string `json:"missingRequiredState,omitempty"`
	OK                   bool     `json:"ok"`
}

func VerifyBundleConsistency(bundlePath string, requiredStatePaths []string) (BundleConsistency, error) {
	report := BundleConsistency{
		Bundle:      filepath.Base(bundlePath),
		CheckedAt:   time.Now().Format(time.RFC3339),
		HashesCount: 0,
	}

	manifest, hashes, err := readBundleIndex(bundlePath)
	if err != nil {
		return report, err
	}
	report.ManifestPresent = manifest != nil
	report.HashesPresent = hashes != nil
	if manifest != nil {
		report.ManifestCount = len(manifest.Files)
	}
	if hashes != nil {
		report.HashesCount = len(hashes)
	}

	manifestMap := map[string]string{}
	if manifest != nil {
		for _, f := range manifest.Files {
			manifestMap[f.Path] = f.SHA256
		}
	}
	hashMap := map[string]string{}
	if hashes != nil {
		for p, h := range hashes {
			hashMap[p] = h
		}
	}

	for p, mh := range manifestMap {
		hh, ok := hashMap[p]
		if !ok {
			report.MissingInHashes = append(report.MissingInHashes, p)
			continue
		}
		if hh != mh {
			report.HashMismatch = append(report.HashMismatch, p)
		}
	}
	for p := range hashMap {
		if _, ok := manifestMap[p]; !ok {
			report.MissingInManifest = append(report.MissingInManifest, p)
		}
	}

	required := dedupeAndSort(requiredStatePaths)
	report.RequiredStatePaths = required
	for _, p := range required {
		_, inManifest := manifestMap[p]
		_, inHashes := hashMap[p]
		if !inManifest || !inHashes {
			report.MissingRequiredState = append(report.MissingRequiredState, p)
		}
	}

	sort.Strings(report.MissingInManifest)
	sort.Strings(report.MissingInHashes)
	sort.Strings(report.HashMismatch)
	sort.Strings(report.MissingRequiredState)

	report.ManifestHashesMatch = len(report.MissingInManifest) == 0 && len(report.MissingInHashes) == 0 && len(report.HashMismatch) == 0
	report.OK = report.ManifestPresent &&
		report.HashesPresent &&
		report.ManifestHashesMatch &&
		len(report.MissingRequiredState) == 0
	return report, nil
}

func readBundleIndex(bundlePath string) (*Manifest, map[string]string, error) {
	f, err := os.Open(bundlePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var manifest *Manifest
	var hashes map[string]string
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		switch hdr.Name {
		case "manifest.json":
			b, readErr := io.ReadAll(tr)
			if readErr != nil {
				return nil, nil, readErr
			}
			var m Manifest
			if err := json.Unmarshal(b, &m); err != nil {
				return nil, nil, err
			}
			manifest = &m
		case "hashes.txt":
			b, readErr := io.ReadAll(tr)
			if readErr != nil {
				return nil, nil, readErr
			}
			parsed, err := parseHashes(string(b))
			if err != nil {
				return nil, nil, err
			}
			hashes = parsed
		}
	}
	if manifest == nil {
		return nil, hashes, nil
	}
	return manifest, hashes, nil
}

func parseHashes(content string) (map[string]string, error) {
	out := map[string]string{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid hashes line: %q", line)
		}
		path := filepath.ToSlash(fields[1])
		out[path] = fields[0]
	}
	return out, nil
}

func dedupeAndSort(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = filepath.ToSlash(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
