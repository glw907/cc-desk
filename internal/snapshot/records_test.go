package snapshot

import (
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

func TestNeedsYouNowEndedRecords(t *testing.T) {
	ended := clock.Add(-time.Minute)
	tests := []struct {
		name   string
		stop   state.StopReason
		ended  bool
		onPass bool
		want   int
	}{
		{"a conductor stop that has not ended", state.StopPermissionPrompt, false, false, 1},
		{"an ended conductor stop", state.StopPermissionPrompt, true, false, 0},
		{"an ended conductor whose launch failed", state.StopLaunchFailed, true, false, 1},
		{"an ended round stop", state.StopSignInExpired, true, false, 0},
		{"a stopped pass whose session ended", state.StopPermissionPrompt, true, true, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := asking("a", 0, true)
			stopped := clock.Add(-10 * time.Minute)
			if tt.onPass {
				b := building("w", "r1", 10, clock.Add(-20*time.Minute))
				b.State, b.StopReason = state.Stopped, tt.stop
				b.Stops = []state.StopInterval{{Start: stopped}}
				b.Session = sessionOn(state.KindPass, "a-w")
				b.Session.Ended = &ended
				rf.Passes = append(rf.Passes, b)
			} else {
				c := sessionOn(state.KindConductor, "a-conductor")
				c.StopReason, c.Stopped = tt.stop, &stopped
				if tt.ended {
					c.Ended = &ended
				}
				rf.Conductor = c
				if tt.stop == state.StopSignInExpired {
					rf.Conductor = nil
					rf.Rounds[0].Session = c
				}
			}
			snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
			if len(snap.NeedsYouNow) != tt.want {
				t.Errorf("needs_you_now = %+v, want %d entries", snap.NeedsYouNow, tt.want)
			}
			if len(snap.Sessions) != 1 {
				t.Errorf("sessions = %d, want the record kept", len(snap.Sessions))
			}
		})
	}
}

func TestStopTime(t *testing.T) {
	rf := asking("a", 0, true)
	b := building("w", "r1", 10, clock.Add(-30*time.Minute))
	start := clock.Add(-10 * time.Minute)
	b.State, b.StopReason = state.Stopped, state.StopScope
	b.Stops = []state.StopInterval{{Start: start}}
	b.Session = sessionOn(state.KindPass, "a-w")
	rf.Passes = append(rf.Passes, b)
	st := &state.State{Releases: []*state.ReleaseFile{rf}}

	snap := Build(st, clock)
	p := snap.Releases[0].Rounds[0].Passes[len(snap.Releases[0].Rounds[0].Passes)-1]
	if p.StoppedAt == nil || !p.StoppedAt.Equal(start) {
		t.Errorf("stopped pass stopped_at = %v, want %v", p.StoppedAt, start)
	}
	if got := snap.Sessions[0].StoppedAt; got == nil || !got.Equal(start) {
		t.Errorf("session stopped_at = %v, want %v", got, start)
	}

	end := clock.Add(-5 * time.Minute)
	resumed := &rf.Passes[len(rf.Passes)-1]
	resumed.State, resumed.StopReason = state.Building, ""
	resumed.Stops = []state.StopInterval{{Start: start, End: &end}}
	snap = Build(st, clock)
	p = snap.Releases[0].Rounds[0].Passes[len(snap.Releases[0].Rounds[0].Passes)-1]
	if p.StoppedAt != nil || snap.Sessions[0].StoppedAt != nil {
		t.Errorf("after it resumed, stopped_at = %v and %v, want null", p.StoppedAt, snap.Sessions[0].StoppedAt)
	}
}

func TestPassReason(t *testing.T) {
	tests := []struct {
		name, own, round, want string
		questions              []string
		kind                   string
	}{
		{"its own", "one line of work", "the round's", "one line of work", nil, "nothing"},
		{"the round's", "", "the round's", "the round's", nil, "nothing"},
		{"neither", "", "", "", nil, "nothing"},
		{"a question", "", "", "", []string{"q1"}, "questions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rf := asking("a", 0, true)
			rf.Rounds[0].NoNeedReason = tt.round
			b := building("w", "r1", 10, clock.Add(-30*time.Minute))
			b.State, b.Started = state.NotStarted, nil
			b.NoNeedReason, b.Questions = tt.own, tt.questions
			if len(tt.questions) > 0 {
				rf.Questions = append(rf.Questions, state.Question{ID: "q1", Text: "T?", Options: []string{"a", "b"}, Recommended: "a"})
			}
			rf.Passes = append(rf.Passes, b)
			snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
			ps := snap.Releases[0].Rounds[0].Passes
			p := ps[len(ps)-1]
			if p.NeedsKind != tt.kind {
				t.Errorf("needs_kind = %s, want %s", p.NeedsKind, tt.kind)
			}
			switch {
			case tt.want == "" && p.NeedsReason != nil:
				t.Errorf("needs_reason = %q, want null", *p.NeedsReason)
			case tt.want != "" && (p.NeedsReason == nil || *p.NeedsReason != tt.want):
				t.Errorf("needs_reason = %v, want %q", p.NeedsReason, tt.want)
			}
		})
	}
}

func TestSessionRefusals(t *testing.T) {
	rf := asking("a", 0, true)
	c := sessionOn(state.KindConductor, "a-conductor")
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	if len(snap.Sessions) != 0 {
		t.Fatalf("sessions = %d before a record exists", len(snap.Sessions))
	}
	rf.Conductor = c
	snap = Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	if r := snap.Sessions[0].Refusals; r == nil || len(r) != 0 {
		t.Errorf("refusals = %v, want an empty list", r)
	}
	c.Refusals = []state.Refusal{{Time: clock, Tool: "Bash", Target: "rm -r build", Reason: "an ask rule"}}
	snap = Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	if r := snap.Sessions[0].Refusals; len(r) != 1 || r[0].Tool != "Bash" || r[0].Target != "rm -r build" {
		t.Errorf("refusals = %+v", r)
	}
}
