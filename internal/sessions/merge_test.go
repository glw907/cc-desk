package sessions

import (
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

func pull(head string, checks ...Check) Pull {
	return Pull{Number: 7, State: "OPEN", Head: head, Checks: checks}
}

func ok(name string) Check     { return Check{Name: name, Bucket: Pass} }
func failed(name string) Check { return Check{Name: name, Bucket: Fail} }

// ready reports pull request 7 at commit abc, accepted and inside the design.
func ready(mod func(pr *state.PullRequest)) func(*state.ReleaseFile, *state.Session, *state.Pass) {
	return func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
		pr := &state.PullRequest{Number: 7, Commit: "abc", Verdict: state.VerdictAccept, InsideDesign: true, Ready: ago(10)}
		if mod != nil {
			mod(pr)
		}
		p.PullRequest = pr
	}
}

func TestDecideMerge(t *testing.T) {
	pr := func(m func(*state.PullRequest)) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return ready(m)
	}
	merged := pull("abc")
	merged.State, merged.Merged = "MERGED", ago(2)
	closed := pull("abc", ok("gate"))
	closed.State = "CLOSED"
	tests := []struct {
		name         string
		mod          func(*state.ReleaseFile, *state.Session, *state.Pass)
		read         PullRead
		wantDecision string
		wantReason   string
		wantWait     *string
		wantOutside  bool
		failures     int
	}{
		{"1 merged, marker unset: outside", pr(nil), PullRead{Pulls: []Pull{merged}}, Merged, "", nil, true, 0},
		{"1 merged, marker set: own", pr(func(p *state.PullRequest) { p.MergeCalled = ago(3) }), PullRead{Pulls: []Pull{merged}}, Merged, "", nil, false, 0},
		{"1 merged with no report", nil, PullRead{Pulls: []Pull{merged}}, Merged, "", nil, true, 0},
		{"2 no report stands", nil, PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, None, "", waitPtr(""), false, 0},
		{"2 closed", pr(nil), PullRead{Pulls: []Pull{closed}}, None, "", waitPtr(""), false, 0},
		{"2 no pull request", pr(nil), PullRead{}, None, "", waitPtr(""), false, 0},
		{"3 question asked", pr(func(p *state.PullRequest) { p.Question = "merge-x-1" }), PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, None, "", waitPtr(""), false, 0},
		{"4 outside the design", pr(func(p *state.PullRequest) { p.InsideDesign = false }), PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Ask, ReasonOutsideDesign, nil, false, 0},
		{"4 outside beats a failed check", pr(func(p *state.PullRequest) { p.InsideDesign = false }), PullRead{Pulls: []Pull{pull("abc", failed("gate"))}}, Ask, ReasonOutsideDesign, nil, false, 0},
		{"5 head moved", pr(nil), PullRead{Pulls: []Pull{pull("def", ok("gate"))}}, None, "", waitPtr(state.WaitCommitMoved), false, 0},
		{"5 moved beats a failed check", pr(nil), PullRead{Pulls: []Pull{pull("def", failed("gate"))}}, None, "", waitPtr(state.WaitCommitMoved), false, 0},
		{"6 no check", pr(nil), PullRead{Pulls: []Pull{pull("abc")}}, None, "", waitPtr(state.WaitNoCheck), false, 0},
		{"7 check running", pr(nil), PullRead{Pulls: []Pull{pull("abc", ok("a"), Check{Name: "b", Bucket: Pending})}}, None, "", waitPtr(state.WaitCheckRunning), false, 0},
		{"8 failed check", pr(nil), PullRead{Pulls: []Pull{pull("abc", ok("a"), failed("b"))}}, Fix, ReasonFix, nil, false, 1},
		{"8 cancelled check", pr(nil), PullRead{Pulls: []Pull{pull("abc", Check{Name: "b", Bucket: Cancel})}}, Fix, ReasonFix, nil, false, 1},
		{"8 verdict fix", pr(func(p *state.PullRequest) { p.Verdict = state.VerdictFix }), PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Fix, ReasonFix, nil, false, 1},
		{"8 one attempt used", pr(func(p *state.PullRequest) { p.FixAttempts = 1 }), PullRead{Pulls: []Pull{pull("abc", failed("gate"))}}, Fix, ReasonFix, nil, false, 1},
		{"9 fix limit", pr(func(p *state.PullRequest) { p.FixAttempts = 2 }), PullRead{Pulls: []Pull{pull("abc", failed("gate"))}}, Ask, ReasonFixLimit, nil, false, 1},
		{"10 merge", pr(nil), PullRead{Pulls: []Pull{pull("abc", ok("a"), Check{Name: "b", Bucket: Skipping})}}, Merge, "", nil, false, 0},
		{"10 verdict fix stops it", pr(func(p *state.PullRequest) { p.Verdict = state.VerdictFix }), PullRead{Pulls: []Pull{pull("abc", ok("a"))}}, Fix, ReasonFix, nil, false, 1},
		{"unreachable", pr(nil), PullRead{Unreachable: true}, None, "", waitPtr(state.WaitUnreachable), false, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, key := world("pass", 10, tt.mod)
			ev := Evidence{
				Records: map[Key]RecordEvidence{key: {Row: rowWorking, HasRow: true}},
				Pulls:   map[Key]PullRead{key: tt.read},
			}
			got := Decide(s, ev, t0)
			if len(got) != 1 {
				t.Fatalf("decisions = %d, want 1", len(got))
			}
			r := got[0]
			if r.Decision.Decision != tt.wantDecision || r.Reason != tt.wantReason {
				t.Errorf("decision = %s/%s, want %s/%s", r.Decision.Decision, r.Reason, tt.wantDecision, tt.wantReason)
			}
			switch {
			case (r.Merge.Wait == nil) != (tt.wantWait == nil):
				t.Errorf("wait = %v, want %v", r.Merge.Wait, tt.wantWait)
			case r.Merge.Wait != nil && *r.Merge.Wait != *tt.wantWait:
				t.Errorf("wait = %q, want %q", *r.Merge.Wait, *tt.wantWait)
			}
			if tt.wantDecision == Merged && r.Merge.Outside != tt.wantOutside {
				t.Errorf("outside = %v", r.Merge.Outside)
			}
			if len(r.Merge.Failures) != tt.failures {
				t.Errorf("failures = %v", r.Merge.Failures)
			}
			if tt.wantDecision == Fix && (r.Counted == nil || *r.Counted) {
				t.Errorf("a fix must be uncounted: %v", r.Counted)
			}
		})
	}
}

