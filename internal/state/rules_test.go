package state

import (
	"slices"
	"strings"
	"testing"
)

// ruleState builds one live release, alpha, with one designed round r1 and
// a second round r2 that comes after r1. Each test adjusts it.
func ruleState() (*State, *ReleaseFile) {
	rf := &ReleaseFile{Path: "releases/alpha.json", Release: Release{
		Project: "alpha", Name: "v1", Repo: "glw907/alpha",
		Rounds: []Round{
			{Name: "r1", Summary: "s", NeedsGeoff: true, DesignWritten: true},
			{Name: "r2", Summary: "s", NeedsGeoff: true, DesignWritten: true, After: []string{"r1"}},
		},
		Questions: []Question{
			{ID: "q1", Text: "t", Options: []string{"a", "b"}, Recommended: "a"},
			{ID: "q2", Text: "t", Options: []string{"a", "b"}, Recommended: "a"},
		},
	}}
	s := &State{
		Releases: []*ReleaseFile{rf},
		Settings: Settings{Format: Format, UsageStopPercent: DefaultUsageStop},
	}
	return s, rf
}

func addPass(rf *ReleaseFile, name, round string, minutes int, st PassState, after ...string) *Pass {
	rf.Passes = append(rf.Passes, Pass{Name: name, Round: round, PlannedMinutes: minutes, State: st, After: after})
	return &rf.Passes[len(rf.Passes)-1]
}

func names(bs []Broken) []string {
	var out []string
	for _, b := range bs {
		out = append(out, b.Name)
	}
	return out
}

func wantNames(t *testing.T, got []Broken, want ...string) {
	t.Helper()
	if g := names(got); !slices.Equal(g, want) {
		t.Errorf("names = %q, want %q", g, want)
	}
}

func openGuess(p *Pass, q string) {
	t := at(9, 0)
	p.Guess = &Guess{Question: q, Recommendation: "a", At: t}
}

func TestRulesBefore(t *testing.T) {
	t.Run("an after pass not done", func(t *testing.T) {
		s, rf := addAfter(Building)
		wantNames(t, s.Breaks(rf, &rf.Passes[1], ""), BreakUnfinished)
	})
	t.Run("an after pass done", func(t *testing.T) {
		s, rf := addAfter(Done)
		wantNames(t, s.Breaks(rf, &rf.Passes[1], ""))
		if !s.Free(rf, &rf.Passes[1]) {
			t.Error("not free")
		}
	})
	t.Run("an after round with no passes", func(t *testing.T) {
		s, rf := ruleState()
		addPass(rf, "x", "r2", 10, NotStarted)
		got := s.Breaks(rf, &rf.Passes[0], "")
		wantNames(t, got, BreakUnfinished)
		if !strings.Contains(got[0].Detail, "round r1 has no passes") {
			t.Errorf("detail = %q", got[0].Detail)
		}
	})
	t.Run("an after round not done", func(t *testing.T) {
		s, rf := ruleState()
		addPass(rf, "a", "r1", 10, Building)
		addPass(rf, "x", "r2", 10, NotStarted)
		wantNames(t, s.Breaks(rf, &rf.Passes[1], ""), BreakUnfinished)
	})
	t.Run("through a pass's own after list", func(t *testing.T) {
		s, rf := ruleState()
		addPass(rf, "a", "r1", 10, NotStarted)
		addPass(rf, "b", "r1", 10, Done, "a")
		addPass(rf, "x", "r1", 10, NotStarted, "b")
		got := s.Breaks(rf, &rf.Passes[2], "")
		wantNames(t, got, BreakUnfinished)
		if !strings.Contains(got[0].Detail, "pass a is not done") {
			t.Errorf("detail = %q", got[0].Detail)
		}
	})
	t.Run("an undesigned round", func(t *testing.T) {
		s, rf := ruleState()
		rf.Rounds[0].DesignWritten = false
		addPass(rf, "x", "r1", 10, NotStarted)
		wantNames(t, s.Breaks(rf, &rf.Passes[0], ""), BreakUnfinished)
	})
}

