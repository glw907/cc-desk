package snapshot

import (
	"slices"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

// closedPass is a pass a ruling closed: done, with minutes, and the mark.
func closedPass(name, round string, planned, actual int) state.Pass {
	p := donePass(name, round, planned, actual, utcTime(19, 22, 0))
	p.Closed = &state.Closed{Ruling: "r-0001", Time: utcTime(19, 22, 0)}
	return p
}

// closedRelease has one merged pass, the closed pass "c", and the pass "held"
// after it, resting on a question.
func closedRelease() *state.ReleaseFile {
	rf := releaseOf("a", []state.Round{{Name: "r", DesignWritten: true, NeedsGeoff: true}},
		donePass("m", "r", 100, 100, utcTime(18, 22, 0)),
		closedPass("c", "r", 100, 500),
		todo("held", "r", 50, "c"),
		todo("later", "r", 50, "held"))
	rf.Passes[2].Questions = []string{"a-q"}
	rf.Questions = []state.Question{{ID: "a-q", Text: "t", Options: []string{"x", "y"}, Recommended: "x"}}
	return rf
}

func TestClosedPassForecast(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := closedRelease()
	s := &state.State{Releases: []*state.ReleaseFile{rf}, Rulings: []state.Ruling{{ID: "r-0001"}}}
	snap := Build(s, now)
	rel := snap.Releases[0]

	for _, name := range []string{"held", "later"} {
		p := findPassOut(t, rel, name)
		if p.NeededTime != nil || p.MinutesBefore != nil {
			t.Errorf("%s: needed %v, minutes before %v, want null", name, p.NeededTime, p.MinutesBefore)
		}
		if p.Free == nil || *p.Free {
			t.Errorf("%s is free", name)
		}
	}
	held := findPassOut(t, rel, "held")
	if !slices.ContainsFunc(held.WhyNot, func(w string) bool { return strings.Contains(w, "was closed") }) {
		t.Errorf("held why_not = %v, want one that says was closed", held.WhyNot)
	}

	c := findPassOut(t, rel, "c")
	if c.Closed == nil || c.Closed.Ruling != "r-0001" || c.State != "done" {
		t.Errorf("closed pass = %+v", c)
	}
	if findPassOut(t, rel, "m").Closed != nil {
		t.Error("a merged pass reads closed")
	}

	// The question rests only on a pass that never starts, so no session
	// waits on it.
	if len(snap.WaitsForGeoff) != 0 || len(snap.NextSession.Rounds) != 0 || len(snap.NextSession.Projects) != 0 {
		t.Errorf("next_session = %+v, waits %+v", snap.NextSession, snap.WaitsForGeoff)
	}
}

func TestClosedPassRoundDone(t *testing.T) {
	rf := releaseOf("a", []state.Round{{Name: "r", DesignWritten: true}},
		donePass("m", "r", 100, 100, utcTime(18, 22, 0)), closedPass("c", "r", 100, 500))
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, utcTime(20, 23, 0))
	if got := snap.Releases[0].Rounds[0].State; got != "done" {
		t.Errorf("round state = %q, want done", got)
	}
}

func TestPaceLeavesOutClosedPass(t *testing.T) {
	now := utcTime(20, 23, 0)
	merged := func() []state.Pass {
		var ps []state.Pass
		for i := range 5 {
			ps = append(ps, donePass("m"+string(rune('a'+i)), "r", 100, 100, utcTime(10+i, 22, 0)))
		}
		return ps
	}
	pace := func(extra ...state.Pass) float64 {
		rf := releaseOf("a", []state.Round{{Name: "r", DesignWritten: true}}, append(merged(), extra...)...)
		return *Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now).Releases[0].Pace
	}
	want := pace()
	if got := pace(closedPass("c", "r", 100, 900)); got != want {
		t.Errorf("pace with a closed pass = %v, want %v", got, want)
	}
	if got := pace(donePass("slow", "r", 100, 900, utcTime(19, 22, 0))); got == want {
		t.Errorf("pace with a merged slow pass = %v, want it to move", got)
	}
}
