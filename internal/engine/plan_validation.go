package engine

import (
	"fmt"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/plugins/actions"
	"opskit/internal/plugins/checks"
	"opskit/internal/plugins/evidence"
	"opskit/internal/templates"
)

// ValidatePlanPlugins ensures every referenced step kind can be resolved by the active registries.
func ValidatePlanPlugins(plan templates.Plan, checkReg *checks.Registry, actionReg *actions.Registry, evidenceReg *evidence.Registry) error {
	if checkReg == nil || actionReg == nil || evidenceReg == nil {
		return fmt.Errorf("%w: plugin registries are required", coreerr.ErrPreconditionFailed)
	}

	missing := []string{}
	for _, stage := range plan.Stages {
		for i, step := range stage.Checks {
			if _, err := checkReg.MustPlugin(step.Kind); err != nil {
				missing = append(missing, fmt.Sprintf("template.stages.%s.checks[%d].kind unsupported: %v", stage.StageID, i, err))
			}
		}
		for i, step := range stage.Actions {
			if _, err := actionReg.MustPlugin(step.Kind); err != nil {
				missing = append(missing, fmt.Sprintf("template.stages.%s.actions[%d].kind unsupported: %v", stage.StageID, i, err))
			}
		}
		for i, step := range stage.Evidence {
			if _, err := evidenceReg.MustPlugin(step.Kind); err != nil {
				missing = append(missing, fmt.Sprintf("template.stages.%s.evidence[%d].kind unsupported: %v", stage.StageID, i, err))
			}
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", coreerr.ErrPreconditionFailed, strings.Join(missing, "; "))
	}
	return nil
}
