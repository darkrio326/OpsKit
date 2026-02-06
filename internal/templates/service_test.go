package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveSingleServiceDeployWithVars(t *testing.T) {
	tpl, vars, err := Resolve(ResolveOptions{
		TemplateRef: "single-service-deploy",
		BaseDir:     "/tmp/opskit",
		VarsRaw:     "SERVICE_NAME=demo,SERVICE_UNIT=demo.service,SERVICE_PORT=19090",
	})
	if err != nil {
		t.Fatalf("resolve template: %v", err)
	}
	if tpl.ID != "single-service-deploy-v1" {
		t.Fatalf("unexpected template id: %s", tpl.ID)
	}
	if vars["SERVICE_PORT"] != "19090" {
		t.Fatalf("vars not merged: %+v", vars)
	}

	dStage := tpl.Stages["D"]
	if len(dStage.Checks) == 0 {
		t.Fatalf("expected D checks")
	}
	unitParam := dStage.Checks[0].Params["unit"]
	if unitParam != "demo.service" {
		t.Fatalf("expected rendered unit param, got %v", unitParam)
	}
	portParam := dStage.Checks[2].Params["port"]
	if portParam != "19090" {
		t.Fatalf("expected rendered port param, got %v", portParam)
	}

	fStage := tpl.Stages["F"]
	if len(fStage.Evidence) < 1 {
		t.Fatalf("expected F evidence steps")
	}
	ePath := fStage.Evidence[0].Params["output"]
	if ePath == nil {
		t.Fatalf("expected evidence output path to be rendered")
	}
}

func TestResolveTemplateVarsDefaultAndRequired(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "t.json")
	content := `{
  "id": "t",
  "name": "t",
  "mode": "manage",
  "vars": {
    "FOO": { "type": "string", "required": true, "default": "bar" },
    "NUM": { "type": "int", "required": true, "default": "3" },
    "FLAG": { "type": "bool", "default": "true" }
  },
  "stages": {
    "A": {
      "checks": [
        { "id": "a.system_info", "kind": "system_info", "params": { "note": "${FOO}", "num": "${NUM}", "flag": "${FLAG}" } }
      ]
    }
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	tpl, vars, err := Resolve(ResolveOptions{TemplateRef: path, BaseDir: tmp})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if vars["FOO"] != "bar" || vars["NUM"] != "3" || vars["FLAG"] != "true" {
		t.Fatalf("unexpected vars: %+v", vars)
	}
	note := tpl.Stages["A"].Checks[0].Params["note"]
	if note != "bar" {
		t.Fatalf("expected rendered note, got %v", note)
	}
}

func TestResolveTemplateVarsMissingRequired(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "t.json")
	content := `{
  "id": "t",
  "name": "t",
  "mode": "manage",
  "vars": {
    "CUSTOM": { "type": "string", "required": true }
  },
  "stages": {
    "A": {
      "checks": [
        { "id": "a.system_info", "kind": "system_info", "params": { "note": "${CUSTOM}" } }
      ]
    }
  }
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	if _, _, err := Resolve(ResolveOptions{TemplateRef: path, BaseDir: tmp}); err == nil {
		t.Fatalf("expected missing required var error")
	}
}
