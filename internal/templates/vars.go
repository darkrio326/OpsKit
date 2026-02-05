package templates

import (
	"fmt"
	"strings"

	"opskit/internal/schema"
)

func ParseVars(raw string) map[string]string {
	out := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return out
}

func DefaultVars(base string) map[string]string {
	return map[string]string{
		"INSTALL_ROOT":         base,
		"CONF_DIR":             fmt.Sprintf("%s/conf", base),
		"EVIDENCE_DIR":         fmt.Sprintf("%s/evidence", base),
		"SERVICE_NAME":         "hello-service",
		"SERVICE_UNIT":         "hello-service.service",
		"SERVICE_PORT":         "18080",
		"SERVICE_EXEC":         fmt.Sprintf("%s/hello-service/release/hello-service.sh", base),
		"PROCESS_MATCH":        "hello-service",
		"SYSTEMD_UNIT_DIR":     "/etc/systemd/system",
		"PACKAGE_FILE":         fmt.Sprintf("%s/packages/hello-service.tar.gz", base),
		"PACKAGE_SHA256":       "replace-with-real-sha256",
		"INVENTORY_FILE":       fmt.Sprintf("%s/evidence/inventory.json", base),
		"STACK_DECLARATION":    fmt.Sprintf("%s/state/stack.json", base),
		"BASELINE_SNAPSHOT":    fmt.Sprintf("%s/evidence/baseline-snapshot.json", base),
		"RECOVER_CIRCUIT_FILE": fmt.Sprintf("%s/state/recover_circuit.json", base),
		"RECOVER_COLLECT_FILE": fmt.Sprintf("%s/evidence/recover-collect.json", base),
		"COLLECT_BUNDLE_DIR":   fmt.Sprintf("%s/bundles", base),
		"RECOVER_TRIGGER":      "manual",
	}
}

func MergeVars(defaults map[string]string, overrides map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range defaults {
		out[k] = v
	}
	for k, v := range overrides {
		out[k] = v
	}
	return out
}

func ApplyVars(t schema.Template, vars map[string]string) schema.Template {
	out := t
	for stageID, stage := range out.Stages {
		for i := range stage.Checks {
			stage.Checks[i].Params = replaceMap(stage.Checks[i].Params, vars)
		}
		for i := range stage.Actions {
			stage.Actions[i].Params = replaceMap(stage.Actions[i].Params, vars)
		}
		for i := range stage.Evidence {
			stage.Evidence[i].Params = replaceMap(stage.Evidence[i].Params, vars)
		}
		out.Stages[stageID] = stage
	}
	return out
}

func replaceMap(in map[string]any, vars map[string]string) map[string]any {
	if in == nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range in {
		out[k] = replaceValue(v, vars)
	}
	return out
}

func replaceValue(v any, vars map[string]string) any {
	switch vv := v.(type) {
	case string:
		for k, val := range vars {
			vv = strings.ReplaceAll(vv, "${"+k+"}", val)
		}
		return vv
	case []any:
		out := make([]any, 0, len(vv))
		for _, x := range vv {
			out = append(out, replaceValue(x, vars))
		}
		return out
	case map[string]any:
		return replaceMap(vv, vars)
	default:
		return v
	}
}
