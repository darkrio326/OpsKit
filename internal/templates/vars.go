package templates

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

func ParseVarsFile(path string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(path) == "" {
		return out, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	content := strings.TrimSpace(string(b))
	if content == "" {
		return out, nil
	}
	if strings.HasPrefix(content, "{") {
		var payload map[string]any
		if err := json.Unmarshal([]byte(content), &payload); err != nil {
			return nil, fmt.Errorf("vars-file json parse failed: %w", err)
		}
		for k, v := range payload {
			encoded, encErr := encodeVarValue(v)
			if encErr != nil {
				return nil, fmt.Errorf("vars-file encode %s: %w", k, encErr)
			}
			out[k] = encoded
		}
		return out, nil
	}
	if strings.HasPrefix(content, "[") {
		return nil, fmt.Errorf("vars-file must be a JSON object (got array)")
	}
	return parseVarsLines(content)
}

func parseVarsLines(content string) (map[string]string, error) {
	out := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("vars-file invalid line %d: %s", lineNo, line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("vars-file empty key at line %d", lineNo)
		}
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeVarValue(v any) (string, error) {
	switch val := v.(type) {
	case nil:
		return "", nil
	case string:
		return val, nil
	case bool:
		return strconv.FormatBool(val), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case []any, map[string]any:
		b, err := json.Marshal(val)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return fmt.Sprintf("%v", val), nil
	}
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

func ApplyVarDefaults(vars map[string]string, specs map[string]schema.VarSpec) map[string]string {
	if len(specs) == 0 {
		return vars
	}
	out := map[string]string{}
	for k, v := range vars {
		out[k] = v
	}
	for name, spec := range specs {
		if _, ok := out[name]; ok {
			continue
		}
		if strings.TrimSpace(spec.Default) != "" {
			out[name] = spec.Default
		}
	}
	return out
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
		if strings.HasPrefix(vv, "${") && strings.HasSuffix(vv, "}") {
			name := strings.TrimSuffix(strings.TrimPrefix(vv, "${"), "}")
			if val, ok := vars[name]; ok {
				if parsed, ok := decodeJSONVar(val); ok {
					return parsed
				}
				return val
			}
		}
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

func decodeJSONVar(val string) (any, bool) {
	raw := strings.TrimSpace(val)
	if raw == "" {
		return val, false
	}
	if !(strings.HasPrefix(raw, "{") || strings.HasPrefix(raw, "[")) {
		return val, false
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return val, false
	}
	switch out.(type) {
	case []any, map[string]any:
		return out, true
	default:
		return val, false
	}
}
