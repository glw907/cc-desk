//go:build unix

package state

import (
	"strings"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	dir := t.TempDir()
	unlock, err := Lock(dir, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	_, err = Lock(dir, 200*time.Millisecond)
	if err == nil {
		t.Fatal("a held lock was taken twice")
	}
	if !strings.Contains(err.Error(), dir) {
		t.Errorf("error %q does not name the folder", err)
	}
	if time.Since(start) < 200*time.Millisecond {
		t.Errorf("gave up after %v, before the wait ended", time.Since(start))
	}

	unlock()
	again, err := Lock(dir, time.Second)
	if err != nil {
		t.Fatalf("lock after release: %v", err)
	}
	again()
}

func TestLockWaitsForRelease(t *testing.T) {
	dir := t.TempDir()
	unlock, err := Lock(dir, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	time.AfterFunc(150*time.Millisecond, unlock)
	second, err := Lock(dir, 3*time.Second)
	if err != nil {
		t.Fatalf("second command did not get the lock after release: %v", err)
	}
	second()
}