func TestDecideMergeRowRulesStillRun(t *testing.T) {
	// A wait runs the row rules: a blocked session with a prompt still stops,
	// and an idle one with a report standing is not ended for idleness.
	s, key := world("pass", 10, ready(nil))
	ev := Evidence{
		Records: map[Key]RecordEvidence{key: {Row: rowPrompt, HasRow: true}},
		Pulls:   map[Key]PullRead{key: {Pulls: []Pull{pull("abc", Check{Name: "b", Bucket: Pending})}}},
	}
	if got := Decide(s, ev, t0)[0]; got.Decision.Decision != Stop {
		t.Errorf("blocked session under a wait: %s, want stop", got.Decision.Decision)
	}

	s, key = world("pass", 10, func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
		ready(nil)(rf, sess, p)
		sess.LastWorking = ago(40 * 60 * 1e9)
	})
	ev = Evidence{
		Records: map[Key]RecordEvidence{key: {Row: rowDone, HasRow: true}},
		Pulls:   map[Key]PullRead{key: {Pulls: []Pull{pull("abc", Check{Name: "b", Bucket: Pending})}}},
	}
	got := Decide(s, ev, t0)[0]
	if got.Decision.Decision != None || got.Detail != "a check is running" {
		t.Errorf("idle session under a wait: %s %q, want none with the wait named", got.Decision.Decision, got.Detail)
	}
}

func TestDecideMergeFixWaitsOutTheUsageHold(t *testing.T) {
	for _, usage := range []int{100, -1} {
		s, key := world("pass", usage, ready(nil))
		ev := Evidence{
			Records: map[Key]RecordEvidence{key: {Row: rowDone, HasRow: true}},
			Pulls:   map[Key]PullRead{key: {Pulls: []Pull{pull("abc", failed("gate"))}}},
		}
		got := Decide(s, ev, t0)[0]
		if got.Decision.Decision != WaitUsage || got.Merge.Failures != nil {
			t.Errorf("usage %d: failed check decided %s with failures %v, want wait_usage and none recorded", usage, got.Decision.Decision, got.Merge.Failures)
		}
	}
}

func TestDecideMergeStoppedPass(t *testing.T) {
	stop := func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
		ready(nil)(rf, sess, p)
		p.State, p.StopReason = state.Stopped, state.StopScope
	}
	merged := pull("abc")
	merged.State, merged.Merged = "MERGED", ago(2)
	s, key := world("pass", 10, stop)
	got := Decide(s, Evidence{Pulls: map[Key]PullRead{key: {Pulls: []Pull{merged}}}}, t0)[0]
	if got.Decision.Decision != Merged {
		t.Errorf("stopped pass, merged pull request: %s, want merged", got.Decision.Decision)
	}
	s, key = world("pass", 10, stop)
	got = Decide(s, Evidence{Pulls: map[Key]PullRead{key: {Pulls: []Pull{pull("abc", failed("gate"))}}}}, t0)[0]
	if got.Decision.Decision != None || got.Reason != string(state.StopScope) || got.Merge.Wait != nil {
		t.Errorf("stopped pass, failed check: %+v, want its stop row and no wait", got)
	}
}

func TestCleanupRow(t *testing.T) {
	tests := []struct {
		name  string
		mod   func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass)
		wants string
	}{
		{"merged, session live", doneWith(false, false), Cleanup},
		{"merged, session ended", doneWith(true, false), Cleanup},
		{"merged and cleaned", doneWith(true, true), None},
		{"done with no pull request", func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) { p.State = state.Done }, None},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := world("pass", 10, tt.mod)
			got := Decide(s, Evidence{}, t0)
			if len(got) != 1 || got[0].Decision.Decision != tt.wants {
				t.Fatalf("decisions = %+v, want %s", got, tt.wants)
			}
		})
	}
}

func doneWith(ended, cleaned bool) func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
	return func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass) {
		p.State = state.Done
		p.PullRequest = &state.PullRequest{Number: 7, Merged: ago(5)}
		if ended {
			sess.Ended = ago(4)
		}
		if cleaned {
			p.PullRequest.Cleaned = ago(1)
		}
	}
}
