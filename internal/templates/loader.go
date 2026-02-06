package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	return decodeTemplateStrict(b)
}

func loadFile(path string) (schema.Template, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return schema.Template{}, err
	}
	return decodeTemplateStrict(b)
}

func decodeTemplateStrict(b []byte) (schema.Template, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var t schema.Template
	if err := dec.Decode(&t); err != nil {
		return schema.Template{}, fmt.Errorf("template: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return schema.Template{}, fmt.Errorf("template: unexpected extra JSON content")
	}
	return t, nil
}
