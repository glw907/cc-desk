package state

import (
	"strings"
	"testing"
	"time"
)

func TestSameStart(t *testing.T) {
	base := time.Date(2026, 11, 18, 8, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		b    time.Time
		want bool
	}{
		{"equal", base, true},
		{"1 second after", base.Add(time.Second), true},
		{"1 second before", base.Add(-time.Second), true},
		{"3 seconds after", base.Add(3 * time.Second), false},
		{"3 seconds before", base.Add(-3 * time.Second), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SameStart(base, tt.b); got != tt.want {
				t.Errorf("SameStart = %v, want %v", got, tt.want)
			}
		})
	}
}

const deskSession = "desk-session"

// deskState is ruleState with one registered desk: PID 500, started at 8:00.
func deskState() (*State, *ReleaseFile) {
	s, rf := ruleState()
	s.Desks = Desks{Format: Format, Desks: []Desk{{SessionID: deskSession, PID: 500, Started: at(8, 0)}}}
	return s, rf
}

// deskCaller is a command with no terminal of its own, run under the desk's
// process, which has one.
func deskCaller() Caller {
	return Caller{SessionID: deskSession, Chain: []Proc{
		{PID: 900, Started: at(8, 5), TTY: 0},
		{PID: 500, Started: at(8, 0), TTY: 34820},
		{PID: 1, Started: at(7, 0)},
	}}
}

var answerQ1 = RulingRequest{Words: "Yes.", Project: "alpha", Round: "r1", Question: "q1", Option: "a"}

