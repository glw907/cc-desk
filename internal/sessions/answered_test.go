package sessions

import (
	"testing"

	"github.com/glw907/cc-desk/internal/state"
)

func TestDecideAnsweredMerge(t *testing.T) {
	closed := pull("abc", ok("gate"))
	closed.State = "CLOSED"
	outside := func(p *state.PullRequest) { p.InsideDesign = false }
	tests := []struct {
		name         string
		option       string
		mod          func(*state.PullRequest)
		read         PullRead
		wantDecision string
		wantReason   string
		wantCommit   string
		wantClosed   bool
	}{
		{"merge at the asked head", OptionMerge, outside, PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Merge, "", "abc", false},
		{"merge pins the asked commit, not the report", OptionMerge, func(p *state.PullRequest) { p.InsideDesign, p.AskedCommit = false, "def" },
			PullRead{Pulls: []Pull{pull("def", ok("gate"))}}, Merge, "", "def", false},
		{"merge after a merge_failed question merges at the asked head", OptionMerge, func(p *state.PullRequest) { p.InsideDesign, p.MergeFailures = true, 0 },
			PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Merge, "", "abc", false},
		{"merge with the head moved asks again", OptionMerge, outside, PullRead{Pulls: []Pull{pull("def", ok("gate"))}}, Ask, ReasonHeadMoved, "def", false},
		{"merge with a failed check asks again", OptionMerge, outside, PullRead{Pulls: []Pull{pull("abc", failed("gate"))}}, Ask, ReasonChecksFailed, "abc", false},
		{"merge with a check running holds", OptionMerge, outside, PullRead{Pulls: []Pull{pull("abc", Check{Name: "gate", Bucket: Pending})}}, None, "", "", false},
		{"merge on a closed pull request holds", OptionMerge, outside, PullRead{Pulls: []Pull{closed}}, None, "", "", false},
		{"fix again beats the fix limit", OptionFixAgain, func(p *state.PullRequest) { p.FixAttempts = 2 }, PullRead{Pulls: []Pull{pull("abc", failed("gate"))}}, Fix, ReasonFix, "abc", false},
		{"fix again with no failure", OptionFixAgain, outside, PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Fix, ReasonFix, "abc", false},
		{"close", OptionClose, outside, PullRead{Pulls: []Pull{pull("abc", ok("gate"))}}, Close, "", "abc", false},
		{"close runs again on a closed pull request", OptionClose, outside, PullRead{Pulls: []Pull{closed}}, Close, "", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, key := world("pass", 10, ready(func(p *state.PullRequest) {
				p.Question = "merge-x-1"
				if tt.mod != nil {
					tt.mod(p)
				}
			}))
			s.Releases[0].Questions = append(s.Releases[0].Questions, state.Question{ID: "merge-x-1", Answer: &state.Answer{Ruling: "r-0001"}})
			s.Rulings = append(s.Rulings, state.Ruling{ID: "r-0001", Option: tt.option, Words: "His words."})
			ev := Evidence{
				Records: map[Key]RecordEvidence{key: {Row: rowWorking, HasRow: true}},
				Pulls:   map[Key]PullRead{key: tt.read},
			}
			got := Decide(s, ev, t0)
			if len(got) != 1 {
				t.Fatalf("decisions = %d, want 1", len(got))
			}
			r := got[0]
			if r.Decision.Decision != tt.wantDecision || r.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want %s/%s", r.Decision.Decision, r.Reason, tt.wantDecision, tt.wantReason)
			}
			if tt.wantCommit != "" && r.Merge.Commit != tt.wantCommit {
				t.Errorf("commit = %q, want %q", r.Merge.Commit, tt.wantCommit)
			}
			if r.Merge.AlreadyClosed != tt.wantClosed {
				t.Errorf("already closed = %v, want %v", r.Merge.AlreadyClosed, tt.wantClosed)
			}
			if tt.wantDecision != None && (r.Merge.Words != "His words." || r.Merge.Ruling != "r-0001") {
				t.Errorf("words %q ruling %q, want the answer's", r.Merge.Words, r.Merge.Ruling)
			}
		})
	}

	t.Run("an open question still holds", func(t *testing.T) {
		s, key := world("pass", 10, ready(func(p *state.PullRequest) { p.Question = "merge-x-1" }))
		ev := Evidence{
			Records: map[Key]RecordEvidence{key: {Row: rowWorking, HasRow: true}},
			Pulls:   map[Key]PullRead{key: {Pulls: []Pull{pull("abc", ok("gate"))}}},
		}
		if got := Decide(s, ev, t0); got[0].Decision.Decision != None {
			t.Errorf("decision = %s, want none", got[0].Decision.Decision)
		}
	})
}
