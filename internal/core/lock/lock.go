package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	coreerr "opskit/internal/core/errors"
)

type FileLock struct {
	path string
}

func Acquire(path string) (*FileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, coreerr.ErrLocked
		}
		return nil, err
	}
	_, _ = f.WriteString(strconv.Itoa(os.Getpid()) + "\n" + time.Now().Format(time.RFC3339) + "\n")
	_ = f.Close()
	return &FileLock{path: path}, nil
}

func (l *FileLock) Release() error {
	if l == nil {
		return nil
	}
	if err := os.Remove(l.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("release lock: %w", err)
	}
	return nil
}
