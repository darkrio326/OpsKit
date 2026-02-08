package schema

type Template struct {
	ID     string                       `json:"id"`
	Name   string                       `json:"name"`
	Mode   string                       `json:"mode"`
	Vars   map[string]VarSpec           `json:"vars,omitempty"`
	Stages map[string]TemplateStageSpec `json:"stages"`
}

type VarSpec struct {
	Type        string   `json:"type,omitempty"`
	Group       string   `json:"group,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Default     string   `json:"default,omitempty"`
	Example     string   `json:"example,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
}

type TemplateStageSpec struct {
	Checks   []TemplateStep `json:"checks,omitempty"`
	Actions  []TemplateStep `json:"actions,omitempty"`
	Evidence []TemplateStep `json:"evidence,omitempty"`
}

type TemplateStep struct {
	ID      string         `json:"id"`
	Kind    string         `json:"kind"`
	Params  map[string]any `json:"params,omitempty"`
	Enabled *bool          `json:"enabled,omitempty"`
}
