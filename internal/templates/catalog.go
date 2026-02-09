package templates

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"opskit/internal/schema"
)

type CatalogItem struct {
	Ref          string
	Aliases      []string
	Source       string
	TemplateID   string
	Name         string
	Mode         string
	ServiceScope string
	Tags         []string
}

type builtinRefSpec struct {
	Ref          string
	Aliases      []string
	File         string
	ServiceScope string
}

type CatalogOptions struct {
	IncludeDemo bool
	DemoDir     string
}

var demoScopeOverrides = map[string]string{
	"demo-blackbox-middleware-manage": "single-service",
	"demo-elk-deploy":                 "multi-service",
	"demo-generic-selfhost-deploy":    "single-service",
	"demo-hello-service":              "single-service",
	"demo-minio-deploy":               "single-service",
	"demo-powerjob-deploy":            "multi-service",
	"demo-runtime-baseline":           "single-service",
	"demo-server-audit":               "single-service",
}

var builtinRefSpecs = []builtinRefSpec{
	{
		Ref:          "generic-manage-v1",
		File:         "default-manage.json",
		ServiceScope: "single-service",
	},
	{
		Ref:          "single-service-deploy-v1",
		Aliases:      []string{"single-service-deploy"},
		File:         "single-service-deploy.json",
		ServiceScope: "single-service",
	},
}

func BuiltinCatalog() ([]CatalogItem, error) {
	items := make([]CatalogItem, 0, len(builtinRefSpecs))
	for _, spec := range builtinRefSpecs {
		t, err := loadBuiltin(spec.File)
		if err != nil {
			return nil, err
		}
		items = append(items, buildCatalogItem(
			spec.Ref,
			spec.Aliases,
			"builtin/"+spec.File,
			t,
			spec.ServiceScope,
			"builtin",
		))
	}
	sortCatalog(items)
	return items, nil
}

func Catalog(opt CatalogOptions) ([]CatalogItem, error) {
	items, err := BuiltinCatalog()
	if err != nil {
		return nil, err
	}
	if !opt.IncludeDemo {
		return items, nil
	}
	demos, err := DemoCatalog(opt.DemoDir)
	if err != nil {
		return nil, err
	}
	out := make([]CatalogItem, 0, len(items)+len(demos))
	out = append(out, items...)
	out = append(out, demos...)
	sortCatalog(out)
	return out, nil
}

func DemoCatalog(dir string) ([]CatalogItem, error) {
	demoDir, ok := resolveDemoDir(dir)
	if !ok {
		return []CatalogItem{}, nil
	}
	entries, err := os.ReadDir(demoDir)
	if err != nil {
		return nil, err
	}
	items := make([]CatalogItem, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		ref := strings.TrimSuffix(name, ".json")
		fullPath := filepath.Join(demoDir, name)
		t, err := loadFile(fullPath)
		if err != nil {
			return nil, err
		}
		source := filepath.ToSlash(filepath.Join("assets", "templates", name))
		scope := detectServiceScope(t)
		if v, ok := demoScopeOverrides[ref]; ok {
			scope = v
		}
		items = append(items, buildCatalogItem(
			ref,
			nil,
			source,
			t,
			scope,
			"demo",
		))
	}
	sortCatalog(items)
	return items, nil
}

func resolveBuiltinRef(ref string) (schema.Template, bool, error) {
	if ref == "" {
		t, err := loadBuiltin("default-manage.json")
		return t, true, err
	}
	for _, spec := range builtinRefSpecs {
		if ref == spec.Ref {
			t, err := loadBuiltin(spec.File)
			return t, true, err
		}
		for _, alias := range spec.Aliases {
			if ref == alias {
				t, err := loadBuiltin(spec.File)
				return t, true, err
			}
		}
	}
	return schema.Template{}, false, nil
}

func resolveDemoRef(ref string) (schema.Template, bool, error) {
	dir, ok := resolveDemoDir("")
	if !ok || strings.TrimSpace(ref) == "" {
		return schema.Template{}, false, nil
	}
	byNamePath := filepath.Join(dir, ref+".json")
	if _, err := os.Stat(byNamePath); err == nil {
		t, loadErr := loadFile(byNamePath)
		return t, true, loadErr
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return schema.Template{}, false, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		fullPath := filepath.Join(dir, e.Name())
		t, loadErr := loadFile(fullPath)
		if loadErr != nil {
			return schema.Template{}, false, loadErr
		}
		if t.ID == ref {
			return t, true, nil
		}
	}
	return schema.Template{}, false, nil
}

func resolveDemoDir(explicit string) (string, bool) {
	candidates := make([]string, 0, 4)
	if v := strings.TrimSpace(explicit); v != "" {
		candidates = append(candidates, v)
	}
	if v := strings.TrimSpace(os.Getenv("OPSKIT_TEMPLATE_DEMO_DIR")); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates, filepath.Join("assets", "templates"))
	if _, srcFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Clean(filepath.Join(filepath.Dir(srcFile), "..", "..", "assets", "templates")))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Clean(filepath.Join(filepath.Dir(exe), "assets", "templates")))
	}

	for _, c := range candidates {
		if strings.TrimSpace(c) == "" {
			continue
		}
		info, err := os.Stat(c)
		if err == nil && info.IsDir() {
			return c, true
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			continue
		}
	}
	return "", false
}

func detectServiceScope(t schema.Template) string {
	if stage, ok := t.Stages["C"]; ok {
		for _, a := range stage.Actions {
			if a.Kind != "declare_stack" || a.Params == nil {
				continue
			}
			raw, ok := a.Params["components"]
			if !ok {
				continue
			}
			n := listLen(raw)
			if n > 1 {
				return "multi-service"
			}
			if n == 1 {
				return "single-service"
			}
		}
	}
	return "single-service"
}

func listLen(v any) int {
	switch vv := v.(type) {
	case []any:
		return len(vv)
	case []string:
		return len(vv)
	default:
		return 0
	}
}

func buildCatalogItem(ref string, aliases []string, source string, t schema.Template, serviceScope string, sourceTag string) CatalogItem {
	tags := []string{t.Mode}
	if strings.TrimSpace(serviceScope) != "" {
		tags = append(tags, serviceScope)
	}
	if strings.TrimSpace(sourceTag) != "" {
		tags = append(tags, sourceTag)
	}
	return CatalogItem{
		Ref:          ref,
		Aliases:      append([]string{}, aliases...),
		Source:       source,
		TemplateID:   t.ID,
		Name:         t.Name,
		Mode:         t.Mode,
		ServiceScope: serviceScope,
		Tags:         tags,
	}
}

func sortCatalog(items []CatalogItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Ref < items[j].Ref
	})
}