func TestRecordRuling(t *testing.T) {
	notDesk := func() Caller { c := deskCaller(); c.SessionID = "other"; return c }
	tests := []struct {
		name    string
		caller  func() Caller
		req     RulingRequest
		setup   func(s *State, rf *ReleaseFile)
		refused string // exact refusal, when the verb refuses
		bad     bool   // a bad request, exit 1
		wantID  string
	}{
		{name: "an answer is recorded", caller: deskCaller, req: answerQ1, wantID: "r-0001"},
		{name: "a ruling with no question is recorded", caller: deskCaller, req: RulingRequest{Words: "Note."}, wantID: "r-0001"},
		{
			name: "an empty session id fails test 1", req: answerQ1, refused: RefuseNotDesk,
			caller: func() Caller { c := deskCaller(); c.SessionID = ""; return c },
		},
		{name: "an unregistered session fails test 1", req: answerQ1, refused: RefuseNotDesk, caller: notDesk},
		{
			name: "a dead desk PID fails test 2", req: answerQ1, refused: RefuseNotUnderDesk,
			caller: func() Caller { c := deskCaller(); c.Chain = c.Chain[:1]; return c },
		},
		{
			name: "a start time 1 second off passes", req: answerQ1, wantID: "r-0001",
			caller: func() Caller { c := deskCaller(); c.Chain[1].Started = at(8, 0).Add(time.Second); return c },
		},
		{
			name: "a start time 3 seconds off fails test 2", req: answerQ1, refused: RefuseNotUnderDesk,
			caller: func() Caller { c := deskCaller(); c.Chain[1].Started = at(8, 0).Add(3 * time.Second); return c },
		},
		{
			name: "a desk process with no terminal fails test 3", req: answerQ1, refused: RefuseNoTerminal,
			caller: func() Caller { c := deskCaller(); c.Chain[1].TTY = 0; return c },
		},
		{
			name: "the caller's own terminal doesn't count", req: answerQ1, refused: RefuseNoTerminal,
			caller: func() Caller { c := deskCaller(); c.Chain[0].TTY = 7; c.Chain[1].TTY = 0; return c },
		},
		{
			name: "a non-desk naming an unknown question gets test 1", refused: RefuseNotDesk, caller: notDesk,
			req: RulingRequest{Words: "x", Project: "alpha", Round: "r1", Question: "nope", Option: "a"},
		},
		{name: "empty words", caller: deskCaller, req: RulingRequest{}, bad: true},
		{name: "words that aren't UTF-8", caller: deskCaller, req: RulingRequest{Words: "\xff\xfe"}, bad: true},
		{name: "an unknown project", caller: deskCaller, bad: true, req: RulingRequest{Words: "x", Project: "zed", Round: "r1", Question: "q1", Option: "a"}},
		{name: "an unknown round", caller: deskCaller, bad: true, req: RulingRequest{Words: "x", Project: "alpha", Round: "zed", Question: "q1", Option: "a"}},
		{name: "an unknown question", caller: deskCaller, bad: true, req: RulingRequest{Words: "x", Project: "alpha", Round: "r1", Question: "zed", Option: "a"}},
		{name: "an unknown option", caller: deskCaller, bad: true, req: RulingRequest{Words: "x", Project: "alpha", Round: "r1", Question: "q1", Option: "zed"}},
		{name: "only some of the four", caller: deskCaller, bad: true, req: RulingRequest{Words: "x", Project: "alpha", Question: "q1"}},
		{
			name: "an answered question is refused", caller: deskCaller, req: answerQ1,
			refused: `cannot record the ruling: question "q1" is already answered`,
			setup:   func(s *State, rf *ReleaseFile) { rf.Questions[0].Answer = &Answer{Ruling: "r-0001"} },
		},
		{
			name: "the next id follows the count", caller: deskCaller, req: answerQ1, wantID: "r-0002",
			setup: func(s *State, rf *ReleaseFile) { s.Rulings = []Ruling{{ID: "r-0001"}} },
		},
		{
			name: "a taken id moves on", caller: deskCaller, req: answerQ1, wantID: "r-0004",
			setup: func(s *State, rf *ReleaseFile) { s.Rulings = []Ruling{{ID: "r-0003"}, {ID: "r-0002"}} },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := deskState()
			if tt.setup != nil {
				tt.setup(s, rf)
			}
			before := len(s.Rulings)
			err := RecordRuling(tt.caller(), tt.req)(s, at(9, 0))
			switch {
			case tt.refused != "":
				if !isRefusal(err) || err.Error() != tt.refused {
					t.Fatalf("error = %v, want the refusal %q", err, tt.refused)
				}
			case tt.bad:
				if err == nil || isRefusal(err) {
					t.Fatalf("error = %v, want a bad request", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
			if tt.wantID == "" {
				if len(s.Rulings) != before || s.newRuling != nil || rf.Dirty {
					t.Error("a failed ruling changed the state")
				}
				return
			}
			got := s.newRuling
			if got == nil || got.ID != tt.wantID || got.Desk != deskSession || got.Words != tt.req.Words || !got.Time.Equal(at(9, 0)) {
				t.Fatalf("ruling = %+v, want id %s", got, tt.wantID)
			}
			if tt.req.Question == "" {
				if rf.Dirty || got.Question != "" {
					t.Error("a ruling with no question touched a release")
				}
				return
			}
			if a := rf.Questions[0].Answer; a == nil || a.Ruling != tt.wantID || !rf.Dirty {
				t.Errorf("answer = %+v, dirty %v", a, rf.Dirty)
			}
		})
	}
}

func TestRecordRulingClosesGuess(t *testing.T) {
	s, rf := deskState()
	p := addPass(rf, "g", "r1", 200, Building)
	t0 := at(8, 0)
	p.Started = &t0
	p.Guess = &Guess{Question: "q1", Recommendation: "a", At: t0}
	if got := s.OpenGuessMinutes(); got != 200 {
		t.Fatalf("open guess minutes = %d, want 200", got)
	}
	if err := RecordRuling(deskCaller(), answerQ1)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	if got := s.OpenGuessMinutes(); got != 0 {
		t.Errorf("open guess minutes = %d after the ruling, want 0", got)
	}
}

func TestRegisterDesk(t *testing.T) {
	t.Run("a new entry", func(t *testing.T) {
		s, _ := deskState()
		d := Desk{SessionID: "s2", PID: 7, Started: at(8, 0)}
		if err := RegisterDesk(d)(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		if len(s.Desks.Desks) != 2 || s.Desks.Desks[1] != d || !s.desksDirty {
			t.Errorf("desks = %+v, dirty %v", s.Desks, s.desksDirty)
		}
	})
	t.Run("a replaced entry", func(t *testing.T) {
		s, _ := deskState()
		d := Desk{SessionID: deskSession, PID: 8, Started: at(8, 30)}
		if err := RegisterDesk(d)(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		if len(s.Desks.Desks) != 1 || s.Desks.Desks[0] != d {
			t.Errorf("desks = %+v", s.Desks.Desks)
		}
	})
	for name, bad := range map[string]Desk{
		"no session id":  {PID: 7, Started: at(8, 0)},
		"a PID of 0":     {SessionID: "s", Started: at(8, 0)},
		"a negative PID": {SessionID: "s", PID: -1, Started: at(8, 0)},
		"no start time":  {SessionID: "s", PID: 7},
	} {
		t.Run(name, func(t *testing.T) {
			s, _ := deskState()
			if err := RegisterDesk(bad)(s, at(9, 0)); err == nil || isRefusal(err) {
				t.Errorf("error = %v, want a bad request", err)
			}
			if s.desksDirty || len(s.Desks.Desks) != 1 {
				t.Error("a bad desk changed the state")
			}
		})
	}
}

const rulingRelease = `{
  "format": 1, "project": "alpha", "name": "v1", "repo": "glw907/alpha",
  "rounds": [{"name": "r1", "summary": "First.", "needs_geoff": true}],
  "passes": [{"name": "p1", "summary": "Build.", "round": "r1", "planned_minutes": 60, "state": "not_started", "questions": ["q1"]}],
  "questions": [{"id": "q1", "text": "t", "options": ["a", "b"], "recommended": "a"}]
}`

func TestCommitRuling(t *testing.T) {
	desks := `{"format":1,"desks":[{"session_id":"` + deskSession + `","pid":500,"started":"` + at(8, 0).Format(time.RFC3339) + `"}]}`
	const line1 = `{"format":1,"id":"r-0001","time":"2026-11-20T22:00:00Z","desk":"d","words":"old"}`

	t.Run("the line lands with a newline put in first", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", rulingRelease)
		writeFile(t, dir, DesksFile, desks)
		writeFile(t, dir, RulingFile, line1) // no trailing newline
		if _, err := Commit(dir, testClock, time.Second, RecordRuling(deskCaller(), answerQ1), render); err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(readFile(t, dir, RulingFile), "\n")
		if len(lines) != 3 || lines[0] != line1 || lines[2] != "" {
			t.Fatalf("rulings.jsonl = %q", lines)
		}
		if !strings.Contains(lines[1], `"id":"r-0002"`) || !strings.Contains(lines[1], `"words":"Yes."`) {
			t.Errorf("new line = %s", lines[1])
		}
		if !strings.Contains(readFile(t, dir, "releases/alpha.json"), `"ruling": "r-0002"`) {
			t.Error("the question's answer was not written")
		}
		if _, err := Load(dir); err != nil {
			t.Errorf("reload: %v", err)
		}
	})

	t.Run("a refused ruling leaves the folder as it was", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", rulingRelease)
		writeFile(t, dir, DesksFile, desks)
		writeFile(t, dir, RulingFile, line1)
		c := deskCaller()
		c.Chain[1].TTY = 0
		_, err := Commit(dir, testClock, time.Second, RecordRuling(c, answerQ1), render)
		if err == nil || err.Error() != RefuseNoTerminal {
			t.Fatalf("error = %v, want %q", err, RefuseNoTerminal)
		}
		if readFile(t, dir, RulingFile) != line1 || readFile(t, dir, "releases/alpha.json") != rulingRelease || exists(dir, SnapshotFile) {
			t.Error("a refusal wrote a file")
		}
	})

	t.Run("a registered desk writes desks.json", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "releases/alpha.json", rulingRelease)
		d := Desk{SessionID: "s9", PID: 12, Started: at(8, 0)}
		if _, err := Commit(dir, testClock, time.Second, RegisterDesk(d), render); err != nil {
			t.Fatal(err)
		}
		got := readFile(t, dir, DesksFile)
		if !strings.Contains(got, `"session_id": "s9"`) || !strings.Contains(got, `"format": 1`) {
			t.Errorf("desks.json = %s", got)
		}
		s, err := Load(dir)
		if err != nil || len(s.Desks.Desks) != 1 {
			t.Errorf("reload: %v, %+v", err, s)
		}
	})
}
