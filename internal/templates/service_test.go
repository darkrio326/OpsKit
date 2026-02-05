package templates

import "testing"

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
