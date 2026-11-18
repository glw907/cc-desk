package state

import (
	"slices"
	"strings"
	"testing"
	"time"
)

// startState is ruleState with a readable reading and the named passes
// already building.
func startState(percent int, building ...string) (*State, *ReleaseFile) {
	s, rf := ruleState()
	s.Usage = &Usage{Format: Format, Percent: percent, ResetText: "Resets 5pm", ReadAt: at(8, 0)}
	for _, name := range building {
		p := addPass(rf, name, "r1", 10, Building)
		t0 := at(8, 0)
		p.Started = &t0
	}
	return s, rf
}

func TestStartOutcomes(t *testing.T) {
	tests := []struct {
		name       string
		percent    int
		building   []string
		wantState  PassState
		wantLimits []string
	}{
		{"started", 40, nil, Building, nil},
		{"waiting on the two-pass limit", 40, []string{"b1", "b2"}, Waiting, []string{LimitTwoPass}},
		{"waiting on usage", 100, nil, Waiting, []string{LimitUsage}},
		{"waiting on both", 100, []string{"b1", "b2"}, Waiting, []string{LimitTwoPass, LimitUsage}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := startState(tt.percent, tt.building...)
			addPass(rf, "x", "r1", 10, NotStarted)
			if err := StartPass("alpha", "x")(s, at(9, 0)); err != nil {
				t.Fatal(err)
			}
			p, _ := rf.pass("x")
			if p.State != tt.wantState {
				t.Fatalf("state = %s, want %s", p.State, tt.wantState)
			}
			if tt.wantState == Building {
				if p.Started == nil || !p.Started.Equal(at(9, 0)) || p.Wait != nil || !s.passEvent {
					t.Errorf("started %v wait %v passEvent %v", p.Started, p.Wait, s.passEvent)
				}
				return
			}
			if p.Started != nil || s.passEvent {
				t.Errorf("a waiting start set started %v or a pass event", p.Started)
			}
			if p.Wait == nil || !p.Wait.Since.Equal(at(9, 0)) || !slices.Equal(p.Wait.Limits, tt.wantLimits) {
				t.Errorf("wait = %+v, want since 9:00 and %q", p.Wait, tt.wantLimits)
			}
			if !rf.Dirty {
				t.Error("release not marked dirty")
			}
		})
	}

	t.Run("refused with the exact names", func(t *testing.T) {
		s, rf := startState(40)
		addPass(rf, "a", "r1", 10, NotStarted)
		x := addPass(rf, "x", "r1", 10, NotStarted, "a")
		x.Questions = []string{"q1"}
		err := StartPass("alpha", "x")(s, at(9, 0))
		if !isRefusal(err) {
			t.Fatalf("error = %v, want a refusal", err)
		}
		want := "cannot start alpha/x: " + BreakUnfinished + " (pass a is not done); " + BreakQuestion + " (q1)"
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err, want)
		}
		if rf.Dirty || x.State != NotStarted || s.passEvent {
			t.Error("a refusal changed the state")
		}
	})
}

