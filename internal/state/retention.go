package state

import (
	"os"
	"path/filepath"

	"opskit/internal/schema"
)

const (
	DefaultMaxReports = 50
	DefaultMaxBundles = 20
)

func ApplyArtifactRetention(paths Paths, artifacts *schema.ArtifactsState, maxReports int, maxBundles int) error {
	if artifacts == nil {
		return nil
	}
	removedReports := trimArtifactRefs(&artifacts.Reports, maxReports)
	removedBundles := trimArtifactRefs(&artifacts.Bundles, maxBundles)

	for _, ref := range append(removedReports, removedBundles...) {
		abs := resolveArtifactPath(paths.Root, ref.Path)
		if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func trimArtifactRefs(items *[]schema.ArtifactRef, max int) []schema.ArtifactRef {
	if items == nil || max <= 0 {
		return nil
	}
	current := *items
	if len(current) <= max {
		return nil
	}
	cut := len(current) - max
	removed := append([]schema.ArtifactRef{}, current[:cut]...)
	*items = append([]schema.ArtifactRef{}, current[cut:]...)
	return removed
}

func resolveArtifactPath(root string, relOrAbs string) string {
	if filepath.IsAbs(relOrAbs) {
		return relOrAbs
	}
	if root == "" {
		return relOrAbs
	}
	return filepath.Join(root, relOrAbs)
}
