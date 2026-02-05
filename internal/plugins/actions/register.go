package actions

func RegisterBuiltins(r *Registry) {
	r.Register(func() Plugin { return &ensurePathsAction{} })
	r.Register(func() Plugin { return &baselineSnapshotAction{} })
	r.Register(func() Plugin { return &ensureUserGroupAction{} })
	r.Register(func() Plugin { return &ensureOwnershipAction{} })
	r.Register(func() Plugin { return &untarAction{} })
	r.Register(func() Plugin { return &sha256VerifyAction{} })
	r.Register(func() Plugin { return &renderTemplatesAction{} })
	r.Register(func() Plugin { return &systemdInstallUnitAction{} })
	r.Register(func() Plugin { return &systemdDaemonReloadAction{} })
	r.Register(func() Plugin { return &systemdEnableStartAction{} })
	r.Register(func() Plugin { return &systemdStartAction{} })
	r.Register(func() Plugin { return &systemdStopAction{} })
	r.Register(func() Plugin { return &captureInventoryAction{} })
	r.Register(func() Plugin { return &declareStackAction{} })
	r.Register(func() Plugin { return &recoverSequenceAction{} })
}
