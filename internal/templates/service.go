package templates

import (
	"opskit/internal/schema"
)

type ResolveOptions struct {
	TemplateRef string
	BaseDir     string
	VarsRaw     string
}

func Resolve(opt ResolveOptions) (schema.Template, map[string]string, error) {
	t, err := Load(opt.TemplateRef)
	if err != nil {
		return schema.Template{}, nil, err
	}
	defaults := DefaultVars(opt.BaseDir)
	overrides := ParseVars(opt.VarsRaw)
	vars := MergeVars(defaults, overrides)
	t = ApplyVars(t, vars)
	if err := schema.ValidateTemplate(t); err != nil {
		return schema.Template{}, nil, err
	}
	return t, vars, nil
}
