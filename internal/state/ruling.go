package state

import (
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode/utf8"
)

// The three guard refusals. Each names the test that failed. They share
// the prefix every ruling refusal starts with.
const (
	rulingPrefix = "cannot record the ruling: "

	RefuseNotDesk      = rulingPrefix + "this session isn't a registered desk"
	RefuseNotUnderDesk = rulingPrefix + "this command isn't running under the desk's process"
	RefuseNoTerminal   = rulingPrefix + "the desk's process has no terminal"
)

// RulingRequest is what a desk session types: Geoff's words, and the
// question they answer. Project, Round, Question, and Option are set
// together or all empty.
type RulingRequest struct {
	Words, Project, Round, Question, Option string
}

// SameStart reports whether two process start times are the same moment.
// The boot time they come from is whole seconds, so two reads of it can
// differ by one; a gap under two seconds matches.
func SameStart(a, b time.Time) bool {
	d := a.Sub(b)
	return d > -2*time.Second && d < 2*time.Second
}

// guard runs the three tests in order and returns the first refusal. Test 3
// reads the terminal of the desk's process, never the caller's own: a desk
// session's commands have none.
func (s *State) guard(c Caller) error {
	i := slices.IndexFunc(s.Desks.Desks, func(d Desk) bool { return c.SessionID != "" && d.SessionID == c.SessionID })
	if i < 0 {
		return Refuse(RefuseNotDesk)
	}
	desk := s.Desks.Desks[i]
	j := slices.IndexFunc(c.Chain, func(p Proc) bool { return p.PID == desk.PID && SameStart(p.Started, desk.Started) })
	if j < 0 {
		return Refuse(RefuseNotUnderDesk)
	}
	if c.Chain[j].TTY == 0 {
		return Refuse(RefuseNoTerminal)
	}
	return nil
}

// nextRulingID is r- and a four-digit number, one more than the count of
// lines already in the log. An id already taken moves to the next number.
func (s *State) nextRulingID() string {
	for n := len(s.Rulings) + 1; ; n++ {
		id := fmt.Sprintf("r-%04d", n)
		if !slices.ContainsFunc(s.Rulings, func(r Ruling) bool { return r.ID == id }) {
			return id
		}
	}
}

// RecordRuling appends one ruling to rulings.jsonl. It checks, in order and
// with the first failure winning: the guard, the words, the four question
// fields, and whether the question is already answered. A ruling that names
// a question answers it, which closes any guess on it and frees what waited.
func RecordRuling(c Caller, r RulingRequest) Event {
	return func(s *State, now time.Time) error {
		if err := s.guard(c); err != nil {
			return err
		}
		if r.Words == "" || !utf8.ValidString(r.Words) {
			return errors.New("the ruling's words must be non-empty UTF-8")
		}
		named := 0
		for _, f := range []string{r.Project, r.Round, r.Question, r.Option} {
			if f != "" {
				named++
			}
		}
		if named != 0 && named != 4 {
			return errors.New("a ruling that answers a question needs project, round, question, and option together")
		}

		var q *Question
		var rf *ReleaseFile
		if named == 4 {
			var err error
			if rf, err = s.liveRelease(r.Project); err != nil {
				return err
			}
			if !slices.ContainsFunc(rf.Rounds, func(rd Round) bool { return rd.Name == r.Round }) {
				return fmt.Errorf("round %q not found in project %q", r.Round, r.Project)
			}
			if q = rf.question(r.Question); q == nil {
				return fmt.Errorf("question %q not found in project %q", r.Question, r.Project)
			}
			if !slices.Contains(q.Options, r.Option) {
				return fmt.Errorf("option %q is not one of question %q's options %q", r.Option, r.Question, q.Options)
			}
			if q.Answer != nil {
				return Refuse(fmt.Sprintf("%squestion %q is already answered", rulingPrefix, r.Question))
			}
		}

		ruling := Ruling{
			Format: Format, ID: s.nextRulingID(), Time: now.UTC(), Desk: c.SessionID,
			Words: r.Words, Project: r.Project, Round: r.Round, Question: r.Question, Option: r.Option,
		}
		s.Rulings = append(s.Rulings, ruling)
		s.newRuling = &ruling
		if q != nil {
			q.Answer = &Answer{Ruling: ruling.ID}
			rf.Dirty = true
		}
		return nil
	}
}

// RegisterDesk adds a desk session, or replaces the entry with the same
// session id. The Desk round calls it when a desk opens; no verb does.
func RegisterDesk(d Desk) Event {
	return func(s *State, now time.Time) error {
		if d.SessionID == "" || d.PID <= 0 || d.Started.IsZero() {
			return errors.New("a desk needs a session id, a PID above 0, and a start time")
		}
		d.Started = d.Started.UTC()
		if i := slices.IndexFunc(s.Desks.Desks, func(e Desk) bool { return e.SessionID == d.SessionID }); i >= 0 {
			s.Desks.Desks[i] = d
		} else {
			s.Desks.Desks = append(s.Desks.Desks, d)
		}
		s.Desks.Format = Format
		s.desksDirty = true
		return nil
	}
}