func addAfter(st PassState) (*State, *ReleaseFile) {
	s, rf := ruleState()
	addPass(rf, "a", "r1", 10, st)
	addPass(rf, "x", "r1", 10, NotStarted, "a")
	return s, rf
}

func TestRulesQuestions(t *testing.T) {
	s, rf := ruleState()
	p := addPass(rf, "x", "r1", 10, NotStarted)
	p.Questions = []string{"q1", "q2"}
	got := s.Breaks(rf, p, "")
	wantNames(t, got, BreakQuestion)
	if got[0].Detail != "q1, q2" {
		t.Errorf("detail = %q", got[0].Detail)
	}
	wantNames(t, s.Breaks(rf, p, "q1"), BreakQuestion)
	if d := s.Breaks(rf, p, "q1")[0].Detail; d != "q2" {
		t.Errorf("guess detail = %q, want q2", d)
	}
	rf.Questions[1].Answer = &Answer{Ruling: "g1"}
	wantNames(t, s.Breaks(rf, p, "q1"))
	if s.Free(rf, p) {
		t.Error("a plain start with q1 open is free")
	}
}

func TestRulesOnePass(t *testing.T) {
	build := func(st PassState, answered bool) (*State, *ReleaseFile) {
		s, rf := ruleState()
		g := addPass(rf, "g", "r1", 10, st)
		g.Questions = []string{"q1"}
		if st != NotStarted {
			t0 := at(9, 0)
			g.Started = &t0
		}
		openGuess(g, "q1")
		if answered {
			rf.Questions[0].Answer = &Answer{Ruling: "g1"}
		}
		addPass(rf, "x", "r1", 10, NotStarted, "g")
		return s, rf
	}
	for _, st := range []PassState{Building, Stopped, Waiting, Done} {
		s, rf := build(st, false)
		if st == Done {
			wantNames(t, s.Breaks(rf, &rf.Passes[1], ""), BreakOnePass)
		} else {
			wantNames(t, s.Breaks(rf, &rf.Passes[1], ""), BreakUnfinished, BreakOnePass)
		}
	}
	s, rf := build(Done, true)
	wantNames(t, s.Breaks(rf, &rf.Passes[1], ""))
}

func TestRulesGuessMinutes(t *testing.T) {
	tests := []struct {
		name    string
		open    int
		planned int
		guess   bool
		broken  bool
	}{
		{"200 plus 100", 200, 100, true, false},
		{"200 plus 150", 200, 150, true, true},
		{"nothing open plus 301", 0, 301, true, true},
		{"a plain start never breaks it", 200, 150, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := ruleState()
			if tt.open > 0 {
				g := addPass(rf, "g", "r1", tt.open, Waiting)
				g.Questions = []string{"q2"}
				openGuess(g, "q2")
			}
			p := addPass(rf, "x", "r1", tt.planned, NotStarted)
			p.Questions = []string{"q1"}
			q := ""
			if tt.guess {
				q = "q1"
			}
			var found bool
			for _, b := range s.Breaks(rf, p, q) {
				if b.Name == BreakMinutes {
					found = true
				}
			}
			if found != tt.broken {
				t.Errorf("breaks the 300-minute rule = %v, want %v", found, tt.broken)
			}
		})
	}
	s, rf := ruleState()
	g := addPass(rf, "g", "r1", 200, Waiting)
	g.Questions = []string{"q1"}
	openGuess(g, "q1")
	if got := s.OpenGuessMinutes(); got != 200 {
		t.Errorf("a waiting guess counts %d, want 200", got)
	}
	rf.Questions[0].Answer = &Answer{Ruling: "g1"}
	if got := s.OpenGuessMinutes(); got != 0 {
		t.Errorf("an answered guess counts %d, want 0", got)
	}
}

