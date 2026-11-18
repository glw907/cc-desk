package e2e

import (
	"slices"
	"strings"
	"testing"
)

// TestGuard runs the pass 1 check fixture, which records no checkout, with
// the launch stand-ins. Every action is decided and none is carried out: no
// session starts or stops, systemd-run never runs, and no record gains a
// relaunch.
func TestGuard(t *testing.T) {
	r := newLaunchRig(t)
	copyTree(t, "testdata/check", r.state)

	got := r.run(t, workedClock, "check")
	if got.code != 0 {
		t.Fatalf("check exit %d, stderr: %s", got.code, got.stderr)
	}
	for _, k := range r.kinds(t) {
		if k == "--bg" || k == "stop" {
			t.Errorf("the check called claude %s with no checkout recorded", k)
		}
	}
	if s := r.scopeLog(t); s != "" {
		t.Errorf("systemd-run ran: %s", s)
	}
	release := readFileIn(t, r.state, "releases/chk.json")
	if strings.Contains(release, `"relaunches"`) {
		t.Errorf("a record gained a relaunch: %s", release)
	}
	decided := decisions(t, got.stdout)
	if len(decided) == 0 {
		t.Fatal("the check printed no decisions")
	}
	if !slices.Contains(r.kinds(t), "agents") {
		t.Error("the check never read the session list")
	}
}
