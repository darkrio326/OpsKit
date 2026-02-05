package fsx

import "os"

type PathSpec struct {
	Path  string
	Perm  os.FileMode
	UID   int
	GID   int
	Chown bool
}

func EnsurePaths(specs []PathSpec) error {
	for _, spec := range specs {
		if spec.Path == "" {
			continue
		}
		perm := spec.Perm
		if perm == 0 {
			perm = 0o755
		}
		if err := EnsureDir(spec.Path, perm); err != nil {
			return err
		}
		if err := Chmod(spec.Path, perm); err != nil {
			return err
		}
		if spec.Chown {
			if err := Chown(spec.Path, spec.UID, spec.GID); err != nil {
				return err
			}
		}
	}
	return nil
}

func Chmod(path string, perm os.FileMode) error {
	if perm == 0 {
		return nil
	}
	return os.Chmod(path, perm)
}

func Chown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}
