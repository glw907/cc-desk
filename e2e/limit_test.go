//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCallLimits runs the merge call, the fetch, and the pull against
// stand-ins whose child sleep holds stdout, with --call-limit 2s. Each run
// returns well inside the 60-second limit, names the limit, and leaves no
// child behind.
func TestCallLimits(t *testing.T) {
	// run reports one pass ready, then runs the check with the sleeping
	// stand-ins and the control the case sets.
	run := func(t *testing.T, pass, control, verb string) (m mergeRig, got result, took time.Duration) {
		t.Helper()
		m = newMergeRig(t, pass)
		head := m.push(t, pass, pass+".txt")
		m.report(t, pass, 1, head, "inside")
		m.answer(t, pass, 1, head, passing)
		m.rowsDone(t)
		gh, bin := m.sleepers(t, m.gh)
		m.write(t, control, verb)
		args := append(m.args(t, "check"), "--gh-command", gh, "--call-limit", "2s")
		start := time.Now()
		got = runEnv(t, m.state, mergeClock, []string{"PATH=" + bin + ":/usr/bin:/bin"}, args...)
		return m, got, time.Since(start)
	}
	// childGone fails when the pid the stand-in recorded still has a /proc
	// entry.
	childGone := func(t *testing.T, m mergeRig, file string) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(m.dir, file))
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

	t.Run("the merge call", func(t *testing.T) {
		m, got, took := run(t, "lm", "gh-sleeps", "")
		if took > 6*time.Second {
			t.Errorf("the run took %v, want under 6s", took)
		}
		if out := lastOutcome(t, m.state); !strings.Contains(out, "merge call failed (1 of 3): stopped at the 2s limit") {
			t.Errorf("outcome = %q, want it to name the limit\n%s%s", out, got.stdout, got.stderr)
		}
		childGone(t, m, "gh-child.pid")
	})
	t.Run("the fetch", func(t *testing.T) {
		m, got, took := run(t, "lf", "git-sleeps", "fetch")
		if took > 6*time.Second {
			t.Errorf("the run took %v, want under 6s", took)
		}
		if got.code == 0 {
			t.Errorf("the check exited 0, want the cleanup's error\n%s%s", got.stdout, got.stderr)
		}
		if out := lastOutcome(t, m.state); !strings.Contains(out, "stopped at the 2s limit") {
			t.Errorf("outcome = %q, want it to name the limit", out)
		}
		childGone(t, m, "git-child.pid")
	})
	t.Run("the pull", func(t *testing.T) {
		m, got, took := run(t, "lp", "git-sleeps", "pull")
		if took > 8*time.Second {
			t.Errorf("the run took %v, want under 8s", took)
		}
		out := lastOutcome(t, m.state)
		if !strings.Contains(out, "checkout not pulled: ") || !strings.Contains(out, "stopped at the 2s limit") {
			t.Errorf("outcome = %q, want a note that names the limit\n%s%s", out, got.stdout, got.stderr)
		}
		childGone(t, m, "git-child.pid")
	})
}
