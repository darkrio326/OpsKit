package lock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	coreerr "opskit/internal/core/errors"
)

const staleLockMaxAge = 24 * time.Hour

type FileLock struct {
	path string
}

func Acquire(path string) (*FileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
		if stale, staleErr := isStale(path); staleErr == nil && stale {
			if rmErr := os.Remove(path); rmErr != nil && !os.IsNotExist(rmErr) {
				return nil, rmErr
			}
			f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err != nil {
				if os.IsExist(err) {
					return nil, coreerr.ErrLocked
				}
				return nil, err
			}
		} else {
			return nil, coreerr.ErrLocked
		}
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

func isStale(path string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) == 0 {
		return true, nil
	}
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err == nil && pid > 0 {
		alive, aliveErr := processAlive(pid)
		if aliveErr == nil {
			return !alive, nil
		}
	}

	if st, statErr := os.Stat(path); statErr == nil {
		return time.Since(st.ModTime()) > staleLockMaxAge, nil
	}
	return false, nil
}

func processAlive(pid int) (bool, error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return false, nil
		}
		var errno syscall.Errno
		if errors.As(err, &errno) {
			if errno == syscall.ESRCH {
				return false, nil
			}
			if errno == syscall.EPERM {
				return true, nil
			}
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "process already finished") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
