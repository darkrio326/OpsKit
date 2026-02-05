package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"opskit/internal/core/fsx"
	"opskit/internal/core/timeutil"
	"opskit/internal/schema"
)

type Paths struct {
	Root        string
	StateDir    string
	ReportsDir  string
	EvidenceDir string
	BundlesDir  string
	CacheDir    string
	UIDir       string
	LockFile    string
}

func NewPaths(base string) Paths {
	return Paths{
		Root:        base,
		StateDir:    filepath.Join(base, "state"),
		ReportsDir:  filepath.Join(base, "reports"),
		EvidenceDir: filepath.Join(base, "evidence"),
		BundlesDir:  filepath.Join(base, "bundles"),
		CacheDir:    filepath.Join(base, "cache"),
		UIDir:       filepath.Join(base, "ui"),
		LockFile:    filepath.Join(base, "state", "opskit.lock"),
	}
}

type Store struct {
	paths Paths
}

func NewStore(paths Paths) *Store {
	return &Store{paths: paths}
}

func (s *Store) Paths() Paths {
	return s.paths
}

func (s *Store) EnsureLayout() error {
	dirs := []string{s.paths.StateDir, s.paths.ReportsDir, s.paths.EvidenceDir, s.paths.BundlesDir, s.paths.CacheDir, s.paths.UIDir}
	for _, d := range dirs {
		if err := fsx.EnsureDir(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) InitStateIfMissing(templateID string) error {
	if err := s.EnsureLayout(); err != nil {
		return err
	}
	if _, err := os.Stat(s.overallPath()); os.IsNotExist(err) {
		if err := s.WriteOverall(DefaultOverall(templateID)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.lifecyclePath()); os.IsNotExist(err) {
		if err := s.WriteLifecycle(DefaultLifecycle()); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.servicesPath()); os.IsNotExist(err) {
		if err := s.WriteServices(schema.ServicesState{Services: []schema.ServiceState{}}); err != nil {
			return err
		}
	}
	if _, err := os.Stat(s.artifactsPath()); os.IsNotExist(err) {
		if err := s.WriteArtifacts(schema.ArtifactsState{Reports: []schema.ArtifactRef{}, Bundles: []schema.ArtifactRef{}}); err != nil {
			return err
		}
	}
	return nil
}

func DefaultOverall(templateID string) schema.OverallState {
	active := []string{}
	if templateID != "" {
		active = []string{templateID}
	}
	return schema.OverallState{
		OverallStatus:   schema.OverallUnknown,
		LastRefreshTime: timeutil.NowISO8601(),
		ActiveTemplates: active,
		OpenIssuesCount: 0,
	}
}

func DefaultLifecycle() schema.LifecycleState {
	stages := make([]schema.StageState, 0, 6)
	for _, id := range []string{"A", "B", "C", "D", "E", "F"} {
		stages = append(stages, schema.StageState{
			StageID: id,
			Name:    stageName(id),
			Status:  schema.StatusNotStarted,
		})
	}
	return schema.LifecycleState{Stages: stages}
}

func stageName(id string) string {
	switch id {
	case "A":
		return "Preflight"
	case "B":
		return "Baseline"
	case "C":
		return "Deploy"
	case "D":
		return "Operate"
	case "E":
		return "Recover"
	case "F":
		return "Accept/Handover"
	default:
		return id
	}
}

func (s *Store) ReadOverall() (schema.OverallState, error) {
	var out schema.OverallState
	err := s.readJSON(s.overallPath(), &out)
	return out, err
}

func (s *Store) ReadLifecycle() (schema.LifecycleState, error) {
	var out schema.LifecycleState
	err := s.readJSON(s.lifecyclePath(), &out)
	return out, err
}

func (s *Store) ReadServices() (schema.ServicesState, error) {
	var out schema.ServicesState
	err := s.readJSON(s.servicesPath(), &out)
	return out, err
}

func (s *Store) ReadArtifacts() (schema.ArtifactsState, error) {
	var out schema.ArtifactsState
	err := s.readJSON(s.artifactsPath(), &out)
	return out, err
}

func (s *Store) WriteOverall(v schema.OverallState) error {
	return fsx.AtomicWriteJSON(s.overallPath(), v)
}

func (s *Store) WriteLifecycle(v schema.LifecycleState) error {
	return fsx.AtomicWriteJSON(s.lifecyclePath(), v)
}

func (s *Store) WriteServices(v schema.ServicesState) error {
	return fsx.AtomicWriteJSON(s.servicesPath(), v)
}

func (s *Store) WriteArtifacts(v schema.ArtifactsState) error {
	return fsx.AtomicWriteJSON(s.artifactsPath(), v)
}

func (s *Store) WriteReportStub(filename string, title string, body string) error {
	path := filepath.Join(s.paths.ReportsDir, filename)
	if err := fsx.EnsureDir(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body><h1>%s</h1><pre>%s</pre></body></html>", title, title, body)
	return os.WriteFile(path, []byte(content), 0o644)
}

func (s *Store) readJSON(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func (s *Store) overallPath() string {
	return filepath.Join(s.paths.StateDir, "overall.json")
}
func (s *Store) lifecyclePath() string {
	return filepath.Join(s.paths.StateDir, "lifecycle.json")
}
func (s *Store) servicesPath() string {
	return filepath.Join(s.paths.StateDir, "services.json")
}
func (s *Store) artifactsPath() string {
	return filepath.Join(s.paths.StateDir, "artifacts.json")
}
