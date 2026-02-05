package checks

func RegisterBuiltins(r *Registry) {
	r.Register(func() Plugin { return &systemInfoCheck{} })
	r.Register(func() Plugin { return &defaultRouteCheck{} })
	r.Register(func() Plugin { return &dnsConfigCheck{} })
	r.Register(func() Plugin { return &systemdAvailableCheck{} })
	r.Register(func() Plugin { return &systemdBasicsCheck{} })
	r.Register(func() Plugin { return &firewallStatusCheck{} })
	r.Register(func() Plugin { return &selinuxStatusCheck{} })
	r.Register(func() Plugin { return &timeSyncStatusCheck{} })
	r.Register(func() Plugin { return &timeDriftCheck{} })
	r.Register(func() Plugin { return &mountCheck{} })
	r.Register(func() Plugin { return &portConflictCheck{} })
	r.Register(func() Plugin { return &portListeningCheck{} })
	r.Register(func() Plugin { return &systemdUnitExistsCheck{} })
	r.Register(func() Plugin { return &systemdUnitActiveCheck{} })
	r.Register(func() Plugin { return &diskUsageCheck{} })
	r.Register(func() Plugin { return &memoryAvailableCheck{} })
	r.Register(func() Plugin { return &loadAverageCheck{} })
}
