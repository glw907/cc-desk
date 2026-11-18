package snapshot

import (
	"slices"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// asking returns a release with one question and one pass resting on it.
// The pass is preceded by blocker minutes of work in the same release.
func asking(project string, blocker int, answered bool) *state.ReleaseFile {
	q := state.Question{ID: project + "-q", Text: "t", Options: []string{"x", "y"}, Recommended: "x"}
	if answered {
		q.Answer = &state.Answer{Carried: &state.Carried{Words: "w", Date: "2026-11-14", Source: "chat"}}
	}
	rf := releaseOf(project, []state.Round{
		{Name: "r1", DesignWritten: true, NeedsGeoff: true},
		{Name: "r2", DesignWritten: true, NeedsGeoff: true},
	}, todo("w1", "r1", max(blocker, 1)), todo("ask", "r2", 30, "w1"))
	rf.Passes[1].Questions = []string{q.ID}
	rf.Questions = []state.Question{q}
	if blocker == 0 {
		rf.Passes = rf.Passes[1:]
		rf.Passes[0].After = nil
	}
	return rf
}

func sessionOf(t *testing.T, now time.Time, rfs ...*state.ReleaseFile) *Snapshot {
	t.Helper()
	return Build(&state.State{Releases: rfs}, now)
}

func TestNextSessionReadyNow(t *testing.T) {
	now := utcTime(20, 23, 0)
	snap := sessionOf(t, now, asking("a", 0, false))
	n := snap.NextSession
	if n.State == nil || *n.State != "ready_now" || n.QuestionCount == nil || *n.QuestionCount != 1 {
		t.Fatalf("next_session = %+v", n)
	}
	wantTime(t, "time", n.Time, now)
	if !slices.Equal(n.Rounds, []string{"a/r2"}) || !slices.Equal(n.Projects, []string{"a"}) {
		t.Errorf("rounds %v projects %v", n.Rounds, n.Projects)
	}
	if len(snap.WaitsForGeoff) != 1 || snap.WaitsForGeoff[0] != (WaitEntry{Project: "a", Round: "r2", Questions: 1}) {
		t.Errorf("waits_for_geoff = %+v", snap.WaitsForGeoff)
	}
}

func TestNextSessionScheduled(t *testing.T) {
	now := utcTime(20, 23, 0)
	snap := sessionOf(t, now, asking("a", 60, false))
	n := snap.NextSession
	if n.State == nil || *n.State != "scheduled" {
		t.Fatalf("state = %v, want scheduled", n.State)
	}
	wantTime(t, "time", n.Time, now.Add(60*time.Minute))
	if n.Rough == nil || !*n.Rough || n.Range != nil {
		t.Errorf("rough %v range %v; want rough with no range", n.Rough, n.Range)
	}
	if len(snap.WaitsForGeoff) != 0 {
		t.Errorf("waits_for_geoff = %+v, want none", snap.WaitsForGeoff)
	}
}

func TestNextSessionOrder(t *testing.T) {
	now := utcTime(20, 23, 0)
	// b's question comes due at the same time as a's; release order breaks the tie.
	snap := sessionOf(t, now, asking("a", 60, false), asking("b", 60, false))
	if got := snap.NextSession.Rounds; !slices.Equal(got, []string{"a/r2", "b/r2"}) {
		t.Errorf("rounds = %v", got)
	}
	snap = sessionOf(t, now, asking("a", 90, false), asking("b", 60, false))
	wantTime(t, "earliest", snap.NextSession.Time, now.Add(60*time.Minute))
}

func TestNextSessionNotDesignedSets(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := asking("a", 0, true)
	rf.Rounds = append(rf.Rounds, state.Round{Name: "later", NeedsGeoff: true, After: []string{"r2"}})
	snap := sessionOf(t, now, rf)
	n := snap.NextSession
	if !slices.Equal(n.NotDesigned, []string{"a/later"}) || n.State == nil || *n.State != "scheduled" {
		t.Errorf("not_designed %v state %v", n.NotDesigned, n.State)
	}
	if n.QuestionCount == nil || *n.QuestionCount != 0 || len(n.Rounds) != 0 || len(n.Projects) != 0 {
		t.Errorf("count %v rounds %v projects %v", n.QuestionCount, n.Rounds, n.Projects)
	}
	wantTime(t, "time", n.Time, now.Add(30*time.Minute))
}

func TestNextSessionWaitingOnDesign(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := releaseOf("a", []state.Round{{Name: "r", DesignWritten: true, NeedsGeoff: true},
		{Name: "later", NeedsGeoff: true, After: []string{"r"}}},
		donePass("d", "r", 10, 10, now.Add(-time.Hour)))
	n := sessionOf(t, now, rf).NextSession
	if !slices.Equal(n.WaitingOnDesign, []string{"a/later"}) || len(n.NotDesigned) != 0 {
		t.Errorf("waiting_on_design %v not_designed %v", n.WaitingOnDesign, n.NotDesigned)
	}
	if n.State == nil || *n.State != "ready_now" || n.Time != nil || n.QuestionCount == nil || *n.QuestionCount != 0 {
		t.Errorf("state %v time %v count %v; want ready_now, null, 0", n.State, n.Time, n.QuestionCount)
	}
}

func TestNextSessionNothingWaiting(t *testing.T) {
	n := sessionOf(t, utcTime(20, 23, 0), asking("a", 0, true)).NextSession
	if n.State == nil || *n.State != "nothing_waiting" || n.Time != nil || n.Range != nil || n.Rough != nil {
		t.Errorf("next_session = %+v", n)
	}
}

func TestNextSessionStopped(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := asking("a", 0, false)
	b := building("w", "r1", 10, now.Add(-20*time.Minute))
	b.State, b.StopReason = state.Stopped, state.StopScope
	b.Stops = []state.StopInterval{{Start: now.Add(-10 * time.Minute)}}
	rf.Passes = append(rf.Passes, b)
	if got := sessionOf(t, now, rf).NextSession.State; got == nil || *got != "needs_you_now" {
		t.Errorf("state = %v, want needs_you_now", got)
	}
}

func TestRenderForecast(t *testing.T) {
	now := utcTime(20, 23, 0)
	_, fc, err := Render(&state.State{Releases: []*state.ReleaseFile{asking("a", 60, false)}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if fc.Project != "a" || fc.Round != "r2" {
		t.Errorf("forecast names %s/%s, want a/r2", fc.Project, fc.Round)
	}
	wantTime(t, "forecast time", fc.Time, now.Add(60*time.Minute))

	_, fc, err = Render(&state.State{Releases: []*state.ReleaseFile{asking("a", 0, true)}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if fc.Project != "" || fc.Round != "" || fc.Time != nil {
		t.Errorf("forecast with nothing waiting = %+v, want no names and no time", fc)
	}
}

// A building predecessor that has used its paced length at the middle pace
// but not at the high ratio leaves the time at now. The range is now to now.
func TestNextSessionReadyNowRange(t *testing.T) {
	now := utcTime(21, 12, 0)
	rf := worked("a")
	rf.Passes = rf.Passes[:12]
	rf.Passes = append(rf.Passes,
		building("p13", "r2", 100, now.Add(-125*time.Minute)),
		todo("ask", "r2", 30, "p13"))
	rf.Passes[13].Questions = []string{"q"}
	rf.Questions = []state.Question{{ID: "q", Text: "t", Options: []string{"x", "y"}, Recommended: "x"}}
	n := sessionOf(t, now, rf).NextSession
	if n.State == nil || *n.State != "ready_now" || n.Range == nil {
		t.Fatalf("next_session = %+v", n)
	}
	wantTime(t, "time", n.Time, now)
	wantTime(t, "range low", &n.Range.Low, now)
	wantTime(t, "range high", &n.Range.High, now)
}

// A pass event with only a question no pass rests on writes a null-time line.
func TestRenderForecastOrphanQuestion(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := asking("a", 0, true)
	rf.Questions = append(rf.Questions, state.Question{ID: "loose", Text: "t", Options: []string{"x"}, Recommended: "x"})
	_, fc, err := Render(&state.State{Releases: []*state.ReleaseFile{rf}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if fc.Project != "" || fc.Round != "" || fc.Time != nil {
		t.Errorf("forecast = %+v, want no names and no time", fc)
	}
}

// A question no pass rests on has no round and is needed now.
func TestNextSessionQuestionWithoutPass(t *testing.T) {
	now := utcTime(20, 23, 0)
	rf := asking("a", 0, true)
	rf.Questions = append(rf.Questions, state.Question{ID: "loose", Text: "t", Options: []string{"x"}, Recommended: "x"})
	snap := sessionOf(t, now, rf)
	n := snap.NextSession
	if n.State == nil || *n.State != "ready_now" || *n.QuestionCount != 1 || len(n.Rounds) != 0 || len(snap.WaitsForGeoff) != 0 {
		t.Errorf("next_session = %+v waits %+v", n, snap.WaitsForGeoff)
	}
	if !slices.Equal(n.Projects, []string{"a"}) {
		t.Errorf("projects = %v", n.Projects)
	}
}
