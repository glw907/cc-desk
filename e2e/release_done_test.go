//go:build unix

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// releaseDone runs `release done proj` on the rig's state folder, with the
// stand-in claude for the session list.
func (r launchRig) releaseDone(t *testing.T) result {
	t.Helper()
	return run(t, r.state, workedClock, "release", "done", "proj", "--claude-command", r.claude)
}

// launchedRig gives a rig whose release holds pass a with a launched session,
// listed as a working row.
func launchedRig(t *testing.T) launchRig {
	t.Helper()
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	requireRun(t, r.run(t, workedClock, "check"))
	return r
}

// markRows sets every listed row's state.
func (r launchRig) markRows(t *testing.T, state string) {
	t.Helper()
	rows := r.readRows(t)
	for _, row := range rows {
		row["state"], row["status"] = state, state
	}
	r.setRows(t, rows...)
}

func TestReleaseDoneThenANewRelease(t *testing.T) {
	r := launchedRig(t)
	r.markRows(t, "done")

	got := r.releaseDone(t)
	requireRun(t, got)
	want := filepath.Join(r.state, "releases", "done", "v1-20261120T230000Z.json")
	if strings.TrimSpace(got.stdout) != want {
		t.Errorf("stdout = %q, want %q", got.stdout, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("the done file is missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.state, "releases", "proj.json")); !os.IsNotExist(err) {
		t.Errorf("releases/proj.json survived: %v", err)
	}
	var snap struct{ Releases []struct{ Project string } }
	if err := json.Unmarshal([]byte(readFileIn(t, r.state, "snapshot.json")), &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Releases) != 0 {
		t.Errorf("the snapshot lists %+v as live", snap.Releases)
	}

	r.addRelease(t, false, "b")
	if !strings.Contains(readFileIn(t, r.state, "releases/proj.json"), `"name": "b"`) {
		t.Error("the new release did not install")
	}
}

func TestReleaseDoneRefuses(t *testing.T) {
	tests := []struct {
		name  string
		rig   func(t *testing.T) launchRig
		state string
		want  string
	}{
		{"no live release", func(t *testing.T) launchRig {
			r := newLaunchRig(t)
			if err := os.MkdirAll(r.state, 0o755); err != nil {
				t.Fatal(err)
			}
			return r
		}, "", "no live release"},
		{"a working row", launchedRig, "working", "pass:a"},
		{"a blocked row", launchedRig, "blocked", "pass:a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.rig(t)
			if tt.state != "" {
				r.markRows(t, tt.state)
			}
			got := r.releaseDone(t)
			if got.code == 0 || !strings.Contains(got.stderr, tt.want) {
				t.Fatalf("exit %d, stderr %q, want a refusal naming %q", got.code, got.stderr, tt.want)
			}
			if tt.state != "" {
				if _, err := os.Stat(filepath.Join(r.state, "releases", "proj.json")); err != nil {
					t.Errorf("a refused verb moved the release: %v", err)
				}
			}
		})
	}

	// Stopped, done, and failed rows, and a missing row, do not hold a release.
	for _, state := range []string{"stopped", "done", "failed"} {
		t.Run("a "+state+" row", func(t *testing.T) {
			r := launchedRig(t)
			r.markRows(t, state)
			requireRun(t, r.releaseDone(t))
		})
	}
	t.Run("no row", func(t *testing.T) {
		r := launchedRig(t)
		r.vanish(t, "b0000001")
		requireRun(t, r.releaseDone(t))
	})
}

func TestReleaseDoneWaitsForTheCheckLock(t *testing.T) {
	r := launchedRig(t)
	r.markRows(t, "done")

	f, err := os.OpenFile(filepath.Join(r.state, "check.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}

	done := make(chan result, 1)
	go func() { done <- r.releaseDone(t) }()
	select {
	case got := <-done:
		t.Fatalf("the verb ran while check.lock was held: exit %d, stderr %s", got.code, got.stderr)
	case <-time.After(time.Second):
	}
	if _, err := os.Stat(filepath.Join(r.state, "releases", "proj.json")); err != nil {
		t.Fatalf("the release moved while check.lock was held: %v", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-done:
		requireRun(t, got)
	case <-time.After(10 * time.Second):
		t.Fatal("the verb did not run after check.lock was released")
	}
}
