package templates

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"opskit/internal/schema"
)

//go:embed builtin/*.json
var builtinFS embed.FS

func Load(ref string) (schema.Template, error) {
	if ref == "" || ref == "generic-manage-v1" {
		return loadBuiltin("default-manage.json")
	}
	if ref == "single-service-deploy" || ref == "single-service-deploy-v1" {
		return loadBuiltin("single-service-deploy.json")
	}
	if strings.HasSuffix(ref, ".json") || strings.HasPrefix(ref, "/") || strings.Contains(ref, string(os.PathSeparator)) {
		return loadFile(ref)
	}
	return schema.Template{}, fmt.Errorf("unknown template id: %s", ref)
}

func loadBuiltin(name string) (schema.Template, error) {
	b, err := builtinFS.ReadFile("builtin/" + name)
	if err != nil {
		return schema.Template{}, err
	}
	var t schema.Template
	if err := json.Unmarshal(b, &t); err != nil {
		return schema.Template{}, err
	}
	return t, nil
}

func loadFile(path string) (schema.Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return schema.Template{}, err
	}
	var t schema.Template
	if err := json.Unmarshal(b, &t); err != nil {
		return schema.Template{}, err
	}
	return t, nil
}
