package templates

import (
	"sort"

	"opskit/internal/schema"
)

type Plan struct {
	TemplateID string
	Stages     []StagePlan
}

type StagePlan struct {
	StageID  string
	Checks   []schema.TemplateStep
	Actions  []schema.TemplateStep
	Evidence []schema.TemplateStep
}

func BuildPlan(t schema.Template, selected []string) Plan {
	return BuildPlanWithOptions(t, selected, false)
}

func BuildPlanWithOptions(t schema.Template, selected []string, includeDisabled bool) Plan {
	selectedSet := map[string]bool{}
	for _, s := range selected {
		selectedSet[s] = true
	}
	stages := []StagePlan{}
	ids := make([]string, 0, len(t.Stages))
	for id := range t.Stages {
		if len(selectedSet) > 0 && !selectedSet[id] {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return stageOrder(ids[i]) < stageOrder(ids[j]) })
	for _, id := range ids {
		spec := t.Stages[id]
		stages = append(stages, StagePlan{
			StageID:  id,
			Checks:   enabled(spec.Checks, includeDisabled),
			Actions:  enabled(spec.Actions, includeDisabled),
			Evidence: enabled(spec.Evidence, includeDisabled),
		})
	}
	return Plan{TemplateID: t.ID, Stages: stages}
}

func enabled(steps []schema.TemplateStep, includeDisabled bool) []schema.TemplateStep {
	out := make([]schema.TemplateStep, 0, len(steps))
	for _, s := range steps {
		if !includeDisabled && s.Enabled != nil && !*s.Enabled {
			continue
		}
		out = append(out, s)
	}
	return out
}

func stageOrder(id string) int {
	switch id {
	case "A":
		return 1
	case "B":
		return 2
	case "C":
		return 3
	case "D":
		return 4
	case "E":
		return 5
	case "F":
		return 6
	default:
		return 99
	}
}
