//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	dir := newState(t)
	before := folder(t, dir)
	requireSnapshot(t, before)

	f, err := os.OpenFile(filepath.Join(dir, "lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	got := run(t, dir, workedClock, "pass", "start", "alpha", "p13")
	elapsed := time.Since(start)

	if got.code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", got.code, got.stderr)
	}
	if elapsed < 10*time.Second || elapsed >= 15*time.Second {
		t.Errorf("waited %v, want at least 10 and under 15 seconds", elapsed)
	}
	sameFolder(t, before, folder(t, dir))
}
