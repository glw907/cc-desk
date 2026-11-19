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

// reportWorld is a release of rounds r1 and r2 with a checkout, each round's
// passes in the given state, and the reports recorded.
func reportWorld(r1, r2 []state.PassState, reports ...state.Report) *state.State {
	rf := &state.ReleaseFile{Release: state.Release{
		Project: "p", Checkout: "/c", Rounds: []state.Round{{Name: "r1"}, {Name: "r2"}},
	}}
	add := func(round string, states []state.PassState) {
		for i, st := range states {
			rf.Passes = append(rf.Passes, state.Pass{Name: round + string(rune('a'+i)), Round: round, State: st})
		}
	}
	add("r1", r1)
	add("r2", r2)
	return &state.State{Releases: []*state.ReleaseFile{rf}, Reports: state.Reports{Reports: reports}, Usage: &state.Usage{Format: 1, Percent: 10, ReadAt: t0}}
}

func TestPlanLaunchesTheConductorForAReport(t *testing.T) {
	done, building := state.Done, state.Building
	round := func(r string) state.Report { return state.Report{Project: "p", Round: r} }
	tests := []struct {
		name     string
		r1, r2   []state.PassState
		reports  []state.Report
		conduct  *state.Session
		want     string
		wantDue  string
		wantStop string
	}{
		{name: "a round with a pass not done", r1: []state.PassState{done, building}},
		{name: "every pass done and no report", r1: []state.PassState{done, done}, r2: []state.PassState{building},
			want: Launch, wantDue: "r1"},
		{name: "every pass done and a report", r1: []state.PassState{done, done}, r2: []state.PassState{building},
			reports: []state.Report{round("r1")}},
		{name: "a round with no pass holds the release report back", r1: []state.PassState{done}, reports: []state.Report{round("r1")}},
		{name: "the release report follows every round's", r1: []state.PassState{done}, r2: []state.PassState{done},
			reports: []state.Report{round("r1"), round("r2")}, want: Launch, wantDue: ReleaseReport},
		{name: "a release holding a pass not done", r1: []state.PassState{done}, r2: []state.PassState{building},
			reports: []state.Report{round("r1")}},
		{name: "the release report recorded", r1: []state.PassState{done}, r2: []state.PassState{done},
			reports: []state.Report{round("r1"), round("r2"), round("")}},
		{name: "a live conductor", r1: []state.PassState{done}, r2: []state.PassState{building},
			conduct: &state.Session{Name: "p-conductor", Kind: state.KindConductor, ForReport: "r1"}},
		{name: "an ended conductor for another report", r1: []state.PassState{done}, r2: []state.PassState{building},
			conduct: &state.Session{Name: "p-conductor", Kind: state.KindConductor, ForReport: "r0", Ended: ago(0)},
			want:    Launch, wantDue: "r1"},
		{name: "an ended conductor with two counted relaunches", r1: []state.PassState{done}, r2: []state.PassState{building},
			conduct: &state.Session{Name: "p-conductor", Kind: state.KindConductor, ForReport: "r1", Ended: ago(0),
				Relaunches: []state.Relaunch{{Cause: "no_report", Counted: true}, {Cause: "no_report", Counted: true}}},
			want: Stop, wantDue: "r1", wantStop: string(state.StopRelaunchLimit)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := reportWorld(tc.r1, tc.r2, tc.reports...)
			s.Releases[0].Conductor = tc.conduct
			if tc.name == "a live conductor" {
				tc.conduct.Launched = t0
			}
			results, err := Plan(s, Evidence{}, t0)
			if err != nil {
				t.Fatal(err)
			}
			var got []Result
			for _, r := range results {
				if r.Record == "conductor" && (r.Decision.Decision == Launch || r.Decision.Decision == WaitUsage || r.Reason == string(state.StopRelaunchLimit)) {
					got = append(got, r)
				}
			}
			if tc.want == "" {
				if len(got) != 0 {
					t.Fatalf("conductor decisions %+v, want none", got)
				}
				return
			}
			if len(got) != 1 || got[0].Decision.Decision != tc.want {
				t.Fatalf("conductor decisions %+v, want one %q", got, tc.want)
			}
			if tc.want == Launch && got[0].Reason != ReasonReportDue {
				t.Errorf("reason %q, want %q", got[0].Reason, ReasonReportDue)
			}
			if tc.wantStop != "" && got[0].Reason != tc.wantStop {
				t.Errorf("reason %q, want %q", got[0].Reason, tc.wantStop)
			}
			if due, _ := ReportDue(s, s.Releases[0]); due != tc.wantDue {
				t.Errorf("report due %q, want %q", due, tc.wantDue)
			}
		})
	}
}

func TestPlanHoldsTheConductorStartUnderTheUsageHold(t *testing.T) {
	s := reportWorld([]state.PassState{state.Done}, nil)
	s.Usage = &state.Usage{Format: 1, Percent: 100, ReadAt: t0}
	results, err := Plan(s, Evidence{}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Record != "conductor" || results[0].Decision.Decision != WaitUsage {
		t.Errorf("results %+v, want one wait_usage for the conductor", results)
	}
}

func TestPlanLeavesAReleaseWithNoCheckoutAlone(t *testing.T) {
	s := reportWorld([]state.PassState{state.Done}, nil)
	s.Releases[0].Checkout = ""
	results, err := Plan(s, Evidence{}, t0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("results %+v, want none", results)
	}
}
