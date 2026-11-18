package state

import (
	"strings"
	"testing"
	"time"
)

// bareSession is the identity a launch hands RecordLaunch.
func bareSession(name, worktree string) Session {
	return Session{
		Name: name, Kind: KindPass, Model: "opus", Effort: "high",
		ShortID: "ab12", SessionID: "s-1", Worktree: worktree, Branch: "one",
	}
}

func loadDir(t *testing.T, json string) (string, *State) {
	t.Helper()
	dir := writeRelease(t, "p", json)
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, s
}

func TestReleaseFields(t *testing.T) {
	tests := []struct {
		name, json, want string
	}{
		{"allow rules and overrides", strings.Replace(sessionRelease("not_started", "", `,"allow_rules":["Bash(go test *)"]`),
			`"planned_minutes":10`, `"planned_minutes":10,"model":"claude-opus-5-5","effort":"xhigh"`, 1), ""},
		{"model alias", strings.Replace(sessionRelease("not_started", "", ""), `"planned_minutes":10`, `"planned_minutes":10,"model":"sonnet"`, 1), ""},
		{"unknown model", strings.Replace(sessionRelease("not_started", "", ""), `"planned_minutes":10`, `"planned_minutes":10,"model":"gpt"`, 1), `model "gpt"`},
		{"unknown effort", strings.Replace(sessionRelease("not_started", "", ""), `"planned_minutes":10`, `"planned_minutes":10,"effort":"huge"`, 1), `effort "huge"`},
		{"unknown key beside them", sessionRelease("not_started", "", `,"allow_rule":["x"]`), "unknown field"},
		{"stop detail while building", strings.Replace(sessionRelease("building", "", ""), `"state":"building"`, `"state":"building","stop_detail":"x"`, 1), "stop_detail"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRelease([]byte(tt.json))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}

	r, err := DecodeRelease([]byte(sessionRelease("not_started", "", `,"allow_rules":["Bash(go test *)"]`)))
	if err != nil {
		t.Fatal(err)
	}
	s := &State{}
	if err := AddRelease(r)(s, testClock); err != nil {
		t.Fatalf("AddRelease refused allow_rules: %v", err)
	}
}

func TestRecordLaunch(t *testing.T) {
	building := sessionRelease("building", "", "")
	withSession := sessionRelease("building", goodSession, "")
	run := func(t *testing.T, json string, ev Event) (*State, error) {
		t.Helper()
		dir, s := loadDir(t, json)
		_, err := Commit(dir, testClock, time.Second, ev, render)
		if err != nil {
			return nil, err
		}
		s, lerr := Load(dir)
		if lerr != nil {
			t.Fatal(lerr)
		}
		return s, nil
	}

	t.Run("a first launch", func(t *testing.T) {
		s, err := run(t, building, RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: bareSession("p-x", "/w/one")}))
		if err != nil {
			t.Fatal(err)
		}
		got := s.Releases[0].Passes[0].Session
		if got == nil || got.LaunchTries != 1 || !got.Launched.Equal(testClock) || got.LastSeen != nil {
			t.Errorf("record = %+v", got)
		}
	})

	t.Run("a launch again", func(t *testing.T) {
		s, err := run(t, withSession, RecordLaunch("p", "pass:x", Launch{Kind: LaunchAgain, Replaces: "ab12", Session: bareSession("p-one", "/w/one")}))
		if err != nil {
			t.Fatal(err)
		}
		got := s.Releases[0].Passes[0].Session
		if got.LaunchTries != 2 || got.LastSeen != nil || len(got.Relaunches) != 0 {
			t.Errorf("record = %+v", got)
		}
	})

	t.Run("a relaunch resets the tries and what the check saw", func(t *testing.T) {
		seen := strings.Replace(goodSession, `"launch_tries":1`, `"launch_tries":2,"last_seen":"2026-11-20T10:05:00Z","last_state":"blocked"`, 1)
		s, err := run(t, sessionRelease("building", seen, ""), RecordLaunch("p", "pass:x", Launch{
			Kind: LaunchRelaunch, Replaces: "ab12", Session: bareSession("p-one", "/w/one"), Cause: "idle", Counted: true,
		}))
		if err != nil {
			t.Fatal(err)
		}
		got := s.Releases[0].Passes[0].Session
		if got.LaunchTries != 1 || got.LastSeen != nil || got.LastState != "" {
			t.Errorf("record = %+v", got)
		}
		if len(got.Relaunches) != 1 || got.Relaunches[0].Cause != "idle" || !got.Relaunches[0].Counted || got.Counted() != 1 {
			t.Errorf("relaunches = %+v", got.Relaunches)
		}
	})

	refused := []struct {
		name, json string
		ev         Event
		want       string
	}{
		{"pass no longer building", sessionRelease("not_started", "", ""),
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: bareSession("p-x", "/w/one")}), "not building"},
		{"a live record", withSession,
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: bareSession("p-x", "/w/two")}), "already has a live session record"},
		{"a launch again with no record", building,
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchAgain, Replaces: "ab12", Session: bareSession("p-x", "/w/one")}), "no longer the live record"},
		{"a relaunch of a record another launch replaced", withSession,
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchRelaunch, Replaces: "zz99", Session: bareSession("p-one", "/w/one")}), "no longer the live record"},
		{"a design record on a pass", building,
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: func() Session { s := bareSession("p-x", "/w/one"); s.Kind = KindDesign; return s }()}), `session kind "design"`},
		{"a record that fails validation", building,
			RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: bareSession("p-x", "relative")}), "not an absolute path"},
		{"an unknown record", building,
			RecordLaunch("p", "round:zzz", Launch{Kind: LaunchFirst, Session: bareSession("p-x", "/w/one")}), "no record"},
	}
	for _, tt := range refused {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := run(t, tt.json, tt.ev); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}

	t.Run("another live record holds the worktree or the name", func(t *testing.T) {
		two := strings.Replace(building, `"passes":[`, `"passes":[{"name":"y","summary":"s","round":"a","planned_minutes":10,"state":"building","started":"2026-11-20T10:00:00Z","session":`+goodSession+`},`, 1)
		for _, tt := range []struct{ name, worktree, want string }{
			{"p-new", "/w/one", "one worktree"},
			{"p-one", "/w/new", "one session"},
		} {
			_, err := run(t, two, RecordLaunch("p", "pass:x", Launch{Kind: LaunchFirst, Session: bareSession(tt.name, tt.worktree)}))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("%s: error = %v, want it to hold %q", tt.name, err, tt.want)
			}
		}
	})

	t.Run("an ended conductor is replaced", func(t *testing.T) {
		ended := strings.Replace(`{"name":"p-conductor","kind":"conductor","short_id":"c1","session_id":"s-c","worktree":"/w/c","branch":"conductor","launched":"2026-11-20T09:00:00Z","launch_tries":1,"ended":"2026-11-20T09:30:00Z"}`, "\n", "", 1)
		json := sessionRelease("not_started", "", `,"conductor":`+ended)
		fresh := bareSession("p-conductor", "/w/c")
		fresh.Kind, fresh.Branch = KindConductor, "conductor"
		s, err := run(t, json, RecordLaunch("p", "conductor", Launch{Kind: LaunchFirst, Session: fresh}))
		if err != nil {
			t.Fatal(err)
		}
		if got := s.Releases[0].Conductor; got.Ended != nil || got.ShortID != "ab12" {
			t.Errorf("conductor = %+v", got)
		}
	})
}

