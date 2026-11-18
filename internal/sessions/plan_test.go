package sessions

import (
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

// planWorld builds one release of four passes: a and b stopped on usage with
// a session each, c and d waiting. Whether it has a checkout and how full the
// weekly usage is vary by case.
func planWorld(checkout string, percent int) *state.State {
	stopped := func(name, id string) state.Pass {
		sess := &state.Session{
			Name: "p-" + name, Kind: state.KindPass, ShortID: id, SessionID: "s-" + id,
			Worktree: "/w/" + name, Branch: name, Launched: *ago(2 * 3600e9), LaunchTries: 1,
			LastSeen: ago(30 * 60e9), LastState: state.SeenBlocked,
		}
		return state.Pass{
			Name: name, Round: "a", State: state.Stopped, StopReason: state.StopUsage, Session: sess,
			Started: ago(2 * 3600e9), Stops: []state.StopInterval{{Start: *ago(30 * 60e9)}},
		}
	}
	waiting := func(name string) state.Pass {
		return state.Pass{Name: name, Round: "a", State: state.Waiting, Wait: &state.Wait{Since: t0, Limits: []string{state.LimitTwoPass}}}
	}
	rf := &state.ReleaseFile{Release: state.Release{
		Project: "p", Checkout: checkout, Rounds: []state.Round{{Name: "a"}},
		Passes: []state.Pass{stopped("a", "aa11"), stopped("b", "bb22"), waiting("c"), waiting("d")},
	}}
	return &state.State{Releases: []*state.ReleaseFile{rf}, Usage: &state.Usage{Format: 1, Percent: percent, ReadAt: t0}}
}

func TestPlan(t *testing.T) {
	tests := []struct {
		name     string
		checkout string
		percent  int
		building []string
		decided  map[string]string
	}{
		{"hold clear with a checkout", "/c", 10, []string{"a", "b"},
			map[string]string{"pass:a": Relaunch, "pass:b": Relaunch}},
		{"hold clear with no checkout", "", 10, []string{"c", "d"},
			map[string]string{"pass:a": Relaunch, "pass:b": Relaunch}},
		{"under the hold", "/c", 100, nil,
			map[string]string{"pass:a": WaitUsage, "pass:b": WaitUsage}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := planWorld(tc.checkout, tc.percent)
			results, err := Plan(s, Evidence{}, t0)
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range results {
				if want, ok := tc.decided[r.Record]; ok && r.Decision.Decision != want {
					t.Errorf("%s: decision %q, want %q", r.Record, r.Decision.Decision, want)
				}
			}
			var building []string
			for _, p := range s.Releases[0].Passes {
				if p.State == state.Building {
					building = append(building, p.Name)
				}
			}
			if len(building) != len(tc.building) {
				t.Fatalf("building %v, want %v", building, tc.building)
			}
			for i := range building {
				if building[i] != tc.building[i] {
					t.Errorf("building %v, want %v", building, tc.building)
				}
			}
		})
	}
}

// A pass that settle starts on its own and has no session gets a launch
// decision, or wait_usage under the hold.
func TestPlanLaunchesABuildingPassWithNoSession(t *testing.T) {
	for _, tc := range []struct {
		percent int
		want    string
	}{{10, Launch}, {100, WaitUsage}} {
		s := planWorld("/c", tc.percent)
		rf := s.Releases[0]
		rf.Passes = []state.Pass{{Name: "e", Round: "a", State: state.Building, Started: ago(0)}}
		results, err := Plan(s, Evidence{}, t0)
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != 1 || results[0].Record != "pass:e" || results[0].Decision.Decision != tc.want {
			t.Errorf("percent %d: results %+v, want one %q for pass:e", tc.percent, results, tc.want)
		}
	}
}
