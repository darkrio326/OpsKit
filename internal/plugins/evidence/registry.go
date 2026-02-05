package evidence

import (
	"context"
	"fmt"

	"opskit/internal/core/executil"
	"opskit/internal/schema"
)

type Request struct {
	ID     string
	Params map[string]any
	Exec   executil.Runner
}

type Result struct {
	EvidenceID string
	Status     schema.Status
	Severity   schema.Severity
	Message    string
	Path       string
	Metrics    []schema.Metric
	Issue      *schema.Issue
}

type Plugin interface {
	Kind() string
	Collect(ctx context.Context, req Request) (Result, error)
}

type Factory func() Plugin

type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]Factory{}}
}

func (r *Registry) Register(factory Factory) {
	p := factory()
	r.factories[p.Kind()] = factory
}

func (r *Registry) MustPlugin(kind string) (Plugin, error) {
	f, ok := r.factories[kind]
	if !ok {
		return nil, fmt.Errorf("evidence plugin not found: %s", kind)
	}
	return f(), nil
}
