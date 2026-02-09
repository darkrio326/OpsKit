package templates

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadTemplate_BuiltinAlias(t *testing.T) {
	if _, err := Load("single-service-deploy"); err != nil {
		t.Fatalf("load single-service-deploy alias: %v", err)
	}
	if _, err := Load("single-service-deploy-v1"); err != nil {
		t.Fatalf("load single-service-deploy-v1 ref: %v", err)
	}
	if _, err := Load("generic-manage-v1"); err != nil {
		t.Fatalf("load generic-manage-v1 ref: %v", err)
	}
}

func TestLoadTemplate_DemoRef(t *testing.T) {
	if _, err := Load("demo-server-audit"); err != nil {
		t.Fatalf("load demo-server-audit ref: %v", err)
	}
	if _, err := Load("demo-elk-deploy-v1"); err != nil {
		t.Fatalf("load demo-elk-deploy-v1 template id: %v", err)
	}
}

func TestBuiltinCatalog(t *testing.T) {
	items, err := BuiltinCatalog()
	if err != nil {
		t.Fatalf("builtin catalog: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least 2 builtin templates, got %d", len(items))
	}
	foundManage := false
	foundSingle := false
	for _, item := range items {
		if strings.TrimSpace(item.Ref) == "" {
			t.Fatalf("empty ref in builtin catalog item: %+v", item)
		}
		if strings.TrimSpace(item.TemplateID) == "" {
			t.Fatalf("empty template id in builtin catalog item: %+v", item)
		}
		if !strings.HasPrefix(item.Source, "builtin/") {
			t.Fatalf("unexpected source: %s", item.Source)
		}
		if strings.TrimSpace(item.ServiceScope) == "" {
			t.Fatalf("empty service scope in builtin catalog item: %+v", item)
		}
		if len(item.Tags) == 0 {
			t.Fatalf("empty tags in builtin catalog item: %+v", item)
		}
		if item.Ref == "generic-manage-v1" {
			foundManage = true
		}
		if item.Ref == "single-service-deploy-v1" {
			foundSingle = true
		}
	}
	if !foundManage {
		t.Fatalf("generic-manage-v1 not found in builtin catalog")
	}
	if !foundSingle {
		t.Fatalf("single-service-deploy-v1 not found in builtin catalog")
	}
}

func TestCatalog_IncludesDemo(t *testing.T) {
	items, err := Catalog(CatalogOptions{IncludeDemo: true})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if len(items) < 3 {
		t.Fatalf("expected builtin + demo templates, got %d", len(items))
	}
	foundDemo := false
	foundMulti := false
	for _, item := range items {
		if strings.HasPrefix(item.Ref, "demo-") {
			foundDemo = true
		}
		if item.ServiceScope == "multi-service" {
			foundMulti = true
		}
	}
	if !foundDemo {
		t.Fatalf("expected demo templates in catalog")
	}
	if !foundMulti {
		t.Fatalf("expected at least one multi-service template in catalog")
	}
}
