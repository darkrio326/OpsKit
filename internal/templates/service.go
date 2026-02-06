package templates

import (
	"strings"

	"opskit/internal/schema"
)

type ResolveOptions struct {
	TemplateRef string
	BaseDir     string
	VarsRaw     string
	VarsFile    string
}

func Resolve(opt ResolveOptions) (schema.Template, map[string]string, error) {
	t, err := Load(opt.TemplateRef)
	if err != nil {
		return schema.Template{}, nil, err
	}
	defaults := DefaultVars(opt.BaseDir)
	vars := ApplyVarDefaults(defaults, t.Vars)
	if strings.TrimSpace(opt.VarsFile) != "" {
		fromFile, err := ParseVarsFile(opt.VarsFile)
		if err != nil {
			return schema.Template{}, nil, err
		}
		vars = MergeVars(vars, fromFile)
	}
	overrides := ParseVars(opt.VarsRaw)
	vars = MergeVars(vars, overrides)
	if err := schema.ValidateVars(t.Vars, vars); err != nil {
		return schema.Template{}, nil, err
	}
	t = ApplyVars(t, vars)
	if err := schema.ValidateTemplate(t); err != nil {
		return schema.Template{}, nil, err
	}
	return t, vars, nil
}
