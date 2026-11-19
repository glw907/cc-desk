package launch

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// slowGit puts a git on PATH that starts a child sleep holding stdout when
// the call's arguments contain verb, and passes every other call to the real
// git. It returns the file the child's pid goes to.
func slowGit(t *testing.T, verb string) string {
	t.Helper()
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	script := "#!/bin/sh\n" +
		"for a in \"$@\"; do\n" +
		"  if [ \"$a\" = " + verb + " ]; then sleep 120 & echo $! > " + pidFile + "; wait; exit 1; fi\n" +
		"done\n" +
		"exec " + real + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return pidFile
}

// childGone fails the test when the pid the stand-in recorded still has a
// /proc entry.
func childGone(t *testing.T, pidFile string) {
	t.Helper()
	b, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("the stand-in recorded no child: %v", err)
	}
	pid := strings.TrimSpace(string(b))
	// The killed child stays in /proc until init reaps it.
	for range 20 {
		if _, err := os.Stat("/proc/" + pid); err != nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("the child %s is still running", pid)
}

func TestPullFetchStopsAtTheLimit(t *testing.T) {
	checkout := repoWithRemote(t)
	pidFile := slowGit(t, "fetch")
	start := time.Now()
	_, _, err := Worktree{Checkout: checkout, Default: "main", Limit: time.Second}.Pull(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "stopped at the 1s limit") {
		t.Fatalf("err = %v, want one naming the limit", err)
	}
	if d := time.Since(start); d > 4*time.Second {
		t.Errorf("returned after %v", d)
	}
	childGone(t, pidFile)
}

func TestPullPullStopsAtTheLimit(t *testing.T) {
	checkout := repoWithRemote(t)
	pidFile := slowGit(t, "pull")
	pulled, note, err := Worktree{Checkout: checkout, Default: "main", Limit: time.Second}.Pull(context.Background(), nil)
	if err != nil || pulled || !strings.Contains(note, "stopped at the 1s limit") {
		t.Fatalf("pulled = %v, note = %q, err = %v", pulled, note, err)
	}
	childGone(t, pidFile)
}

func TestPullCancelledContextIsNotTheLimit(t *testing.T) {
	checkout := repoWithRemote(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := Worktree{Checkout: checkout, Default: "main"}.Pull(ctx, nil)
	if err == nil || strings.Contains(err.Error(), "limit") {
		t.Errorf("err = %v, want a plain cancellation", err)
	}
}
