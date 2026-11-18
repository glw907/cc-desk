package state

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func loaded(t *testing.T) *State {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "releases/alpha.json", goodRelease)
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func at(h, m int) time.Time { return time.Date(2026, 11, 20, h, m, 0, 0, time.UTC) }

func isRefusal(err error) bool {
	_, ok := errors.AsType[*RefusedError](err)
	return ok
}

// TestPassEvents runs every verb against every from-state.
func TestPassEvents(t *testing.T) {
	stopped := func(p *Pass) {
		t0, t1 := at(9, 0), at(9, 30)
		p.State, p.Started, p.StopReason = Stopped, &t0, StopScope
		p.Stops = []StopInterval{{Start: t1}}
	}
	building := func(p *Pass) {
		t0 := at(9, 0)
		p.State, p.Started = Building, &t0
	}
	done := func(p *Pass) {
		t0, t1, n := at(9, 0), at(10, 0), 60
		p.State, p.Started, p.Finished, p.ActualMinutes = Done, &t0, &t1, &n
	}
	waiting := func(p *Pass) { p.State = Waiting }
	none := func(*Pass) {}

	tests := []struct {
		name  string
		setup func(*Pass)
		verb  string
		want  PassState // "" means a refusal
	}{
		{"start not_started", none, "start", Building},
		{"start waiting", waiting, "start", ""},
		{"start building", building, "start", ""},
		{"start stopped", stopped, "start", Building},
		{"start done", done, "start", ""},
		{"finish not_started", none, "finish", ""},
		{"finish waiting", waiting, "finish", ""},
		{"finish building", building, "finish", Done},
		{"finish stopped", stopped, "finish", ""},
		{"finish done", done, "finish", ""},
		{"stop not_started", none, "stop", ""},
		{"stop waiting", waiting, "stop", ""},
		{"stop building", building, "stop", Stopped},
		{"stop stopped", stopped, "stop", ""},
		{"stop done", done, "stop", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := loaded(t)
			tt.setup(&s.Releases[0].Passes[0])
			var ev Event
			switch tt.verb {
			case "start":
				ev = StartPass("alpha", "p1")
			case "finish":
				ev = FinishPass("alpha", "p1")
			case "stop":
				ev = StopPass("alpha", "p1", StopSecret)
			}
			err := ev(s, at(11, 0))
			if tt.want == "" {
				if !isRefusal(err) {
					t.Fatalf("error = %v, want a refusal", err)
				}
				if s.Releases[0].Dirty {
					t.Error("a refusal marked the release dirty")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := s.Releases[0].Passes[0].State; got != tt.want {
				t.Errorf("state = %s, want %s", got, tt.want)
			}
			if !s.Releases[0].Dirty {
				t.Error("release not marked dirty")
			}
		})
	}
}

func TestPassTimes(t *testing.T) {
	s := loaded(t)
	p := &s.Releases[0].Passes[0]
	if err := StartPass("alpha", "p1")(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	if p.Started == nil || !p.Started.Equal(at(9, 0)) {
		t.Errorf("started = %v", p.Started)
	}
	if err := StopPass("alpha", "p1", StopUsage)(s, at(9, 40)); err != nil {
		t.Fatal(err)
	}
	if p.StopReason != StopUsage || len(p.Stops) != 1 || !p.Stops[0].Start.Equal(at(9, 40)) || p.Stops[0].End != nil {
		t.Errorf("stop = %+v reason %s", p.Stops, p.StopReason)
	}
	if err := StartPass("alpha", "p1")(s, at(10, 10)); err != nil {
		t.Fatal(err)
	}
	if p.Stops[0].End == nil || !p.Stops[0].End.Equal(at(10, 10)) || p.StopReason != "" || !p.Started.Equal(at(9, 0)) {
		t.Errorf("resume left %+v reason %q started %v", p.Stops, p.StopReason, p.Started)
	}
	// 9:00 to 11:00 is 120 minutes, less the 30 stopped: 90.
	if err := FinishPass("alpha", "p1")(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if p.ActualMinutes == nil || *p.ActualMinutes != 90 {
		t.Errorf("actual = %v, want 90 with stopped time left out", p.ActualMinutes)
	}
	if !p.Finished.Equal(at(11, 0)) {
		t.Errorf("finished = %v", p.Finished)
	}
}

func TestActualMinutesRound(t *testing.T) {
	s := loaded(t)
	start := time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)
	if err := StartPass("alpha", "p1")(s, start); err != nil {
		t.Fatal(err)
	}
	if err := FinishPass("alpha", "p1")(s, start.Add(89*time.Second)); err != nil {
		t.Fatal(err)
	}
	if got := *s.Releases[0].Passes[0].ActualMinutes; got != 1 {
		t.Errorf("89 seconds rounded to %d minutes, want 1", got)
	}
}

func TestPassEventErrors(t *testing.T) {
	s := loaded(t)
	tests := []struct {
		name string
		ev   Event
		want string
	}{
		{"no project", StartPass("nope", "p1"), `project "nope" not found`},
		{"no pass", FinishPass("alpha", "nope"), `pass "nope" not found in project "alpha"`},
		{"bad reason", StopPass("alpha", "p1", "tired"), `stop reason "tired"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ev(s, at(9, 0))
			if err == nil || isRefusal(err) || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want a plain error holding %q", err, tt.want)
			}
		})
	}
}

func TestAddRelease(t *testing.T) {
	newRel := func(project string) *Release {
		r, err := DecodeRelease([]byte(strings.Replace(goodRelease, `"alpha"`, `"`+project+`"`, 1)))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	s := loaded(t)

	if err := AddRelease(newRel("beta"))(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	if got := s.Releases[len(s.Releases)-1]; got.Path != "releases/beta.json" || !got.Dirty {
		t.Errorf("added release = %+v", got)
	}
	if err := AddRelease(newRel("alpha"))(s, at(9, 0)); !isRefusal(err) {
		t.Errorf("a live project: error = %v, want a refusal", err)
	}
	if err := AddRelease(newRel("../evil"))(s, at(9, 0)); err == nil || isRefusal(err) {
		t.Errorf("a path-like project name: error = %v", err)
	}

	started := newRel("gamma")
	st := at(9, 0)
	started.Passes[0].State, started.Passes[0].Started = Building, &st
	if err := AddRelease(started)(s, at(9, 0)); err == nil || isRefusal(err) {
		t.Errorf("a started pass in a new release: error = %v", err)
	}

	ruled := newRel("delta")
	ruled.Questions = []Question{{ID: "q", Text: "t", Options: []string{"a"}, Recommended: "a", Answer: &Answer{Ruling: "x"}}}
	if err := AddRelease(ruled)(s, at(9, 0)); err == nil || isRefusal(err) {
		t.Errorf("a ruling answer in a new release: error = %v", err)
	}
}

func TestAddReleaseAfterDoneRelease(t *testing.T) {
	s := loaded(t)
	s.Releases[0].Done = true
	r, err := DecodeRelease([]byte(goodRelease))
	if err != nil {
		t.Fatal(err)
	}
	if err := AddRelease(r)(s, at(9, 0)); err != nil {
		t.Errorf("a project whose release is done refused: %v", err)
	}
}

const design = `{"format":1,
  "passes":[{"name":"p2","summary":"More.","round":"r1","planned_minutes":30,"state":"not_started","after":["p1"],"questions":["q1"]}],
  "questions":[{"id":"q1","text":"Which?","options":["a","b"],"recommended":"b"}]}`

func TestDesignRound(t *testing.T) {
	s := loaded(t)
	s.Releases[0].Rounds[0].DesignWritten = false
	d, err := DecodeDesign([]byte(design))
	if err != nil {
		t.Fatal(err)
	}
	if err := DesignRound("alpha", "r1", d)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	rf := s.Releases[0]
	if len(rf.Passes) != 2 || len(rf.Questions) != 1 || !rf.Rounds[0].DesignWritten || !rf.Dirty {
		t.Errorf("after design: %d passes, %d questions, written %v", len(rf.Passes), len(rf.Questions), rf.Rounds[0].DesignWritten)
	}
	if err := DesignRound("alpha", "r1", d)(s, at(9, 0)); !isRefusal(err) {
		t.Errorf("a second design: error = %v, want a refusal", err)
	}
}

func TestDesignRoundBadInputs(t *testing.T) {
	tests := []struct {
		name    string
		project string
		round   string
		json    string
		want    string
	}{
		{"no project", "nope", "r1", design, `project "nope" not found`},
		{"no round", "alpha", "r9", design, `round "r9" not found`},
		{"undefined question", "alpha", "r1", strings.Replace(design, `"questions":["q1"]`, `"questions":["q7"]`, 1), `question "q7", which does not exist`},
		{"unknown pass before", "alpha", "r1", strings.Replace(design, `"after":["p1"]`, `"after":["zz"]`, 1), `pass "zz", which does not exist`},
		{"other round", "alpha", "r1", strings.Replace(design, `"round":"r1"`, `"round":"r2"`, 1), `is in round "r2", not "r1"`},
		{"answered", "alpha", "r1", strings.Replace(design, `"recommended":"b"`, `"recommended":"b","answer":{"ruling":"x"}`, 1), "carries no answers"},
		{"duplicate pass", "alpha", "r1", strings.Replace(design, `"name":"p2"`, `"name":"p1"`, 1), `pass "p1" is defined twice`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := loaded(t)
			s.Releases[0].Rounds[0].DesignWritten = false
			d, err := DecodeDesign([]byte(tt.json))
			if err != nil {
				t.Fatal(err)
			}
			err = DesignRound(tt.project, tt.round, d)(s, at(9, 0))
			if err == nil || isRefusal(err) || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want a plain error holding %q", err, tt.want)
			}
			if s.Releases[0].Dirty || len(s.Releases[0].Passes) != 1 || s.Releases[0].Rounds[0].DesignWritten {
				t.Error("a failed design changed the release")
			}
		})
	}
}

func TestDecodeDesignStrict(t *testing.T) {
	for _, in := range []string{`{"format":1}`, `{"format":1,"passes":[],"x":1}`, `{"passes":[{}]}`, ``} {
		if _, err := DecodeDesign([]byte(in)); err == nil {
			t.Errorf("DecodeDesign(%q) succeeded", in)
		}
	}
}
