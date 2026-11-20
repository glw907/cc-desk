//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// usageResetFixture installs the usage-reset fixture in the rig: a pass
// stopped at 2026-11-20T20:00Z on a limit whose text resets at 10:59pm
// Alaska time, which is 07:59Z on the 21st. The fixture names the checkout
// and the worktree as @CLONE@, so the test makes the worktree in the rig's
// clone and fills the name in.
func (r launchRig) usageResetFixture(t *testing.T) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "usage-reset", "releases", "proj.json"))
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(r.clone, ".claude", "worktrees", "a")
	git(t, r.clone, "worktree", "add", "--no-track", wt, "-b", "a", "origin/main")
	body := strings.ReplaceAll(string(b), "@CLONE@", r.clone)
	if err := os.MkdirAll(filepath.Join(r.state, "releases"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.state, "releases", "proj.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// relaunchesOf gives the relaunch entries of a pass's session record.
func relaunchesOf(t *testing.T, r launchRig, pass string) []map[string]any {
	t.Helper()
	var out []map[string]any
	rel, _ := r.sessionOf(t, pass)["relaunches"].([]any)
	for _, e := range rel {
		out = append(out, e.(map[string]any))
	}
	return out
}

func TestUsageReset(t *testing.T) {
	r := newLaunchRig(t)
	r.usageResetFixture(t)

	// 07:54Z is five minutes before the reset: the check waits.
	got := r.check(t, "2026-11-21T07:54:00Z")
	if !strings.Contains(got.stdout, `"decision":"wait_usage"`) {
		t.Errorf("check output %q lacks wait_usage", got.stdout)
	}
	if n := len(r.launches(t)); n != 0 {
		t.Errorf("%d launches before the reset, want 0", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "stopped" || p["stop_reason"] != "usage" {
		t.Errorf("pass before the reset = %v, want stopped on usage", p)
	}
	want := "usage limit holds the relaunch; waits for the reset at 2026-11-21T07:59:00Z"
	if outs := logOutcomes(t, r.state); !slices.Equal(outs, []string{want}) {
		t.Errorf("log outcomes = %v, want [%q]", outs, want)
	}

	// 08:04Z is five minutes after: the relaunch runs and is not counted.
	r.check(t, "2026-11-21T08:04:00Z")
	if n := len(r.launches(t)); n != 1 {
		t.Fatalf("%d launches after the reset, want 1", n)
	}
	if p := r.passRecord(t, "a"); p["state"] != "building" {
		t.Errorf("pass after the reset = %v, want building", p)
	}
	rel := relaunchesOf(t, r, "a")
	if len(rel) != 1 || rel[0]["cause"] != "usage" || rel[0]["counted"] == true {
		t.Errorf("relaunches = %v, want one usage relaunch, uncounted", rel)
	}
	if outs := logOutcomes(t, r.state); len(outs) != 2 || !strings.HasPrefix(outs[1], "launched ") {
		t.Errorf("log outcomes = %v, want the wait, then the launch", outs)
	}
}

func TestUsageNoTime(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	r.check(t, minutes(t, 1))
	r.check(t, minutes(t, 2)) // the new row works, so the check has seen it

	clock := 2
	next := func() string { clock++; return minutes(t, clock) }
	var counted []bool
	for n := 1; n <= 3; n++ {
		requireRun(t, r.do(t, next(), "pass", "stop", "proj", "a", "--reason", "usage"))
		r.check(t, next())
		if got := len(r.launches(t)); got != n+1 {
			t.Fatalf("after usage stop %d: %d launches, want %d", n, got, n+1)
		}
		rel := relaunchesOf(t, r, "a")
		if len(rel) != n || rel[n-1]["cause"] != "usage" {
			t.Fatalf("after usage stop %d the record holds %v", n, rel)
		}
		counted = append(counted, rel[n-1]["counted"] == true)
		r.check(t, next()) // the relaunched session works
	}
	if want := []bool{false, true, true}; !slices.Equal(counted, want) {
		t.Errorf("relaunches counted %v, want %v", counted, want)
	}

	requireRun(t, r.do(t, next(), "pass", "stop", "proj", "a", "--reason", "usage"))
	got := r.check(t, next())
	if n := len(r.launches(t)); n != 4 {
		t.Errorf("%d launches after the fourth stop, want 4 (no relaunch)", n)
	}
	p := r.passRecord(t, "a")
	if p["state"] != "stopped" || p["stop_reason"] != "relaunch_limit" {
		t.Errorf("pass = %v, want stopped with relaunch_limit", p)
	}
	if !strings.Contains(got.stdout, `"reason":"relaunch_limit"`) {
		t.Errorf("check output %q lacks the relaunch_limit stop", got.stdout)
	}
	open := 0
	stops, _ := p["stops"].([]any)
	for _, s := range stops {
		if _, closed := s.(map[string]any)["end"]; !closed {
			open++
		}
	}
	if open != 1 {
		t.Errorf("%d open stop intervals, want 1: %v", open, stops)
	}
	needs := r.needs(t)
	if len(needs) != 1 || needs[0]["reason"] != "relaunch_limit" {
		t.Errorf("needs_you_now = %v, want one relaunch_limit entry", needs)
	}
}

// TestUsageWithoutAPIKey runs the usage verb with ANTHROPIC_API_KEY set and a
// usage command that fails while it sees the key, as Claude Code bills the key
// ahead of the claude.ai login.
func TestUsageWithoutAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "leaked")
	dir := t.TempDir()
	resultFile := filepath.Join(dir, "result.json")
	if err := os.WriteFile(resultFile, []byte(usageResult(40)), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, "usage.sh")
	script := "#!/bin/sh\nif [ -n \"$ANTHROPIC_API_KEY\" ]; then echo 'Credit balance too low' >&2; exit 1; fi\ncat '" + resultFile + "'\n"
	if err := os.WriteFile(cmd, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	got := runWith(t, cmd, newState(t), "2026-11-20T20:00:00Z", "usage")
	if got.code != 0 || !strings.Contains(got.stdout, "40% used") {
		t.Errorf("usage = %d %q %q, want 40%% used", got.code, got.stdout, got.stderr)
	}
}
