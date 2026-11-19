package sessions

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

func startWorld(checkout string, percent int, passes ...state.Pass) *state.State {
	for i := range passes {
		passes[i].Round = "a"
	}
	rf := &state.ReleaseFile{Release: state.Release{
		Project: "p", Checkout: checkout, Rounds: []state.Round{{Name: "a", DesignWritten: true}},
		Passes:    passes,
		Questions: []state.Question{{ID: "q", Options: []string{"yes", "no"}, Recommended: "yes"}},
	}}
	return &state.State{Releases: []*state.ReleaseFile{rf}, Usage: &state.Usage{Format: 1, Percent: percent, ReadAt: t0}}
}

func TestPlanStartsFreePasses(t *testing.T) {
	notStarted := func(name string) state.Pass { return state.Pass{Name: name, State: state.NotStarted} }
	building := func(name string) state.Pass {
		return state.Pass{Name: name, State: state.Building, Started: ago(time.Hour), Session: &state.Session{Name: "p-" + name}}
	}
	waiting := func(name string) state.Pass {
		return state.Pass{Name: name, State: state.Waiting, Wait: &state.Wait{Since: t0, Limits: []string{state.LimitTwoPass}}}
	}
	onGuess := notStarted("g")
	onGuess.Questions = []string{"q"}
	closed := state.Pass{Name: "x", State: state.Done, Closed: &state.Closed{}}
	afterClosed := notStarted("y")
	afterClosed.After = []string{"x"}

	tests := []struct {
		name     string
		checkout string
		percent  int
		passes   []state.Pass
		started  []string
		decided  []string
	}{
		{"a free pass starts then launches", "/c", 10, []state.Pass{notStarted("f")},
			[]string{"f"}, []string{"pass:f:start", "pass:f:launch"}},
		{"a waiting pass takes the slot first", "/c", 10, []state.Pass{building("b"), waiting("w"), notStarted("f")},
			nil, []string{"pass:w:launch"}},
		{"one building, two free, one start", "/c", 10, []state.Pass{building("b"), notStarted("f1"), notStarted("f2")},
			[]string{"f1"}, []string{"pass:f1:start", "pass:f1:launch"}},
		{"two building, none", "/c", 10, []state.Pass{building("b1"), building("b2"), notStarted("f")},
			nil, nil},
		{"under the usage hold", "/c", 100, []state.Pass{notStarted("f")}, nil, nil},
		{"no checkout", "", 10, []state.Pass{notStarted("f")}, nil, nil},
		{"free only on a guess", "/c", 10, []state.Pass{onGuess}, nil, nil},
		{"after a closed pass", "/c", 10, []state.Pass{closed, afterClosed}, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := startWorld(tc.checkout, tc.percent, tc.passes...)
			results, err := Plan(s, Evidence{}, t0)
			if err != nil {
				t.Fatal(err)
			}
			var decided, started []string
			for _, r := range results {
				decided = append(decided, r.Record+":"+r.Decision.Decision)
				if r.Decision.Decision == Start {
					started = append(started, strings.TrimPrefix(r.Record, "pass:"))
					if r.Reason != "free" {
						t.Errorf("%s: reason %q, want free", r.Record, r.Reason)
					}
				}
			}
			if !slices.Equal(started, tc.started) {
				t.Errorf("started %v, want %v (decisions %v)", started, tc.started, decided)
			}
			for _, want := range tc.decided {
				if !slices.Contains(decided, want) {
					t.Errorf("decisions %v lack %s", decided, want)
				}
			}
		})
	}
}
