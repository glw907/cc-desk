package state

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// QuestionFile is the input of the question add verb: one whole question and
// the not-started passes of a round that rest on it. A working session
// writes the whole question, so Background, Causes, and Reason are required
// here, though a release file may hold a question without them.
type QuestionFile struct {
	ID          string            `json:"id"`
	Text        string            `json:"text"`
	Options     []string          `json:"options"`
	Recommended string            `json:"recommended"`
	Background  []string          `json:"background"`
	Causes      map[string]string `json:"causes"`
	Reason      string            `json:"reason"`
	NeverGuess  bool              `json:"never_guess"`
	Passes      []string          `json:"passes"`
}

// DecodeQuestion strictly decodes a question file and names the first part
// that is missing or wrong. A key set to null reads as missing.
func DecodeQuestion(data []byte) (*QuestionFile, error) {
	var q QuestionFile
	if err := decodeStrict(data, &q); err != nil {
		return nil, err
	}
	for _, f := range []struct {
		name  string
		empty bool
	}{
		{"id", q.ID == ""},
		{"text", q.Text == ""},
		{"options", len(q.Options) == 0},
		{"recommended", q.Recommended == ""},
		{"background", len(q.Background) == 0},
		{"causes", len(q.Causes) == 0},
		{"reason", q.Reason == ""},
		{"passes", len(q.Passes) == 0},
	} {
		if f.empty {
			return nil, fmt.Errorf("missing required field %q", f.name)
		}
	}
	for i, b := range q.Background {
		if strings.TrimSpace(b) == "" {
			return nil, fmt.Errorf("background paragraph %d is empty", i+1)
		}
	}
	return &q, nil
}

// AddQuestion adds the question to the round's release and to each named
// pass, which must be a not-started pass of the round. A pass that is
// waiting or further along is refused: the question would have to move it
// back.
func AddQuestion(project, round string, qf *QuestionFile) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.liveRelease(project)
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(rf.Rounds, func(r Round) bool { return r.Name == round }) {
			return fmt.Errorf("round %q not found in project %q", round, project)
		}
		if rf.question(qf.ID) != nil {
			return Refuse(fmt.Sprintf("question %q already exists in project %q", qf.ID, project))
		}
		next := rf.Release
		next.Passes = slices.Clone(rf.Passes)
		for _, name := range qf.Passes {
			i := slices.IndexFunc(next.Passes, func(p Pass) bool { return p.Name == name })
			switch {
			case i < 0:
				return fmt.Errorf("pass %q not found in project %q", name, project)
			case next.Passes[i].Round != round:
				return fmt.Errorf("pass %q is in round %q, not %q", name, next.Passes[i].Round, round)
			case next.Passes[i].State != NotStarted:
				return Refuse(fmt.Sprintf("pass %q is %s; a question can be added only to a pass that has not started", name, next.Passes[i].State))
			case slices.Contains(next.Passes[i].Questions, qf.ID):
				return fmt.Errorf("pass %q is named twice", name)
			}
			next.Passes[i].Questions = append(slices.Clone(next.Passes[i].Questions), qf.ID)
		}
		next.Questions = append(slices.Clone(rf.Questions), Question{
			ID: qf.ID, Text: qf.Text, Options: qf.Options, Recommended: qf.Recommended,
			Background: qf.Background, Causes: qf.Causes, Reason: qf.Reason, NeverGuess: qf.NeverGuess,
		})
		if err := validateRelease(&next); err != nil {
			return fmt.Errorf("question %q: %w", qf.ID, err)
		}
		rf.Release = next
		rf.Dirty = true
		return nil
	}
}
