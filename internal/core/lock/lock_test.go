package lock

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	coreerr "opskit/internal/core/errors"
)

func TestAcquireRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "opskit.lock")
	l, err := Acquire(path)
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release lock: %v", err)
	}
}

func TestAcquireReturnsLockedForLiveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "opskit.lock")
	l, err := Acquire(path)
	if err != nil {
		t.Fatalf("acquire first lock: %v", err)
	}
	defer func() { _ = l.Release() }()

	_, err = Acquire(path)
	if !errors.Is(err, coreerr.ErrLocked) {
		t.Fatalf("expected ErrLocked, got %v", err)
	}
}

func TestAcquireRecoversStaleLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "opskit.lock")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir lock dir: %v", err)
	}
	stalePID := strconv.Itoa(999999)
	staleTime := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
	if err := os.WriteFile(path, []byte(stalePID+"\n"+staleTime+"\n"), 0o644); err != nil {
		t.Fatalf("write stale lock: %v", err)
	}

	l, err := Acquire(path)
	if err != nil {
		t.Fatalf("acquire should recover stale lock: %v", err)
	}
	defer func() { _ = l.Release() }()
}
