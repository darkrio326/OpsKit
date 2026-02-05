package evidence

func RegisterBuiltins(r *Registry) {
	r.Register(func() Plugin { return &fileHashEvidence{} })
	r.Register(func() Plugin { return &commandOutputEvidence{} })
	r.Register(func() Plugin { return &dirHashEvidence{} })
	r.Register(func() Plugin { return &processArgsEvidence{} })
}
