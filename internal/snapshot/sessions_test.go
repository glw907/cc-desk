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

func TestNeedsYouNowFailedLaunchShowsItsError(t *testing.T) {
	now := clock
	rf := asking("a", 0, true)
	b := building("w", "r1", 10, now.Add(-20*time.Minute))
	b.State, b.StopReason, b.StopDetail = state.Stopped, state.StopLaunchFailed, "claude --bg: exit status 1"
	b.Stops = []state.StopInterval{{Start: now.Add(-10 * time.Minute)}}
	rf.Passes = append(rf.Passes, b)
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now)

	if len(snap.NeedsYouNow) != 1 || snap.NeedsYouNow[0].Cause != "claude --bg: exit status 1" {
		t.Fatalf("needs_you_now = %+v, want the launch error as the cause", snap.NeedsYouNow)
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

func TestNoCheckEntry(t *testing.T) {
	for _, tc := range []struct {
		name string
		ago  time.Duration
		want bool
	}{{"29 minutes", 29 * time.Minute, false}, {"30 minutes", 30 * time.Minute, true}, {"31 minutes", 31 * time.Minute, true}} {
		t.Run(tc.name, func(t *testing.T) {
			ready := clock.Add(-tc.ago)
			b := building("w", "r1", 30, clock.Add(-2*time.Hour))
			b.Session = sessionOn(state.KindPass, "a-w")
			b.PullRequest = &state.PullRequest{Number: 7, Commit: "abc", Ready: &ready, Wait: state.WaitNoCheck}
			rf := releaseOf("a", []state.Round{{Name: "r1", DesignWritten: true}}, b)
			snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
			if !tc.want {
				if len(snap.NeedsYouNow) != 0 || needsState(t, snap) == "needs_you_now" {
					t.Errorf("entries = %+v, state = %s, want none", snap.NeedsYouNow, needsState(t, snap))
				}
				return
			}
			if len(snap.NeedsYouNow) != 1 {
				t.Fatalf("needs_you_now = %+v, want one entry", snap.NeedsYouNow)
			}
			e := snap.NeedsYouNow[0]
			if e.Reason != "no_check_registered" || e.Cause != "7" || e.Todo != "See why no check has started on pull request 7" ||
				e.Session == nil || *e.Session != "a-w" || e.Round == nil || *e.Round != "r1" {
				t.Errorf("entry = %+v", e)
			}
			if got := needsState(t, snap); got != "needs_you_now" {
				t.Errorf("next_session.state = %s, want needs_you_now", got)
			}
		})
	}
}

func TestNoCheckEntryNeedsTheNoCheckWait(t *testing.T) {
	ready := clock.Add(-time.Hour)
	b := building("w", "r1", 30, clock.Add(-2*time.Hour))
	b.PullRequest = &state.PullRequest{Number: 7, Commit: "abc", Ready: &ready, Wait: state.WaitCheckRunning}
	rf := releaseOf("a", []state.Round{{Name: "r1", DesignWritten: true}}, b)
	if snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock); len(snap.NeedsYouNow) != 0 {
		t.Errorf("needs_you_now = %+v, want none while a check runs", snap.NeedsYouNow)
	}
}

func TestCleanupNote(t *testing.T) {
	merged := clock.Add(-time.Hour)
	for _, tc := range []struct{ note, want string }{{"", ""}, {"branch left: not fully merged", "branch left: not fully merged"}} {
		b := building("w", "r1", 30, clock.Add(-3*time.Hour))
		minutes := 120
		b.State, b.Finished, b.ActualMinutes = state.Done, &merged, &minutes
		b.Session = sessionOn(state.KindPass, "a-w")
		b.PullRequest = &state.PullRequest{Number: 7, Merged: &merged, CleanupNote: tc.note}
		rf := releaseOf("a", []state.Round{{Name: "r1", DesignWritten: true}}, b)
		snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
		if len(snap.Sessions) != 1 || snap.Sessions[0].Cleanup != tc.want {
			t.Errorf("sessions = %+v, want cleanup %q", snap.Sessions, tc.want)
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