func TestStartOnGuess(t *testing.T) {
	build := func() (*State, *ReleaseFile) {
		s, rf := startState(40)
		p := addPass(rf, "x", "r1", 100, NotStarted)
		p.Questions = []string{"q1"}
		return s, rf
	}
	t.Run("records the question, the recommendation, and the clock", func(t *testing.T) {
		s, rf := build()
		if err := StartOnGuess("alpha", "x", "q1")(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		p, _ := rf.pass("x")
		if p.State != Building || p.Guess == nil || p.Guess.Question != "q1" || p.Guess.Recommendation != "a" || !p.Guess.At.Equal(at(9, 0)) {
			t.Errorf("pass = %+v guess %+v", p, p.Guess)
		}
	})
	t.Run("an unknown question is a plain error", func(t *testing.T) {
		s, _ := build()
		err := StartOnGuess("alpha", "x", "nope")(s, at(9, 0))
		if err == nil || isRefusal(err) || !strings.Contains(err.Error(), `question "nope"`) {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("a question the pass does not rest on is refused without a rule name", func(t *testing.T) {
		s, _ := build()
		err := StartOnGuess("alpha", "x", "q2")(s, at(9, 0))
		if !isRefusal(err) || !strings.Contains(err.Error(), `"q2"`) {
			t.Fatalf("error = %v", err)
		}
		for _, n := range []string{BreakUnfinished, BreakQuestion, BreakOnePass, BreakMinutes} {
			if strings.Contains(err.Error(), n) {
				t.Errorf("error names %q: %v", n, err)
			}
		}
	})
	t.Run("an answered question is refused", func(t *testing.T) {
		s, rf := build()
		rf.Questions[0].Answer = &Answer{Ruling: "g1"}
		if err := StartOnGuess("alpha", "x", "q1")(s, at(9, 0)); !isRefusal(err) || !strings.Contains(err.Error(), "already answered") {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("another unanswered question is still named", func(t *testing.T) {
		s, rf := build()
		p, _ := rf.pass("x")
		p.Questions = []string{"q1", "q2"}
		err := StartOnGuess("alpha", "x", "q1")(s, at(9, 0))
		if !isRefusal(err) || !strings.Contains(err.Error(), BreakQuestion+" (q2)") {
			t.Errorf("error = %v", err)
		}
	})
	t.Run("a guess start that waits counts toward 300", func(t *testing.T) {
		s, rf := startState(100)
		p := addPass(rf, "x", "r1", 200, NotStarted)
		p.Questions = []string{"q1"}
		if err := StartOnGuess("alpha", "x", "q1")(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		if p.State != Waiting || p.Guess == nil || s.OpenGuessMinutes() != 200 {
			t.Errorf("state %s guess %+v minutes %d", p.State, p.Guess, s.OpenGuessMinutes())
		}
		y := addPass(rf, "y", "r1", 150, NotStarted)
		y.Questions = []string{"q2"}
		err := StartOnGuess("alpha", "y", "q2")(s, at(9, 5))
		if !isRefusal(err) || !strings.Contains(err.Error(), BreakMinutes) {
			t.Errorf("error = %v", err)
		}
	})
}

func TestStartWaitingPass(t *testing.T) {
	s, rf := startState(100)
	addPass(rf, "x", "r1", 10, NotStarted)
	if err := StartPass("alpha", "x")(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	err := StartPass("alpha", "x")(s, at(9, 5))
	if !isRefusal(err) || !strings.Contains(err.Error(), "is waiting and cannot start") {
		t.Errorf("error = %v", err)
	}
}

func TestSettle(t *testing.T) {
	waitOn := func(rf *ReleaseFile, name string, since time.Time, limits ...string) *Pass {
		p := addPass(rf, name, "r1", 10, Waiting)
		p.Wait = &Wait{Since: since, Limits: limits}
		return p
	}
	t.Run("one free slot starts the oldest and keeps the other", func(t *testing.T) {
		s, rf := startState(40, "b1", "b2")
		late := waitOn(rf, "late", at(9, 30), LimitTwoPass)
		early := waitOn(rf, "early", at(9, 10), LimitTwoPass)
		rf.Passes[0].State, rf.Passes[0].Stops = Done, nil
		fin, mins := at(8, 30), 30
		rf.Passes[0].Finished, rf.Passes[0].ActualMinutes = &fin, &mins
		settle(s, at(10, 0))
		if early.State != Building || !early.Started.Equal(at(10, 0)) || early.Wait != nil || !s.passEvent {
			t.Errorf("early = %+v passEvent %v", early, s.passEvent)
		}
		if late.State != Waiting || late.Wait == nil || !late.Wait.Since.Equal(at(9, 30)) {
			t.Errorf("late = %+v, want it still waiting since 9:30", late)
		}
	})
	t.Run("a pass waiting on usage alone starts when the reading drops", func(t *testing.T) {
		s, rf := startState(99)
		p := waitOn(rf, "x", at(9, 0), LimitUsage)
		settle(s, at(10, 0))
		if p.State != Building {
			t.Errorf("state = %s", p.State)
		}
	})
	t.Run("a pass still held gets its limits updated", func(t *testing.T) {
		s, rf := startState(100)
		p := waitOn(rf, "x", at(9, 0), LimitTwoPass)
		rf.Dirty = false
		settle(s, at(10, 0))
		if p.State != Waiting || !slices.Equal(p.Wait.Limits, []string{LimitUsage}) || !rf.Dirty {
			t.Errorf("pass = %+v wait %+v dirty %v", p, p.Wait, rf.Dirty)
		}
		rf.Dirty = false
		settle(s, at(10, 5))
		if rf.Dirty {
			t.Error("an unchanged limit list marked the release dirty")
		}
	})
	t.Run("a resumed stopped pass ignores both limits", func(t *testing.T) {
		s, rf := startState(100, "b1", "b2")
		p := addPass(rf, "x", "r1", 10, Stopped)
		t0 := at(8, 0)
		p.Started, p.StopReason = &t0, StopScope
		p.Stops = []StopInterval{{Start: at(8, 30)}}
		if err := StartPass("alpha", "x")(s, at(9, 0)); err != nil {
			t.Fatal(err)
		}
		if p.State != Building || s.Building() != 3 {
			t.Errorf("state %s, building %d", p.State, s.Building())
		}
	})
}
