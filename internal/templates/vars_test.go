package templates

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"opskit/internal/schema"
)

func TestParseVarsFileJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "vars.json")
	content := `{
  "NAME": "demo",
  "PORT": 18080,
  "FLAG": true,
  "PORTS": [80, 443],
  "META": {"env":"test"}
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write vars file: %v", err)
	}
	vars, err := ParseVarsFile(path)
	if err != nil {
		t.Fatalf("parse vars file: %v", err)
	}
	if vars["NAME"] != "demo" || vars["PORT"] != "18080" || vars["FLAG"] != "true" {
		t.Fatalf("unexpected scalar vars: %+v", vars)
	}
	if vars["PORTS"] != "[80,443]" {
		t.Fatalf("unexpected array encoding: %s", vars["PORTS"])
	}
	if vars["META"] != "{\"env\":\"test\"}" {
		t.Fatalf("unexpected object encoding: %s", vars["META"])
	}
}

func TestParseVarsFileLines(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "vars.env")
	content := `
# comment
NAME=demo
PORT=18080
export FLAG=true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write vars file: %v", err)
	}
	vars, err := ParseVarsFile(path)
	if err != nil {
		t.Fatalf("parse vars file: %v", err)
	}
	if vars["NAME"] != "demo" || vars["PORT"] != "18080" || vars["FLAG"] != "true" {
		t.Fatalf("unexpected vars: %+v", vars)
	}
}

func TestApplyVarsJSONPlaceholder(t *testing.T) {
	tpl := schema.Template{
		ID:   "t",
		Name: "t",
		Mode: "manage",
		Stages: map[string]schema.TemplateStageSpec{
			"A": {
				Checks: []schema.TemplateStep{
					{ID: "a", Kind: "system_info", Params: map[string]any{
						"ports": "${PORTS}",
						"meta":  "${META}",
					}},
				},
			},
		},
	}
	vars := map[string]string{
		"PORTS": "[80,443]",
		"META":  "{\"env\":\"test\"}",
	}
	out := ApplyVars(tpl, vars)
	ports := out.Stages["A"].Checks[0].Params["ports"]
	meta := out.Stages["A"].Checks[0].Params["meta"]
	if _, ok := ports.([]any); !ok {
		t.Fatalf("expected ports to be array, got %T", ports)
	}
	if !reflect.DeepEqual(meta, map[string]any{"env": "test"}) {
		t.Fatalf("unexpected meta: %#v", meta)
	}
}