func TestRulesLimits(t *testing.T) {
	read := func(percent int, errText string) *Usage {
		return &Usage{Format: Format, Percent: percent, Error: errText, ReadAt: at(9, 0)}
	}
	tests := []struct {
		name  string
		build []PassState
		usage *Usage
		stop  int
		want  []string
	}{
		{"one building, reading 99", []PassState{Building}, read(99, ""), 100, nil},
		{"two building", []PassState{Building, Building}, read(99, ""), 100, []string{LimitTwoPass}},
		{"a stopped pass is not building", []PassState{Building, Stopped}, read(99, ""), 100, nil},
		{"reading 100", nil, read(100, ""), 100, []string{LimitUsage}},
		{"reading 87 under stop 85", nil, read(87, ""), 85, []string{LimitUsage}},
		{"an error reading", nil, read(0, "bad"), 100, []string{LimitUsage}},
		{"no reading", nil, nil, 100, []string{LimitUsage}},
		{"both", []PassState{Building, Building}, read(100, ""), 100, []string{LimitTwoPass, LimitUsage}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := ruleState()
			for i, st := range tt.build {
				p := addPass(rf, "b"+string(rune('0'+i)), "r1", 10, st)
				t0 := at(9, 0)
				p.Started = &t0
			}
			s.Usage, s.Settings.UsageStopPercent = tt.usage, tt.stop
			if got := s.Limits(); !slices.Equal(got, tt.want) {
				t.Errorf("limits = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckWait(t *testing.T) {
	t0 := at(9, 0)
	good := &Wait{Since: t0, Limits: []string{LimitUsage}}
	guess := &Guess{Question: "q1", Recommendation: "a", At: t0}
	tests := []struct {
		name  string
		setup func(*Pass)
		want  string // empty loads
	}{
		{"waiting with a wait", func(p *Pass) { p.State, p.Wait = Waiting, good }, ""},
		{"waiting with both limits", func(p *Pass) {
			p.State, p.Wait = Waiting, &Wait{Since: t0, Limits: []string{LimitTwoPass, LimitUsage}}
		}, ""},
		{"waiting with no wait", func(p *Pass) { p.State = Waiting }, "needs a"},
		{"waiting with no limits", func(p *Pass) { p.State, p.Wait = Waiting, &Wait{Since: t0} }, "needs a"},
		{"waiting with no since", func(p *Pass) { p.State, p.Wait = Waiting, &Wait{Limits: []string{LimitUsage}} }, "needs a"},
		{"waiting on an unknown limit", func(p *Pass) {
			p.State, p.Wait = Waiting, &Wait{Since: t0, Limits: []string{"rain"}}
		}, `limit "rain"`},
		{"not started with a wait", func(p *Pass) { p.Wait = good }, "carries a wait"},
		{"waiting with a guess", func(p *Pass) { p.State, p.Wait, p.Guess = Waiting, good, guess }, ""},
		{"not started with a guess", func(p *Pass) { p.Guess = guess }, "has not started"},
		{"building with a guess", func(p *Pass) { p.State, p.Started, p.Guess = Building, &t0, guess }, ""},
		{"a guess on a question it does not rest on", func(p *Pass) {
			p.State, p.Started, p.Guess = Building, &t0, &Guess{Question: "q2", At: t0}
		}, "does not rest on"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := Release{
				Project: "alpha", Name: "v1", Repo: "r",
				Rounds:    []Round{{Name: "r1", Summary: "s", NeedsGeoff: true}},
				Questions: []Question{{ID: "q1", Text: "t", Options: []string{"a"}, Recommended: "a"}, {ID: "q2", Text: "t", Options: []string{"a"}, Recommended: "a"}},
				Passes:    []Pass{{Name: "p", Summary: "s", Round: "r1", PlannedMinutes: 10, State: NotStarted, Questions: []string{"q1"}}},
			}
			tt.setup(&rel.Passes[0])
			err := validateRelease(&rel)
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			}
		})
	}
}
