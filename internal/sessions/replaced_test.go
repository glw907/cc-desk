package sessions

import (
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

func TestPlanReplacedRow(t *testing.T) {
	world := func(percent int, mark bool) *state.State {
		sess := &state.Session{
			Name: "p-b", Kind: state.KindPass, ShortID: "bb22", SessionID: "s-bb22",
			Worktree: "/w/b", Branch: "b", Launched: *ago(3600e9), LaunchTries: 1,
		}
		p := state.Pass{Name: "b", State: state.Building, Started: ago(3600e9), Session: sess}
		if mark {
			p.Replaced = &state.Replaced{Ruling: "r-0002", Time: t0}
		}
		rf := &state.ReleaseFile{Release: state.Release{
			Project: "p", Checkout: "/c", Rounds: []state.Round{{Name: "a"}},
			Passes: []state.Pass{p},
		}}
		rf.Passes[0].Round = "a"
		return &state.State{Releases: []*state.ReleaseFile{rf}, Usage: &state.Usage{Format: 1, Percent: percent, ReadAt: t0}}
	}
	live := Evidence{Records: map[Key]RecordEvidence{
		{"p", "pass:b"}: {Row: Row{ID: "bb22", State: "working", Status: "busy"}, HasRow: true},
	}}
	done := Evidence{Records: map[Key]RecordEvidence{
		{"p", "pass:b"}: {Row: Row{ID: "bb22", State: "done", Status: "idle"}, HasRow: true},
	}}
	tests := []struct {
		name    string
		ev      Evidence
		percent int
		mark    bool
		want    string
		counted bool
	}{
		{"a marked live pass is ended and relaunched uncounted", live, 10, true, EndAndRelaunch, false},
		{"a marked pass whose session finished is relaunched uncounted", done, 10, true, EndAndRelaunch, false},
		{"under the usage hold the relaunch waits", live, 100, true, WaitUsage, false},
		{"no mark leaves the pass alone", live, 10, false, "none", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := Plan(world(tc.percent, tc.mark), tc.ev, t0)
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			for _, r := range results {
				if r.Record != "pass:b" {
					continue
				}
				got = r.Decision.Decision
				if got == EndAndRelaunch {
					if r.Reason != ReasonReplaced || r.Counted == nil || *r.Counted {
						t.Errorf("reason %q counted %v, want replaced and uncounted", r.Reason, r.Counted)
					}
				}
			}
			if got != tc.want {
				t.Errorf("decision %q, want %q", got, tc.want)
			}
		})
	}
}
