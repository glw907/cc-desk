package sessions

import (
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// signInWorld is a pass stopped on the sign-in, with resumes sign-in
// relaunches already on its live session record and a checkout.
func signInWorld(usage, resumes int, mod func(rf *state.ReleaseFile, sess *state.Session)) (*state.State, Key) {
	return world("pass", usage, func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
		rf.Checkout = "/c"
		p.State, p.StopReason = state.Stopped, state.StopSignInExpired
		p.Stops = []state.StopInterval{{Start: *ago(20 * time.Minute)}}
		sess.LastState = state.SeenBlocked
		for range resumes {
			sess.Relaunches = append(sess.Relaunches, state.Relaunch{Time: *ago(30 * time.Minute), Cause: string(state.StopSignInExpired)})
		}
		if mod != nil {
			mod(rf, sess)
		}
	})
}

func TestDecideSignIn(t *testing.T) {
	ended := func(_ *state.ReleaseFile, sess *state.Session) { sess.Ended = ago(10 * time.Minute) }
	noCheckout := func(rf *state.ReleaseFile, _ *state.Session) { rf.Checkout = "" }
	otherCauses := func(_ *state.ReleaseFile, sess *state.Session) {
		sess.Relaunches = append(sess.Relaunches,
			state.Relaunch{Cause: ReasonIdle, Counted: true},
			state.Relaunch{Cause: string(state.StopUsage)})
	}
	signIn := string(state.StopSignInExpired)
	// The stop began 20 minutes before t0.
	fresh, stale := *ago(5 * time.Minute), *ago(30 * time.Minute)
	tests := []struct {
		name     string
		usage    int
		resumes  int
		mod      func(*state.ReleaseFile, *state.Session)
		signedIn time.Time
		decision string
		reason   string
		counted  *bool
		needRead bool
	}{
		{"no good read stays stopped", 10, 0, nil, time.Time{}, None, signIn, nil, true},
		{"a sign-in from before the stop stays stopped", 10, 0, nil, stale, None, signIn, nil, true},
		{"a new sign-in resumes it uncounted", 10, 0, nil, fresh, Relaunch, signIn, new(bool), true},
		{"one resume spent still resumes", 10, 1, nil, fresh, Relaunch, signIn, new(bool), true},
		{"other relaunches do not spend a resume", 10, 1, otherCauses, fresh, Relaunch, signIn, new(bool), true},
		{"two resumes spent stays stopped", 10, 2, nil, fresh, None, signIn, nil, false},
		{"the usage hold waits", 100, 0, nil, fresh, WaitUsage, signIn, nil, true},
		{"the usage hold with no good read stays stopped", 100, 0, nil, time.Time{}, None, signIn, nil, true},
		{"no checkout stays stopped", 10, 0, noCheckout, fresh, None, signIn, nil, false},
		{"an ended record decides none as before", 10, 0, ended, fresh, None, "", nil, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, key := signInWorld(tc.usage, tc.resumes, tc.mod)
			if got := NeedsSignInRead(s); got != tc.needRead {
				t.Errorf("NeedsSignInRead = %v, want %v", got, tc.needRead)
			}
			got := Decide(s, Evidence{Records: map[Key]RecordEvidence{}, SignedIn: tc.signedIn}, t0)
			if len(got) != 1 {
				t.Fatalf("%d results, want 1", len(got))
			}
			d := got[0].Decision
			if d.Record != key.Record || d.Decision != tc.decision || d.Reason != tc.reason {
				t.Errorf("decision = %+v, want %s with reason %q", d, tc.decision, tc.reason)
			}
			if (d.Counted == nil) != (tc.counted == nil) || (d.Counted != nil && *d.Counted != *tc.counted) {
				t.Errorf("counted = %v, want %v", d.Counted, tc.counted)
			}
		})
	}
}

func TestDecideSignInLeavesOtherStopsAlone(t *testing.T) {
	s, _ := signInWorld(10, 0, nil)
	s.Releases[0].Passes[0].StopReason = state.StopUnreadable
	if NeedsSignInRead(s) {
		t.Error("an unreadable stop asked for the sign-in read")
	}
	got := Decide(s, Evidence{SignedIn: *ago(5 * time.Minute)}, t0)
	if d := got[0].Decision; d.Decision != None || d.Reason != string(state.StopUnreadable) {
		t.Errorf("decision = %+v, want none for the unreadable stop", d)
	}

	// A conductor stopped on the sign-in has no pass to resume.
	s, _ = world("conductor", 10, func(rf *state.ReleaseFile, sess *state.Session, _ *state.Pass) {
		sess.StopReason, sess.Stopped = state.StopSignInExpired, ago(time.Minute)
	})
	if NeedsSignInRead(s) {
		t.Error("a conductor's sign-in stop asked for the sign-in read")
	}
	for _, r := range Decide(s, Evidence{SignedIn: *ago(5 * time.Minute)}, t0) {
		if r.Record == "conductor" && r.Decision.Decision == Relaunch {
			t.Errorf("conductor decision = %+v, want no relaunch", r.Decision)
		}
	}
}

func TestPlanResumesASignInStop(t *testing.T) {
	s, _ := signInWorld(10, 0, nil)
	results, err := Plan(s, Evidence{SignedIn: *ago(5 * time.Minute)}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if d := results[0].Decision; d.Decision != Relaunch {
		t.Fatalf("decision = %+v, want relaunch", d)
	}
	p := s.Releases[0].Passes[0]
	if p.State != state.Building || p.StopReason != "" {
		t.Errorf("pass reads %s with stop %q, want building with no stop", p.State, p.StopReason)
	}
	if end := p.Stops[0].End; end == nil || !end.Equal(t0) {
		t.Errorf("stop end = %v, want %s", end, t0)
	}
	if n := p.Session.Counted(); n != 0 {
		t.Errorf("counted relaunches = %d, want 0", n)
	}

}
