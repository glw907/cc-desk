package state

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

// The three guard refusals. Each names the test that failed. They share
// the prefix every ruling refusal starts with.
const (
	rulingPrefix = "cannot record the ruling: "

	RefuseNotDesk      = rulingPrefix + reasonNotDesk
	RefuseNotUnderDesk = rulingPrefix + reasonNotUnderDesk
	RefuseNoTerminal   = rulingPrefix + reasonNoTerminal
)

// The reasons a guard refusal gives. A verb adds its own prefix.
const (
	reasonNotDesk      = "this session isn't a registered desk"
	reasonNotUnderDesk = "this command isn't running under the desk's process"
	reasonNoTerminal   = "the desk's process has no terminal"
)

// RulingRequest is what a desk session types: Geoff's words, and the
// question they answer. Project, Round, Question, and Option are set
// together or all empty. Replaces names the ruling this one replaces, and
// needs the four.
type RulingRequest struct {
	Words, Project, Round, Question, Option, Replaces string
}

// SameStart reports whether two process start times are the same moment.
// The boot time they come from is whole seconds, so two reads of it can
// differ by one; a gap under two seconds matches.
func SameStart(a, b time.Time) bool {
	d := a.Sub(b)
	return d > -2*time.Second && d < 2*time.Second
}

// Guard runs the three tests in order and returns the first refusal, which
// starts with prefix. Test 3 reads the terminal of the desk's process, never
// the caller's own: a desk session's commands have none. A verb calls it
// before any side effect to refuse early, and its event calls it again
// inside the lock.
func (s *State) Guard(c Caller, prefix string) error {
	i := slices.IndexFunc(s.Desks.Desks, func(d Desk) bool { return c.SessionID != "" && d.SessionID == c.SessionID })
	if i < 0 {
		return Refuse(prefix + reasonNotDesk)
	}
	desk := s.Desks.Desks[i]
	j := slices.IndexFunc(c.Chain, func(p Proc) bool { return p.PID == desk.PID && SameStart(p.Started, desk.Started) })
	if j < 0 {
		return Refuse(prefix + reasonNotUnderDesk)
	}
	if c.Chain[j].TTY == 0 {
		return Refuse(prefix + reasonNoTerminal)
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
		if err := s.Guard(c, rulingPrefix); err != nil {
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
		if r.Replaces != "" && named != 4 {
			return errors.New("a ruling that replaces another needs project, round, question, and option")
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
			if r.Replaces != "" {
				if err := s.checkReplace(rf, q, r); err != nil {
					return err
				}
			} else if q.Answer != nil {
				return Refuse(fmt.Sprintf("%squestion %q is already answered", rulingPrefix, r.Question))
			}
		}

		ruling := Ruling{
			Format: Format, ID: s.nextRulingID(), Time: now.UTC(), Desk: c.SessionID,
			Words: r.Words, Project: r.Project, Round: r.Round, Question: r.Question, Option: r.Option,
			Replaces: r.Replaces,
		}
		s.Rulings = append(s.Rulings, ruling)
		s.newRuling = &ruling
		if q != nil {
			q.Answer = &Answer{Ruling: ruling.ID}
			if r.Replaces != "" {
				markReplaced(rf, q, ruling)
			}
			rf.Dirty = true
		}
		return nil
	}
}

// checkReplace refuses a replacement the log or the passes don't allow, in
// this order: the ruling doesn't exist, another already replaced it, it
// answered a different question, it isn't the question's current answer, a
// pass resting on the question has finished, or a merge question's merge was
// called. A pass that has not started, is waiting, is stopped, or is building
// takes the replacement.
func (s *State) checkReplace(rf *ReleaseFile, q *Question, r RulingRequest) error {
	i := slices.IndexFunc(s.Rulings, func(g Ruling) bool { return g.ID == r.Replaces })
	if i < 0 {
		return Refuse(fmt.Sprintf("%sruling %q does not exist", rulingPrefix, r.Replaces))
	}
	if j := slices.IndexFunc(s.Rulings, func(g Ruling) bool { return g.Replaces == r.Replaces }); j >= 0 {
		return Refuse(fmt.Sprintf("%sruling %q is already replaced by %s", rulingPrefix, r.Replaces, s.Rulings[j].ID))
	}
	old := s.Rulings[i]
	if old.Project != r.Project || old.Round != r.Round || old.Question != r.Question {
		return Refuse(fmt.Sprintf("%sruling %q does not answer question %q of %s/%s", rulingPrefix, r.Replaces, r.Question, r.Project, r.Round))
	}
	if q.Answer == nil || q.Answer.Ruling != r.Replaces {
		return Refuse(fmt.Sprintf("%sruling %q is not question %q's current answer", rulingPrefix, r.Replaces, r.Question))
	}
	merge := strings.HasPrefix(q.ID, "merge-")
	for k := range rf.Passes {
		p := &rf.Passes[k]
		if !slices.Contains(p.Questions, q.ID) {
			continue
		}
		switch {
		case p.State == Done:
			return Refuse(fmt.Sprintf("%spass %q, which rests on question %q, has finished; a new change is needed", rulingPrefix, p.Name, q.ID))
		case merge && p.PullRequest != nil && p.PullRequest.MergeCalled != nil:
			return Refuse(fmt.Sprintf("%spass %q, which rests on question %q, has had its merge called; a new change is needed", rulingPrefix, p.Name, q.ID))
		}
	}
	return nil
}

// markReplaced marks each building pass, or stopped pass with a session
// record, that rests on the question whose answer a ruling replaced. The
// check reads the mark and relaunches the session with the new words. A merge
// question takes no mark: the merge rows read its latest answer.
func markReplaced(rf *ReleaseFile, q *Question, ruling Ruling) {
	if strings.HasPrefix(q.ID, "merge-") {
		return
	}
	for k := range rf.Passes {
		p := &rf.Passes[k]
		if slices.Contains(p.Questions, q.ID) && (p.State == Building || p.State == Stopped) && p.Session != nil {
			p.Replaced = &Replaced{Ruling: ruling.ID, Time: ruling.Time}
		}
	}
}

// StartConversation records a design conversation for a round of the live
// release, and refuses a second for the same round or a round whose design
// is written. DesignRound ends it.
func StartConversation(project, round, desk string) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.liveRelease(project)
		if err != nil {
			return err
		}
		i := slices.IndexFunc(rf.Rounds, func(r Round) bool { return r.Name == round })
		if i < 0 {
			return fmt.Errorf("round %q not found in project %q", round, project)
		}
		if desk == "" {
			return errors.New("a conversation needs the desk's session id")
		}
		if rf.Rounds[i].DesignWritten {
			return Refuse(fmt.Sprintf("round %q of project %q already has its design written", round, project))
		}
		if slices.ContainsFunc(s.Desks.Conversations, func(c Conversation) bool { return c.Project == project && c.Round == round }) {
			return Refuse(fmt.Sprintf("round %q of project %q already has a conversation", round, project))
		}
		s.Desks.Conversations = append(s.Desks.Conversations, Conversation{Project: project, Round: round, Desk: desk, Started: now.UTC()})
		s.Desks.Format = Format
		s.desksDirty = true
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
