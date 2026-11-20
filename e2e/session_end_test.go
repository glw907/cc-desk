//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

// sessionEnd runs `session end` on the rig's state folder with the stand-ins.
func (r launchRig) sessionEnd(t *testing.T, verb ...string) result {
	t.Helper()
	return r.run(t, workedClock, append([]string{"session", "end"}, verb...)...)
}

func (r launchRig) stopCalls(t *testing.T) int {
	t.Helper()
	return strings.Count(strings.Join(r.kinds(t), " "), "stop")
}

func TestSessionEndStopsTheSession(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(r launchRig)
		want    string
		stops   int
	}{
		{"a working row", func(r launchRig) {}, "stopped session b0000001", 1},
		{"a stopped row", func(r launchRig) { r.markRows(t, "stopped") }, "session b0000001 already gone", 0},
		{"a missing row", func(r launchRig) { r.vanish(t, "b0000001") }, "session b0000001 already gone", 0},
		{"a done row whose stop fails", func(r launchRig) {
			r.markRows(t, "done")
			r.control(t, "stop-fails", true)
		}, "session b0000001 stop failed, row stays listed done", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := launchedRig(t)
			tt.prepare(r)
			before := readFileIn(t, r.state, "releases/proj.json")
			got := r.sessionEnd(t, "pass", "proj", "a")
			requireRun(t, got)
			if strings.TrimSpace(got.stdout) != tt.want {
				t.Errorf("stdout = %q, want %q", got.stdout, tt.want)
			}
			if n := r.stopCalls(t); n != tt.stops {
				t.Errorf("stop calls = %d, want %d", n, tt.stops)
			}
			if after := readFileIn(t, r.state, "releases/proj.json"); after != before {
				t.Error("the verb changed the release file")
			}
		})
	}
}

func TestSessionEndConductor(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.run(t, workedClock, "session", "start", "conductor", "proj"))
	got := r.sessionEnd(t, "conductor", "proj")
	requireRun(t, got)
	if strings.TrimSpace(got.stdout) != "stopped session b0000001" {
		t.Errorf("stdout = %q", got.stdout)
	}
	if rows := r.readRows(t); rows[0]["state"] != "stopped" {
		t.Errorf("row = %v, want stopped", rows[0])
	}
}

func TestSessionEndRefuses(t *testing.T) {
	tests := []struct {
		name string
		rig  func(t *testing.T) launchRig
		args []string
		want string
	}{
		{"a pass never launched", func(t *testing.T) launchRig {
			r := newLaunchRig(t)
			r.addRelease(t, true, "a")
			return r
		}, []string{"pass", "proj", "a"}, "proj/pass:a has no open session"},
		{"a conductor never launched", func(t *testing.T) launchRig {
			r := newLaunchRig(t)
			r.addRelease(t, true, "a")
			return r
		}, []string{"conductor", "proj"}, "proj/conductor has no open session"},
		{"an unknown kind", newLaunchRig, []string{"design", "proj", "r1"}, "unknown session kind"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.rig(t)
			stops := r.stopCalls(t)
			got := r.sessionEnd(t, tt.args...)
			if got.code == 0 || !strings.Contains(got.stderr, tt.want) {
				t.Fatalf("exit %d, stderr %q, want a refusal naming %q", got.code, got.stderr, tt.want)
			}
			if n := r.stopCalls(t); n != stops {
				t.Errorf("a refused verb ran %d stop calls", n-stops)
			}
		})
	}
}

func TestSessionEndWaitsForTheCheckLock(t *testing.T) {
	r := launchedRig(t)
	f, err := os.OpenFile(filepath.Join(r.state, "check.lock"), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}

	done := make(chan result, 1)
	go func() { done <- r.sessionEnd(t, "pass", "proj", "a") }()
	select {
	case got := <-done:
		t.Fatalf("the verb ran while check.lock was held: exit %d, stderr %s", got.code, got.stderr)
	case <-time.After(time.Second):
	}
	if n := r.stopCalls(t); n != 0 {
		t.Fatalf("%d stop calls while check.lock was held", n)
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

// A session the verb ended is read like any other stopped row: once a check
// has seen it, the next run counts a missing relaunch.
func TestSessionEndThenCheckReadsTheDeath(t *testing.T) {
	r := launchedRig(t)
	requireRun(t, r.run(t, minutes(t, 1), "check"))
	requireRun(t, r.sessionEnd(t, "pass", "proj", "a"))

	got := r.check(t, minutes(t, 6))
	if !strings.Contains(got.stdout, `"decision":"relaunch"`) || !strings.Contains(got.stdout, `"reason":"missing"`) {
		t.Errorf("check output %q lacks a counted missing relaunch", got.stdout)
	}
	rel, _ := r.sessionOf(t, "a")["relaunches"].([]any)
	if len(rel) != 1 {
		t.Fatalf("relaunches = %v, want one", rel)
	}
	if got := startLogLines(t, r.state, "pass:a"); !slices.ContainsFunc(got, func(l [2]string) bool { return l[0] == "relaunch" }) {
		t.Errorf("the check log lines for pass:a = %v, want a relaunch", got)
	}
}
