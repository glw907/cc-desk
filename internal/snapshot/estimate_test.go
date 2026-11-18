package snapshot

import (
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

func utcTime(day, h, m int) time.Time { return time.Date(2026, 11, day, h, m, 0, 0, time.UTC) }

// donePass is a finished pass that ended at the given time.
func donePass(name, round string, planned, actual int, end time.Time) state.Pass {
	start := end.Add(-time.Duration(actual) * time.Minute)
	return state.Pass{
		Name: name, Round: round, PlannedMinutes: planned, State: state.Done,
		Started: &start, Finished: &end, ActualMinutes: &actual,
	}
}

func todo(name, round string, planned int, after ...string) state.Pass {
	return state.Pass{Name: name, Round: round, PlannedMinutes: planned, State: state.NotStarted, After: after}
}

func building(name, round string, planned int, start time.Time, after ...string) state.Pass {
	p := todo(name, round, planned, after...)
	p.State, p.Started = state.Building, &start
	return p
}

func releaseOf(project string, rounds []state.Round, passes ...state.Pass) *state.ReleaseFile {
	return &state.ReleaseFile{
		Path:    "releases/" + project + ".json",
		Release: state.Release{Format: 1, Project: project, Name: "v1", Repo: "r", Rounds: rounds, Passes: passes},
	}
}

// worked is the design's worked case: twelve finished passes planned at
// 100 minutes, the two oldest taking 300, and three unfinished passes.
func worked(project string) *state.ReleaseFile {
	actual := []int{300, 300, 150, 80, 120, 120, 120, 120, 120, 120, 125, 125}
	var passes []state.Pass
	for i, a := range actual {
		passes = append(passes, donePass("p"+string(rune('a'+i)), "r1", 100, a, utcTime(8+i, 22, 0)))
	}
	passes = append(passes,
		todo("p13", "r2", 100),
		todo("p14", "r2", 200, "p13"),
		todo("p15", "r2", 60, "p14"))
	return releaseOf(project, []state.Round{
		{Name: "r1", DesignWritten: true, NeedsGeoff: true},
		{Name: "r2", DesignWritten: true, NeedsGeoff: true, After: []string{"r1"}},
	}, passes...)
}

func findPassOut(t *testing.T, r Release, name string) Pass {
	t.Helper()
	for _, rd := range r.Rounds {
		for _, p := range rd.Passes {
			if p.Name == name {
				return p
			}
		}
	}
	t.Fatalf("pass %q not in snapshot", name)
	return Pass{}
}

func wantTime(t *testing.T, what string, got *time.Time, want time.Time) {
	t.Helper()
	if got == nil || !got.Equal(want) {
		t.Errorf("%s = %v, want %s", what, got, want.Format(time.RFC3339))
	}
}

func TestEstimateWorkedCase(t *testing.T) {
	snap := Build(&state.State{Releases: []*state.ReleaseFile{worked("alpha")}}, utcTime(20, 23, 0))
	r := snap.Releases[0]
	if r.Pace == nil || *r.Pace != 1.2 || r.PaceLow == nil || *r.PaceLow != 0.8 || r.PaceHigh == nil || *r.PaceHigh != 1.5 {
		t.Errorf("pace = %v, %v, %v; want 1.2, 0.8, 1.5", r.Pace, r.PaceLow, r.PaceHigh)
	}
	if r.Rough == nil || *r.Rough {
		t.Errorf("rough = %v, want false", r.Rough)
	}
	p15 := findPassOut(t, r, "p15")
	if p15.MinutesBefore == nil || *p15.MinutesBefore != 300 {
		t.Errorf("p15 minutes_before = %v, want 300", p15.MinutesBefore)
	}
	wantTime(t, "p15 needed_time", p15.NeededTime, utcTime(21, 5, 0))
	p13 := findPassOut(t, r, "p13")
	if p13.MinutesBefore == nil || *p13.MinutesBefore != 0 {
		t.Errorf("p13 minutes_before = %v, want 0", p13.MinutesBefore)
	}
	wantTime(t, "p13 needed_time", p13.NeededTime, utcTime(20, 23, 0))
	wantTime(t, "p14 needed_time", findPassOut(t, r, "p14").NeededTime, utcTime(21, 1, 0))
}

func TestEstimatePaceLastTen(t *testing.T) {
	rounds := []state.Round{{Name: "r", DesignWritten: true}}
	live := releaseOf("a", rounds,
		donePass("n1", "r", 100, 200, utcTime(5, 12, 0)),
		donePass("n2", "r", 100, 100, utcTime(7, 12, 0)),
		donePass("n3", "r", 100, 100, utcTime(9, 12, 0)),
		donePass("n4", "r", 100, 100, utcTime(11, 12, 0)),
		donePass("n5", "r", 100, 100, utcTime(13, 12, 0)),
		donePass("n6", "r", 100, 100, utcTime(15, 12, 0)))
	old := releaseOf("a", rounds,
		donePass("o1", "r", 100, 400, utcTime(1, 12, 0)),
		donePass("o2", "r", 100, 100, utcTime(6, 12, 0)),
		donePass("o3", "r", 100, 100, utcTime(8, 12, 0)),
		donePass("o4", "r", 100, 100, utcTime(10, 12, 0)),
		donePass("o5", "r", 100, 100, utcTime(12, 12, 0)),
		donePass("o6", "r", 100, 100, utcTime(14, 12, 0)))
	old.Done, old.Path = true, "releases/done/old.json"
	other := releaseOf("b", rounds, donePass("x1", "r", 100, 900, utcTime(16, 12, 0)))

	got := paces(&state.State{Releases: []*state.ReleaseFile{live, other, old}})["a"]
	// Twelve finished passes; the last ten by finish time drop o1 and n1.
	if got.mid != (ratio{1000, 1000}) || got.low != (ratio{100, 100}) || got.high != (ratio{100, 100}) || got.rough {
		t.Errorf("pace = %+v, want 1000/1000 with both ends 100/100", got)
	}
	if _, ok := paces(&state.State{Releases: []*state.ReleaseFile{other}})["a"]; ok {
		t.Error("project a has a pace from another project's passes")
	}
}

func TestEstimateRough(t *testing.T) {
	rf := releaseOf("a", []state.Round{{Name: "r", DesignWritten: true}},
		donePass("n1", "r", 100, 300, utcTime(5, 12, 0)),
		donePass("n2", "r", 100, 300, utcTime(6, 12, 0)),
		todo("t", "r", 100))
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, utcTime(20, 23, 0))
	r := snap.Releases[0]
	if r.Pace == nil || *r.Pace != 1.0 || r.Rough == nil || !*r.Rough || r.PaceLow != nil || r.PaceHigh != nil {
		t.Errorf("release = pace %v rough %v ends %v %v; want 1.0, rough, no ends", r.Pace, r.Rough, r.PaceLow, r.PaceHigh)
	}
}

