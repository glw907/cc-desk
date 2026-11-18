package sessions

import (
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

var t0 = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

func ago(d time.Duration) *time.Time {
	t := t0.Add(-d)
	return &t
}

// world builds one live release holding one session record of the given kind
// (pass, conductor, or design), seen 5 minutes ago and launched an hour ago.
// usage is the weekly percent; a negative value is an unreadable reading.
func world(kind string, usage int, mod func(rf *state.ReleaseFile, sess *state.Session, p *state.Pass)) (*state.State, Key) {
	sess := &state.Session{
		Name: "p-x", Kind: state.SessionKind(kind), ShortID: "ab12", SessionID: "s-1",
		Worktree: "/w/x", Branch: "x", Launched: *ago(time.Hour), LaunchTries: 1,
		LastSeen: ago(5 * time.Minute), LastState: state.SeenWorking,
	}
	p := &state.Pass{Name: "x", Round: "a", State: state.Building, Started: ago(time.Hour)}
	rf := &state.ReleaseFile{Release: state.Release{Project: "p", Rounds: []state.Round{{Name: "a"}}}}
	key := Key{"p", "pass:x"}
	switch kind {
	case "pass":
		p.Session = sess
	case "conductor":
		rf.Conductor = sess
		key.Record = "conductor"
	case "design":
		rf.Rounds[0].Session = sess
		key.Record = "round:a"
	}
	if mod != nil {
		mod(rf, sess, p)
	}
	rf.Passes = []state.Pass{*p}
	if kind == "pass" {
		sess = rf.Passes[0].Session
	}
	u := &state.Usage{Format: 1, Percent: usage, ReadAt: t0}
	if usage < 0 {
		u = &state.Usage{Format: 1, Error: "unreadable", ReadAt: t0}
	}
	return &state.State{Releases: []*state.ReleaseFile{rf}, Usage: u}, key
}

var (
	rowWorking = Row{ID: "ab12", State: "working", Status: "busy"}
	rowPrompt  = Row{ID: "ab12", State: "blocked", Status: "waiting", WaitingFor: "permission prompt"}
	rowBlocked = Row{ID: "ab12", State: "blocked", Status: "idle"}
	rowDone    = Row{ID: "ab12", State: "done", Status: "idle"}
)

const (
	limitDetail = "You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)"
	lostScreen  = "API Error: Couldn't connect through your proxy (ERR_PROXY_TUNNEL)"
)

func found(r Row) RecordEvidence { return RecordEvidence{Row: r, HasRow: true} }

func TestDecide(t *testing.T) {
	counted2 := func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
		s.Relaunches = []state.Relaunch{{Time: t0, Cause: "x", Counted: true}, {Time: t0, Cause: "x", Counted: true}}
	}
	stoppedWith := func(reason state.StopReason) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return func(_ *state.ReleaseFile, s *state.Session, p *state.Pass) {
			p.State, p.StopReason = state.Stopped, reason
			s.StopReason = ""
		}
	}
	idle := func(by time.Duration) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.LastWorking = ago(by) }
	}
	never := func(tries int) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.LastSeen, s.LaunchTries = nil, tries }
	}
	ended := func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.Ended = ago(time.Minute) }
	finished := func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.Finished = ago(time.Minute) }
	stoppedRecord := func(reason state.StopReason) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
			s.StopReason, s.Stopped = reason, ago(time.Minute)
		}
	}
	chain := func(fs ...func(*state.ReleaseFile, *state.Session, *state.Pass)) func(*state.ReleaseFile, *state.Session, *state.Pass) {
		return func(rf *state.ReleaseFile, s *state.Session, p *state.Pass) {
			for _, f := range fs {
				f(rf, s, p)
			}
		}
	}

	tests := []struct {
		name    string
		kind    string
		usage   int
		mod     func(*state.ReleaseFile, *state.Session, *state.Pass)
		ev      RecordEvidence
		probe   ProbeResult
		want    string
		reason  string
		detail  string
		counted string // "", "true", or "false"
		seen    Observed
	}{
		{name: "row 1 ended", kind: "pass", usage: 40, mod: ended, want: None},
		{name: "row 2 pass done", kind: "pass", usage: 40, mod: func(_ *state.ReleaseFile, _ *state.Session, p *state.Pass) { p.State = state.Done }, want: None},
		{name: "row 3 conductor finished", kind: "conductor", usage: 40, mod: finished, want: End},
		{name: "row 3 design writer finished", kind: "design", usage: 40, mod: finished, want: End},
		{name: "row 4 pass stopped on usage", kind: "pass", usage: 40, mod: stoppedWith(state.StopUsage), want: Relaunch, reason: "usage", counted: "false"},
		{name: "row 4 conductor stopped on usage", kind: "conductor", usage: 40, mod: stoppedRecord(state.StopUsage), want: Relaunch, reason: "usage", counted: "false"},
		{name: "row 4 at the stop line", kind: "pass", usage: 100, mod: stoppedWith(state.StopUsage), want: WaitUsage, reason: "usage"},
		{name: "row 4 unreadable reading", kind: "pass", usage: -1, mod: stoppedWith(state.StopUsage), want: WaitUsage, reason: "usage"},
		{name: "row 5 pass stopped for another reason", kind: "pass", usage: 40, mod: stoppedWith(state.StopScope), want: None, reason: "scope"},
		{name: "row 5 design writer stopped", kind: "design", usage: 40, mod: stoppedRecord(state.StopPermissionPrompt), want: None, reason: "permission_prompt"},
		{name: "row 6 working", kind: "pass", usage: 40, ev: found(rowWorking), want: None, seen: Observed{State: "working", Seen: true, Working: true}},
		{name: "row 7 permission prompt", kind: "pass", usage: 40, ev: found(rowPrompt), want: Stop, reason: "permission_prompt", seen: Observed{State: "blocked", Seen: true}},
		{name: "row 7 permission prompt with an unanswered prompt", kind: "pass", usage: 40, mod: func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
			s.Prompts = []state.Prompt{{Time: *ago(2 * time.Minute), Tool: "Read", Answer: "yes"}, {Time: *ago(time.Minute), Tool: "Bash", Command: "rm -rf node_modules"}}
		}, ev: found(rowPrompt), want: Stop, reason: "permission_prompt", detail: "Bash: rm -rf node_modules", seen: Observed{State: "blocked", Seen: true}},
		{name: "row 8 sign-in expired", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Job: Job{Detail: "Login expired"}}, want: Stop, reason: "sign_in_expired"},
		{name: "row 9 usage limit", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Job: Job{Detail: limitDetail}}, want: Stop, reason: "usage", detail: limitDetail},
		{name: "row 10 lost connection at the limit", kind: "pass", usage: 40, mod: counted2, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, probe: ProbeUp, want: Stop, reason: "relaunch_limit"},
		{name: "row 10 under the usage hold", kind: "pass", usage: 100, mod: counted2, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, want: Stop, reason: "relaunch_limit"},
		{name: "row 11 lost connection, probe up", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, probe: ProbeUp, want: EndAndRelaunch, reason: "lost_connection", counted: "true"},
		{name: "row 11 at the stop line", kind: "pass", usage: 100, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, want: WaitUsage},
		{name: "row 11 unreadable reading", kind: "pass", usage: -1, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, want: WaitUsage},
		{name: "row 12 lost connection, probe down", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: lostScreen}, probe: ProbeDown, want: WaitLink},
		{name: "row 13 unreadable block", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: "something odd\nstill odd  here "}, want: Stop, reason: "unreadable_stop", detail: "still odd here"},
		{name: "row 14 idle at the relaunch limit", kind: "pass", usage: 40, mod: chain(idle(16*time.Minute), counted2), ev: found(rowDone), want: Stop, reason: "relaunch_limit", seen: Observed{State: "idle", Seen: true}},
		{name: "row 15 idle 16 minutes", kind: "pass", usage: 40, mod: idle(16 * time.Minute), ev: found(rowDone), want: EndAndRelaunch, reason: "idle", counted: "true"},
		{name: "row 15 at the stop line", kind: "pass", usage: 100, mod: idle(16 * time.Minute), ev: found(rowDone), want: WaitUsage},
		{name: "row 15 unreadable reading", kind: "pass", usage: -1, mod: idle(16 * time.Minute), ev: found(rowDone), want: WaitUsage},
		{name: "row 16 idle 14 minutes", kind: "pass", usage: 40, mod: idle(14 * time.Minute), ev: found(rowDone), want: None},
		{name: "row 16 ready report stands", kind: "pass", usage: 40, mod: chain(idle(16*time.Minute), func(_ *state.ReleaseFile, _ *state.Session, p *state.Pass) {
			p.PullRequest = &state.PullRequest{Number: 1, Ready: ago(10 * time.Minute)}
		}), ev: found(rowDone), want: None},
		{name: "row 17 no row, never seen", kind: "pass", usage: 40, mod: never(1), want: LaunchAgain, seen: Observed{State: "missing"}},
		{name: "row 17 at the stop line", kind: "pass", usage: 100, mod: never(1), want: WaitUsage},
		{name: "row 17 unreadable reading", kind: "pass", usage: -1, mod: never(1), want: WaitUsage},
		{name: "row 18 second launch failed", kind: "pass", usage: 40, mod: never(2), want: Stop, reason: "launch_failed"},
		{name: "row 18 at the stop line", kind: "pass", usage: 100, mod: never(2), want: Stop, reason: "launch_failed"},
		{name: "row 19 gone at the relaunch limit", kind: "pass", usage: 40, mod: counted2, want: Stop, reason: "relaunch_limit"},
		{name: "row 20 gone under two", kind: "pass", usage: 40, want: Relaunch, reason: "missing", counted: "true", seen: Observed{State: "missing"}},
		{name: "row 20 at the stop line", kind: "pass", usage: 100, want: WaitUsage},
		{name: "row 20 unreadable reading", kind: "pass", usage: -1, want: WaitUsage},
		{name: "row 20 for a conductor", kind: "conductor", usage: 40, want: Relaunch, reason: "missing", counted: "true"},

		{name: "a stopped row is no row", kind: "pass", usage: 40, ev: found(Row{ID: "ab12", State: "stopped"}), want: Relaunch, reason: "missing", counted: "true"},
		{name: "uncounted relaunches do not reach the limit", kind: "pass", usage: 40, mod: func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
			s.Relaunches = []state.Relaunch{{Time: t0, Cause: "x"}, {Time: t0, Cause: "x"}, {Time: t0, Cause: "x"}}
		}, want: Relaunch, reason: "missing", counted: "true"},
		{name: "idle runs from a launch later than the last working", kind: "pass", usage: 40, mod: func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
			s.LastWorking, s.Launched = ago(30*time.Minute), *ago(5 * time.Minute)
		}, ev: found(rowDone), want: None},
		{name: "idle runs from the last working after the launch", kind: "pass", usage: 40, mod: func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) {
			s.LastWorking, s.Launched = ago(16*time.Minute), *ago(time.Hour)
		}, ev: found(rowDone), want: EndAndRelaunch, reason: "idle", counted: "true"},
		{name: "idle with no last working runs from the launch", kind: "pass", usage: 40, ev: found(rowDone), want: EndAndRelaunch, reason: "idle", counted: "true"},
		{name: "a limit with no reset text keeps the sentence", kind: "pass", usage: 40, ev: RecordEvidence{Row: rowBlocked, HasRow: true, Screen: "x You've hit your Fable limit"}, want: Stop, reason: "usage", detail: "You've hit your Fable limit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, key := world(tt.kind, tt.usage, tt.mod)
			ev := Evidence{Records: map[Key]RecordEvidence{}, Probe: tt.probe}
			if tt.ev.HasRow || tt.ev.Screen != "" || tt.ev.Job != (Job{}) {
				ev.Records[key] = tt.ev
			} else {
				ev.Records[key] = RecordEvidence{}
			}
			got := Decide(s, ev, t0)
			if len(got) != 1 {
				t.Fatalf("got %d results, want 1", len(got))
			}
			r := got[0]
			if r.Decision.Decision != tt.want || r.Reason != tt.reason || r.Detail != tt.detail {
				t.Fatalf("got %s/%q/%q, want %s/%q/%q", r.Decision.Decision, r.Reason, r.Detail, tt.want, tt.reason, tt.detail)
			}
			c := ""
			if r.Counted != nil {
				c = map[bool]string{true: "true", false: "false"}[*r.Counted]
			}
			if c != tt.counted {
				t.Errorf("counted = %q, want %q", c, tt.counted)
			}
			if r.Project != "p" || r.Record != key.Record || r.Session != "p-x" {
				t.Errorf("names = %q %q %q", r.Project, r.Record, r.Session)
			}
			if tt.seen != (Observed{}) && r.Observed != tt.seen {
				t.Errorf("observed = %+v, want %+v", r.Observed, tt.seen)
			}
		})
	}
}

