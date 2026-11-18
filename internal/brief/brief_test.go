package brief

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

var clock = time.Date(2026, 11, 20, 23, 0, 0, 0, time.UTC)

func pass(name, round string, minutes int, st state.PassState, after []string, questions ...string) state.Pass {
	return state.Pass{Name: name, Round: round, PlannedMinutes: minutes, State: st, After: after, Questions: questions}
}

func question(id string) state.Question {
	return state.Question{ID: id, Text: id + "?", Options: []string{"yes", "no"}, Recommended: "yes"}
}

func release(rounds []state.Round, passes []state.Pass, qs ...state.Question) *state.ReleaseFile {
	return &state.ReleaseFile{Release: state.Release{
		Project: "p", Rounds: rounds, Passes: passes, Questions: qs,
	}}
}

func names(es []entry) []string {
	var out []string
	for _, e := range es {
		out = append(out, e.q.ID)
	}
	return out
}

func waitNames(e entry) []string {
	var out []string
	for _, p := range e.waits {
		out = append(out, p.Name)
	}
	return out
}

func TestBriefHoldsUp(t *testing.T) {
	r1 := state.Round{Name: "r1"}
	r2 := state.Round{Name: "r2", After: []string{"r1"}}
	tests := []struct {
		name    string
		rounds  []state.Round
		passes  []state.Pass
		wait    []string
		minutes int
	}{
		{"directly", []state.Round{r1},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q")}, []string{"a"}, 10},
		{"through a pass", []state.Round{r1},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q"), pass("b", "r1", 20, state.NotStarted, []string{"a"}), pass("c", "r1", 5, state.NotStarted, nil)},
			[]string{"a", "b"}, 30},
		{"through a chain", []state.Round{r1},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q"), pass("b", "r1", 20, state.NotStarted, []string{"a"}), pass("c", "r1", 40, state.NotStarted, []string{"b"})},
			[]string{"a", "b", "c"}, 70},
		{"through a round", []state.Round{r1, r2},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q"), pass("b", "r2", 25, state.NotStarted, nil)},
			[]string{"a", "b"}, 35},
		{"a pass counted once", []state.Round{r1, r2},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q"), pass("b", "r1", 10, state.NotStarted, []string{"a"}, "q"), pass("c", "r2", 30, state.NotStarted, []string{"a", "b"})},
			[]string{"a", "b", "c"}, 50},
		{"a done pass left out", []state.Round{r1},
			[]state.Pass{pass("a", "r1", 10, state.NotStarted, nil, "q"), pass("b", "r1", 20, state.Done, []string{"a"})},
			[]string{"a"}, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &state.State{Releases: []*state.ReleaseFile{release(tt.rounds, tt.passes, question("q"))}}
			es := waiting(s, "")
			if len(es) != 1 {
				t.Fatalf("waiting = %v, want q", names(es))
			}
			if got := waitNames(es[0]); !slices.Equal(got, tt.wait) || es[0].minutes != tt.minutes {
				t.Errorf("waits %v for %d minutes, want %v for %d", got, es[0].minutes, tt.wait, tt.minutes)
			}
		})
	}
}

func TestBriefOrderAndTieBreak(t *testing.T) {
	r := []state.Round{{Name: "r1"}}
	rf := release(r, []state.Pass{
		pass("small", "r1", 10, state.NotStarted, nil, "q-small"),
		pass("big", "r1", 50, state.NotStarted, nil, "q-big"),
		pass("tie-b", "r1", 20, state.NotStarted, nil, "q-b"),
		pass("tie-a", "r1", 20, state.NotStarted, nil, "q-a"),
	}, question("q-small"), question("q-big"), question("q-b"), question("q-a"))
	got := names(waiting(&state.State{Releases: []*state.ReleaseFile{rf}}, ""))
	want := []string{"q-big", "q-a", "q-b", "q-small"}
	if !slices.Equal(got, want) {
		t.Errorf("order = %v, want %v (most minutes first, then the id)", got, want)
	}
}

