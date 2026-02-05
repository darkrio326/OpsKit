package schema

type Template struct {
	ID     string                       `json:"id"`
	Name   string                       `json:"name"`
	Mode   string                       `json:"mode"`
	Stages map[string]TemplateStageSpec `json:"stages"`
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
