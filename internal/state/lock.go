//go:build unix

package state

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// LockWait is how long a command waits for the state lock.
const LockWait = 10 * time.Second

// Lock takes the exclusive lock on dir/lock, waiting up to wait for
// another command to release it. It returns the function that releases it.
// The lock file stays empty and stays on disk.
func Lock(dir string, wait time.Duration) (unlock func(), err error) {
	f, err := os.OpenFile(filepath.Join(dir, "lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock in %s: %w", dir, err)
	}
	deadline := time.Now().Add(wait)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, nil
		}
		if err != syscall.EWOULDBLOCK {
			f.Close()
			return nil, fmt.Errorf("lock %s: %w", dir, err)
		}
		if !time.Now().Before(deadline) {
			f.Close()
			return nil, fmt.Errorf("state folder %s is locked by another cc-desk command; waited %s", dir, wait)
		}
		time.Sleep(25 * time.Millisecond)
	}
}