func TestBriefMergeQuestionBlock(t *testing.T) {
	r := []state.Round{{Name: "b2"}}
	p := pass("b4", "b2", 60, state.Building, nil, "merge-b4-1")
	p.Summary = "Add the report verbs"
	p.PullRequest = &state.PullRequest{Number: 12, Commit: "3f2a9c1", Question: "merge-b4-1"}
	q := question("merge-b4-1")
	q.Options = []string{"close", "merge", "fix again"}
	q.Recommended = "merge"
	rf := release(r, []state.Pass{p}, q)
	out := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock)
	for _, want := range []string{
		`1. p / b2, merge question merge-b4-1 on pull request 12, "Add the report verbs", at commit 3f2a9c1, holds up 60 planned minutes`,
		"last commit is still 3f2a9c1. Otherwise it asks again.",
		"   Recommended: merge.\n",
		"   Waits: b4 (60 minutes)\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("brief lacks %q:\n%s", want, out)
		}
	}
	m, f, c := strings.Index(out, "     merge:"), strings.Index(out, "     fix again:"), strings.Index(out, "     close:")
	if m < 0 || f < m || c < f {
		t.Errorf("answers are not in the order merge, fix again, close:\n%s", out)
	}

	q.Causes = map[string]string{"merge": "Its own words."}
	rf.Questions[0] = q
	if out := Build(&state.State{Releases: []*state.ReleaseFile{rf}}, clock); !strings.Contains(out, "merge: Its own words.") {
		t.Errorf("a question's own causes were not used:\n%s", out)
	}
}

func TestBriefOpeningAndEmptySections(t *testing.T) {
	out := Build(&state.State{}, clock)
	for _, want := range []string{
		"Brief at 2026-11-20T23:00:00Z\n",
		"Next question forecast: none forecast\n",
		"Waiting: nothing\n",
		"Could come back in the next 6 hours:\n  Nothing.\n",
		"No desk session is recorded, so the last 24 hours, since 2026-11-19T23:00:00Z:\n  Nothing.\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("empty brief lacks %q:\n%s", want, out)
		}
	}

	rf := release([]state.Round{{Name: "r1"}}, []state.Pass{pass("a", "r1", 10, state.Building, nil)})
	rf.Questions = nil
	s := &state.State{Releases: []*state.ReleaseFile{rf}}
	if out := Build(s, clock); !strings.Contains(out, "Waiting: nothing\n") {
		t.Errorf("a desk with nothing to ask:\n%s", out)
	}
}

func TestBriefSinceAndComeBack(t *testing.T) {
	line := func(h int, outcome, reason string) state.CheckLogLine {
		return state.CheckLogLine{Format: 1, Time: clock.Add(time.Duration(-h) * time.Hour), Project: "p", Record: "pass:x", Decision: "launch", Outcome: outcome, Reason: reason}
	}
	s := &state.State{
		Desks:    state.Desks{Format: 1, Desks: []state.Desk{{SessionID: "d", PID: 1, Started: clock.Add(-3 * time.Hour)}}},
		CheckLog: []state.CheckLogLine{line(5, "before", ""), line(2, "after", "dropped connection"), line(1, "later", "")},
	}
	out := Build(s, clock)
	if strings.Contains(out, "before") || !strings.Contains(out, "launch: after (reason: dropped connection)") || !strings.Contains(out, "launch: later\n") {
		t.Errorf("log lines since the desk's start:\n%s", out)
	}
	if !strings.Contains(out, "Since the last desk session, 2026-11-20T20:00:00Z:") {
		t.Errorf("heading:\n%s", out)
	}

	later := &state.State{Desks: state.Desks{Format: 1, Desks: []state.Desk{{SessionID: "d", PID: 1, Started: clock.Add(-30 * time.Minute)}}}, CheckLog: s.CheckLog}
	if out := Build(later, clock); !strings.Contains(out, "  Nothing.\n") {
		t.Errorf("no line since a recent desk:\n%s", out)
	}
}

func TestBriefBuildRound(t *testing.T) {
	rounds := []state.Round{{Name: "r1"}, {Name: "r2"}}
	rf := release(rounds, []state.Pass{
		pass("a", "r1", 10, state.NotStarted, nil, "q1"),
		pass("b", "r2", 20, state.NotStarted, nil, "q2"),
	}, question("q1"), question("q2"))
	s := &state.State{Releases: []*state.ReleaseFile{rf}}

	out, err := BuildRound(s, clock, "p", "r1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "No design recorded yet.") || !strings.Contains(out, "question q1") || strings.Contains(out, "q2") {
		t.Errorf("round brief:\n%s", out)
	}
	if strings.Contains(out, "Could come back") || strings.Contains(out, "Since the last") {
		t.Errorf("round brief holds the desk's sections:\n%s", out)
	}

	for _, tt := range []struct{ project, round, want string }{
		{"nope", "r1", `"nope"`},
		{"p", "r9", `"r9"`},
	} {
		if _, err := BuildRound(s, clock, tt.project, tt.round); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("BuildRound(%s, %s) error = %v, want one naming %s", tt.project, tt.round, err, tt.want)
		}
	}
}
