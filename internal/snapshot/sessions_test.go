package snapshot

import (
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

func sessionOn(kind state.SessionKind, name string) *state.Session {
	return &state.Session{
		Name: name, Kind: kind, ShortID: "ab12cd34", SessionID: "ab12cd34-0000", Worktree: "/w/" + name,
		Branch: name, Launched: clock.Add(-time.Hour), LaunchTries: 1,
	}
}

func needsState(t *testing.T, snap *Snapshot) string {
	t.Helper()
	if snap.NextSession.State == nil {
		t.Fatal("next_session.state is null")
	}
	return *snap.NextSession.State
}

func TestNeedsYouNowStoppedConductor(t *testing.T) {
	rf := asking("a", 0, true)
	c := sessionOn(state.KindConductor, "a-conductor")
	stopped := clock.Add(-time.Minute)
	c.StopReason, c.Stopped, c.StopDetail = state.StopPermissionPrompt, &stopped, "Bash: rm -r build"
	rf.Conductor = c
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)

	if got := needsState(t, snap); got != "needs_you_now" {
		t.Errorf("next_session.state = %s, want needs_you_now", got)
	}
	if len(snap.NeedsYouNow) != 1 {
		t.Fatalf("needs_you_now = %+v, want one entry", snap.NeedsYouNow)
	}
	e := snap.NeedsYouNow[0]
	if e.Round != nil || e.Session == nil || *e.Session != "a-conductor" || e.Reason != "permission_prompt" ||
		e.Cause != "Bash: rm -r build" || e.Todo != "answer it at the desk" {
		t.Errorf("entry = %+v", e)
	}
	if len(snap.Sessions) != 1 || snap.Sessions[0].Pass != nil || snap.Sessions[0].StopReason == nil {
		t.Errorf("sessions = %+v", snap.Sessions)
	}
}

func TestNeedsYouNowStoppedPassTakesItsCauseFromStopDetail(t *testing.T) {
	now := clock
	rf := asking("a", 0, true)
	b := building("w", "r1", 10, now.Add(-20*time.Minute))
	b.State, b.StopReason = state.Stopped, state.StopUnreadable
	b.Stops = []state.StopInterval{{Start: now.Add(-10 * time.Minute)}}
	b.Session = sessionOn(state.KindPass, "a-w")
	b.Session.StopDetail = "the last line of the screen"
	rf.Passes = append(rf.Passes, b)
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now)

	if len(snap.NeedsYouNow) != 1 {
		t.Fatalf("needs_you_now = %+v, want one entry", snap.NeedsYouNow)
	}
	e := snap.NeedsYouNow[0]
	if e.Cause != "the last line of the screen" || e.Reason != "unreadable_stop" ||
		e.Todo != "read the session's log, then start the pass again" || e.Round == nil || *e.Round != "r1" {
		t.Errorf("entry = %+v", e)
	}
	if s := snap.Sessions[len(snap.Sessions)-1]; s.StopReason == nil || *s.StopReason != "unreadable_stop" || s.Pass == nil || *s.Pass != "w" {
		t.Errorf("session entry = %+v", s)
	}
}

func TestNeedsYouNowOlderStopHasNoSession(t *testing.T) {
	now := clock
	rf := asking("a", 0, true)
	b := building("w", "r1", 10, now.Add(-20*time.Minute))
	b.State, b.StopReason = state.Stopped, state.StopUsage
	b.Stops = []state.StopInterval{{Start: now.Add(-10 * time.Minute)}}
	rf.Passes = append(rf.Passes, b)
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now)

	if len(snap.NeedsYouNow) != 1 {
		t.Fatalf("needs_you_now = %+v, want one entry", snap.NeedsYouNow)
	}
	if e := snap.NeedsYouNow[0]; e.Session != nil || e.Reason != "usage" || e.Todo != "wait for the reset, or reset early" {
		t.Errorf("entry = %+v", e)
	}
}

func TestNeedsYouNowListFailures(t *testing.T) {
	rf := asking("a", 0, true)
	for n, want := range map[int]string{2: "nothing_waiting", 3: "needs_you_now"} {
		st := &state.State{Releases: []*state.ReleaseFile{rf}, Check: state.CheckFile{Format: 1, ListFailures: n, Decisions: []state.Decision{}}}
		snap := Build(st, clock)
		if got := needsState(t, snap); got != want {
			t.Errorf("%d failures: next_session.state = %s, want %s", n, got, want)
		}
		if n < 3 {
			if len(snap.NeedsYouNow) != 0 {
				t.Errorf("%d failures: needs_you_now = %+v, want none", n, snap.NeedsYouNow)
			}
			continue
		}
		if len(snap.NeedsYouNow) != 1 {
			t.Fatalf("needs_you_now = %+v, want one entry", snap.NeedsYouNow)
		}
		e := snap.NeedsYouNow[0]
		if e.Reason != ListUnreadable || e.Project != "" || e.Round != nil || e.Session != nil || e.Todo == "" || e.Cause == "" {
			t.Errorf("entry = %+v", e)
		}
	}
}

func TestLastCheck(t *testing.T) {
	if got := Build(&state.State{}, clock).LastCheck; got != nil {
		t.Errorf("last_check = %+v, want null before any run", got)
	}
	run := clock.Add(-time.Minute)
	d := state.Decision{Project: "a", Record: "pass:w", Session: "a-w", Decision: "none"}
	st := &state.State{Check: state.CheckFile{Format: 1, ListFailures: 1, LastRun: &run, Decisions: []state.Decision{d}}}
	got := Build(st, clock).LastCheck
	if got == nil || !got.Time.Equal(run) || got.ListFailures != 1 || len(got.Decisions) != 1 {
		t.Errorf("last_check = %+v", got)
	}
}