func TestEndAndFinishSession(t *testing.T) {
	conductor := `{"name":"p-conductor","kind":"conductor","short_id":"c1","session_id":"s-c","worktree":"/w/c","branch":"conductor","launched":"2026-11-20T09:00:00Z","launch_tries":1}`
	json := sessionRelease("building", goodSession, `,"conductor":`+conductor)
	dir, _ := loadDir(t, json)

	if _, err := Commit(dir, testClock, time.Second, FinishSession("p", "conductor"), render); err != nil {
		t.Fatal(err)
	}
	s, _ := Load(dir)
	if f := s.Releases[0].Conductor.Finished; f == nil || !f.Equal(testClock) {
		t.Errorf("finished = %v", f)
	}
	if _, err := Commit(dir, testClock, time.Second, FinishSession("p", "pass:x"), render); err == nil {
		t.Error("a pass was finished as a session")
	}

	if _, err := Commit(dir, testClock, time.Second, EndSession("p", "pass:x"), render); err != nil {
		t.Fatal(err)
	}
	s, _ = Load(dir)
	if e := s.Releases[0].Passes[0].Session.Ended; e == nil || !e.Equal(testClock) {
		t.Errorf("ended = %v", e)
	}
	if _, err := Commit(dir, testClock, time.Second, FinishSession("p", "conductor"), render); err != nil {
		t.Fatalf("an unended conductor could not be finished again: %v", err)
	}
}

func TestStartNamesTheLiveRecord(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("building", goodSession, ""))
	_, err := Commit(dir, testClock, time.Second, StartPass("p", "x"), render)
	want := `pass "x" is building and cannot start: session p-one is live`
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func TestStopDetailClearsOnResume(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("building", "", ""))
	if _, err := Commit(dir, testClock, time.Second, StopOnLaunchFailure("p", "x", "claude --bg failed"), render); err != nil {
		t.Fatal(err)
	}
	s, _ := Load(dir)
	p := s.Releases[0].Passes[0]
	if p.State != Stopped || p.StopReason != StopLaunchFailed || p.StopDetail != "claude --bg failed" {
		t.Fatalf("pass = %+v", p)
	}
	writeFile(t, dir, UsageFile, `{"format":1,"percent":10,"reset_text":"x","read_at":"2026-11-20T22:58:00Z"}`)
	if _, err := Commit(dir, testClock.Add(time.Minute), time.Second, StartPass("p", "x"), render); err != nil {
		t.Fatal(err)
	}
	s, _ = Load(dir)
	if p := s.Releases[0].Passes[0]; p.State != Building || p.StopDetail != "" {
		t.Errorf("pass after the resume = %+v", p)
	}
}

func TestRecordCheckWithNoSession(t *testing.T) {
	dir, _ := loadDir(t, sessionRelease("building", "", ""))
	d := Decision{Project: "p", Record: "pass:x", Session: "p-x"}
	for _, decision := range []string{"launch", "wait_usage"} {
		d.Decision = decision
		if _, err := Commit(dir, testClock, time.Second, RecordCheck([]Recorded{{Decision: d}}), render); err != nil {
			t.Fatalf("%s: %v", decision, err)
		}
		s, _ := Load(dir)
		if len(s.Check.Decisions) != 1 || s.Check.Decisions[0].Decision != decision || s.Releases[0].Passes[0].Session != nil {
			t.Errorf("%s: check = %+v", decision, s.Check)
		}
	}
	d.Decision = "relaunch"
	if _, err := Commit(dir, testClock, time.Second, RecordCheck([]Recorded{{Decision: d}}), render); err == nil {
		t.Error("a relaunch for a pass with no record was accepted")
	}
}