func TestDecideRowsLeaveOtherNamesAlone(t *testing.T) {
	// A conductor stopped for a reason other than usage reads none, and
	// one finished and stopped is ended: rows 3 and 5 come in that order.
	s, key := world("conductor", 40, func(_ *state.ReleaseFile, sess *state.Session, _ *state.Pass) {
		sess.Finished, sess.StopReason, sess.Stopped = ago(time.Minute), state.StopScope, ago(time.Minute)
	})
	got := Decide(s, Evidence{Records: map[Key]RecordEvidence{key: {}}}, t0)
	if got[0].Decision.Decision != End {
		t.Fatalf("decision = %s, want end", got[0].Decision.Decision)
	}
}

func TestDecideNoEvidenceLeavesTheRecordAlone(t *testing.T) {
	s, _ := world("pass", 40, nil)
	got := Decide(s, Evidence{}, t0)
	if len(got) != 1 || got[0].Decision.Decision != None || got[0].Detail == "" || got[0].Observed != (Observed{}) {
		t.Fatalf("got %+v", got)
	}
}

func TestDecideSecondReadFindsTheRow(t *testing.T) {
	// The first read had no row, which alone would launch again. The second
	// read found one, and the evidence holds only the second read's rows.
	s, key := world("pass", 40, func(_ *state.ReleaseFile, sess *state.Session, _ *state.Pass) { sess.LastSeen = nil })
	first := Decide(s, Evidence{Records: map[Key]RecordEvidence{key: {}}}, t0)
	second := Decide(s, Evidence{Records: map[Key]RecordEvidence{key: found(rowWorking)}}, t0)
	if first[0].Decision.Decision != LaunchAgain || second[0].Decision.Decision != None {
		t.Fatalf("first %s, second %s", first[0].Decision.Decision, second[0].Decision.Decision)
	}
}

func TestWatched(t *testing.T) {
	tests := []struct {
		name string
		kind string
		mod  func(*state.ReleaseFile, *state.Session, *state.Pass)
		want int
	}{
		{"a live pass", "pass", nil, 1},
		{"an ended record", "pass", func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.Ended = ago(time.Minute) }, 0},
		{"a done pass", "pass", func(_ *state.ReleaseFile, _ *state.Session, p *state.Pass) { p.State = state.Done }, 0},
		{"a stopped pass", "pass", func(_ *state.ReleaseFile, _ *state.Session, p *state.Pass) {
			p.State, p.StopReason = state.Stopped, state.StopScope
		}, 0},
		{"a finished conductor", "conductor", func(_ *state.ReleaseFile, s *state.Session, _ *state.Pass) { s.Finished = ago(time.Minute) }, 0},
		{"a live conductor", "conductor", nil, 1},
		{"a live design writer", "design", nil, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := world(tt.kind, 40, tt.mod)
			if got := Watched(s); len(got) != tt.want {
				t.Fatalf("Watched = %d records, want %d", len(got), tt.want)
			}
		})
	}
}