func TestEstimateBuilding(t *testing.T) {
	now := utcTime(20, 23, 0)
	rounds := []state.Round{{Name: "r", DesignWritten: true}}
	after := func(p state.Pass) Pass {
		rf := releaseOf("a", rounds, p, todo("next", "r", 100, p.Name))
		return findPassOut(t, Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now).Releases[0], "next")
	}

	// Pace 1.0 (rough): planned 100 minutes, 30 minutes building leaves 70.
	got := after(building("b", "r", 100, now.Add(-30*time.Minute)))
	wantTime(t, "building 30 minutes", got.NeededTime, now.Add(70*time.Minute))

	got = after(building("b", "r", 100, now.Add(-150*time.Minute)))
	wantTime(t, "building past its length", got.NeededTime, now)

	stopped := building("b", "r", 100, now.Add(-60*time.Minute))
	stopped.State = state.Stopped
	stopped.StopReason = state.StopScope
	closedEnd := now.Add(-40 * time.Minute)
	stopped.Stops = []state.StopInterval{
		{Start: now.Add(-50 * time.Minute), End: &closedEnd},
		{Start: now.Add(-10 * time.Minute)},
	}
	// 60 minutes since the start, less 10 closed and 10 open: 40 active.
	got = after(stopped)
	wantTime(t, "stopped pass", got.NeededTime, now.Add(60*time.Minute))
}

func TestEstimatePredecessors(t *testing.T) {
	now := utcTime(20, 23, 0)
	rounds := []state.Round{
		{Name: "r1", DesignWritten: true},
		{Name: "r2", DesignWritten: true, After: []string{"r1", "empty"}},
		{Name: "empty", PlannedMinutes: 30, After: []string{"r1"}},
		{Name: "r3", DesignWritten: true, After: []string{"r2"}},
	}
	rf := releaseOf("a", rounds,
		todo("a1", "r1", 10),
		todo("a2", "r1", 20, "a1"),
		todo("b1", "r2", 40, "a2"),
		todo("c1", "r3", 50, "b1", "a1"))
	// c1 comes after b1 and a1 directly, and through r3's round after r2:
	// b1, a2, a1 (once each), and the empty round's 30 once.
	snap := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, now)
	p := findPassOut(t, snap.Releases[0], "c1")
	if p.MinutesBefore == nil || *p.MinutesBefore != 10+20+40+30 {
		t.Errorf("c1 minutes_before = %v, want 100", p.MinutesBefore)
	}
	wantTime(t, "c1 needed_time", p.NeededTime, now.Add(100*time.Minute))

	var empty Round
	for _, rd := range snap.Releases[0].Rounds {
		if rd.Name == "empty" {
			empty = rd
		}
	}
	if empty.State != "not_designed" || empty.MinutesBefore == nil || *empty.MinutesBefore != 30 {
		t.Errorf("empty round = state %s minutes_before %v, want a not_designed round with 30", empty.State, empty.MinutesBefore)
	}
}
