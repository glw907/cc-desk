package state

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// replaceState is deskState with q1 answered by r-0001, q2 by r-0002, and a
// not-started pass p1 resting on q1.
func replaceState() (*State, *ReleaseFile) {
	s, rf := deskState()
	s.Rulings = []Ruling{
		{Format: Format, ID: "r-0001", Time: at(8, 30), Desk: "d", Words: "Yes.", Project: "alpha", Round: "r1", Question: "q1", Option: "a"},
		{Format: Format, ID: "r-0002", Time: at(8, 31), Desk: "d", Words: "Other.", Project: "alpha", Round: "r1", Question: "q2", Option: "a"},
	}
	rf.Questions[0].Answer = &Answer{Ruling: "r-0001"}
	rf.Questions[1].Answer = &Answer{Ruling: "r-0002"}
	addPass(rf, "p1", "r1", 30, NotStarted).Questions = []string{"q1"}
	return s, rf
}

func TestReplaceRuling(t *testing.T) {
	replace := RulingRequest{Words: "No, the other.", Project: "alpha", Round: "r1", Question: "q1", Option: "b", Replaces: "r-0001"}
	setState := func(st PassState) func(s *State, rf *ReleaseFile) {
		return func(s *State, rf *ReleaseFile) { rf.Passes[0].State = st }
	}
	tests := []struct {
		name    string
		req     func(r RulingRequest) RulingRequest
		setup   func(s *State, rf *ReleaseFile)
		refused string // a substring of the refusal
		bad     bool   // a bad request, exit 1
	}{
		{name: "a good replacement"},
		{name: "no ruling by that id", refused: `ruling "r-0009" does not exist`, req: func(r RulingRequest) RulingRequest { r.Replaces = "r-0009"; return r }},
		{
			name: "already replaced", refused: `ruling "r-0001" is already replaced by r-0003`,
			setup: func(s *State, rf *ReleaseFile) {
				s.Rulings = append(s.Rulings, Ruling{ID: "r-0003", Replaces: "r-0001"})
			},
		},
		{name: "a different question", refused: `does not answer question "q1"`, req: func(r RulingRequest) RulingRequest { r.Replaces = "r-0002"; return r }},
		{
			name: "a different round", refused: "does not answer question",
			setup: func(s *State, rf *ReleaseFile) { s.Rulings[0].Round = "r2" },
		},
		{
			name: "a different project", refused: "does not answer question",
			setup: func(s *State, rf *ReleaseFile) { s.Rulings[0].Project = "zed" },
		},
		{
			name: "not the current answer", refused: `is not question "q1"'s current answer`,
			setup: func(s *State, rf *ReleaseFile) { rf.Questions[0].Answer = &Answer{Ruling: "r-0002"} },
		},
		{name: "a merged pass", refused: "has merged; a new change is needed", setup: setState(Done)},
		{name: "a building pass", refused: "has started; replacing it while it builds comes in pass 3", setup: setState(Building)},
		{name: "a stopped pass", refused: "has started; replacing it while it builds comes in pass 3", setup: setState(Stopped)},
		{name: "a waiting pass", refused: "has started; replacing it while it builds comes in pass 3", setup: setState(Waiting)},
		{
			name: "merged wins over started", refused: "has merged",
			setup: func(s *State, rf *ReleaseFile) {
				addPass(rf, "p0", "r1", 30, Building).Questions = []string{"q1"}
				rf.Passes[0].State = Done
			},
		},
		{name: "no question fields", bad: true, req: func(r RulingRequest) RulingRequest {
			return RulingRequest{Words: r.Words, Replaces: r.Replaces}
		}},
		{name: "an unknown option", bad: true, req: func(r RulingRequest) RulingRequest { r.Option = "zed"; return r }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, rf := replaceState()
			if tt.setup != nil {
				tt.setup(s, rf)
			}
			req := replace
			if tt.req != nil {
				req = tt.req(req)
			}
			lines, answer := len(s.Rulings), rf.Questions[0].Answer
			err := RecordRuling(deskCaller(), req)(s, at(9, 0))
			if tt.refused != "" || tt.bad {
				if tt.refused != "" {
					if !isRefusal(err) || !strings.Contains(err.Error(), tt.refused) || !strings.HasPrefix(err.Error(), rulingPrefix) {
						t.Fatalf("error = %v, want a refusal with %q", err, tt.refused)
					}
				} else if err == nil || isRefusal(err) {
					t.Fatalf("error = %v, want a bad request", err)
				}
				if len(s.Rulings) != lines || s.newRuling != nil || rf.Questions[0].Answer != answer || rf.Dirty {
					t.Error("a refused replacement changed the state")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := s.newRuling
			if got == nil || got.ID != "r-0003" || got.Replaces != "r-0001" || got.Option != "b" {
				t.Fatalf("ruling = %+v", got)
			}
			if len(s.Rulings) != lines+1 || s.Rulings[0].ID != "r-0001" {
				t.Errorf("rulings = %+v, want both lines kept", s.Rulings)
			}
			if a := rf.Questions[0].Answer; a == nil || a.Ruling != "r-0003" || !rf.Dirty {
				t.Errorf("answer = %+v, dirty %v", a, rf.Dirty)
			}
		})
	}

	t.Run("an answered question without replaces is still refused", func(t *testing.T) {
		s, _ := replaceState()
		req := replace
		req.Replaces = ""
		if err := RecordRuling(deskCaller(), req)(s, at(9, 0)); !isRefusal(err) || !strings.Contains(err.Error(), "already answered") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("a carried answer is no ruling to replace", func(t *testing.T) {
		s, rf := replaceState()
		rf.Questions[0].Answer = &Answer{Carried: &Carried{}}
		if err := RecordRuling(deskCaller(), replace)(s, at(9, 0)); !isRefusal(err) || !strings.Contains(err.Error(), "current answer") {
			t.Errorf("error = %v", err)
		}
	})

	t.Run("a replacement names an earlier ruling when it loads", func(t *testing.T) {
		dir := t.TempDir()
		data := `{"format":1,"id":"r-0001","time":"2026-11-20T09:00:00Z","desk":"d","words":"w","replaces":"r-0002"}` + "\n" +
			`{"format":1,"id":"r-0002","time":"2026-11-20T09:01:00Z","desk":"d","words":"w"}` + "\n"
		if err := os.WriteFile(filepath.Join(dir, "rulings.jsonl"), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Load(dir)
		fe, ok := err.(*FileError)
		if !ok || fe.File != "rulings.jsonl" || fe.Line != 1 || !strings.Contains(err.Error(), "replaces") {
			t.Errorf("error = %v, want a rulings.jsonl line 1 error", err)
		}
	})
}
