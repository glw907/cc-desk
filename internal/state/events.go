package state

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

var projectName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// liveRelease finds the live release of a project.
func (s *State) liveRelease(project string) (*ReleaseFile, error) {
	for _, rf := range s.Releases {
		if !rf.Done && rf.Project == project {
			return rf, nil
		}
	}
	return nil, fmt.Errorf("project %q not found", project)
}

func (rf *ReleaseFile) pass(name string) (*Pass, error) {
	for i := range rf.Passes {
		if rf.Passes[i].Name == name {
			return &rf.Passes[i], nil
		}
	}
	return nil, fmt.Errorf("pass %q not found in project %q", name, rf.Project)
}

// AddRelease installs r as releases/<project>.json. It refuses a project
// that already has a live release. The release must be new: every pass
// not started, and no answer but a carried-over one.
func AddRelease(r *Release) Event {
	return func(s *State, now time.Time) error {
		if !projectName.MatchString(r.Project) {
			return fmt.Errorf("project name %q must be lower case letters, digits, dot, dash, or underscore", r.Project)
		}
		for i := range r.Passes {
			p := &r.Passes[i]
			if p.State != NotStarted || p.Started != nil || p.Finished != nil || len(p.Stops) > 0 ||
				p.StopReason != "" || p.ActualMinutes != nil || p.Wait != nil || p.Guess != nil ||
				p.Session != nil || p.PullRequest != nil {
				return fmt.Errorf("pass %q: a new release's passes must be not_started with no events", p.Name)
			}
		}
		if r.Checkout != "" || r.DefaultBranch != "" || r.Conductor != nil {
			return errors.New("a new release carries no checkout, default_branch, or conductor; release checkout and the session verbs set them")
		}
		for i := range r.Rounds {
			if r.Rounds[i].Session != nil {
				return fmt.Errorf("round %q: a new release's rounds carry no session record", r.Rounds[i].Name)
			}
		}
		for i := range r.Questions {
			if a := r.Questions[i].Answer; a != nil && a.Ruling != "" {
				return fmt.Errorf("question %q: only a carried-over answer can come in with a release", r.Questions[i].ID)
			}
		}
		if _, err := s.liveRelease(r.Project); err == nil {
			return Refuse(fmt.Sprintf("project %q already has a live release", r.Project))
		}
		s.Releases = append(s.Releases, &ReleaseFile{
			Path:    "releases/" + r.Project + ".json",
			Dirty:   true,
			Release: *r,
		})
		return nil
	}
}

// doneStampLayout is the UTC time layout in the name of a done release's file.
const doneStampLayout = "20060102T150405Z"

// ReleaseDone moves the project's live release to
// releases/done/<name>-<UTC stamp>.json and marks it done. Commit writes the
// new file and then removes releases/<project>.json. It refuses a project
// with no live release. When path is not nil it receives the new file's path
// relative to the state folder.
func ReleaseDone(project string, path *string) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.liveRelease(project)
		if err != nil {
			return Refuse(fmt.Sprintf("project %q has no live release", project))
		}
		to := "releases/done/" + rf.Name + "-" + now.UTC().Format(doneStampLayout) + ".json"
		if slices.ContainsFunc(s.Releases, func(o *ReleaseFile) bool { return o.Path == to }) {
			return Refuse(fmt.Sprintf("%s already exists", to))
		}
		rf.moved, rf.Path, rf.Done, rf.Dirty = rf.Path, to, true, true
		if path != nil {
			*path = to
		}
		return nil
	}
}

// Design is a round's passes and questions, as a planning session writes
// them.
type Design struct {
	Format    int        `json:"format"`
	Passes    []Pass     `json:"passes"`
	Questions []Question `json:"questions,omitempty"`
	// Design, Context, and Calls are the round fields of the same names.
	// All three are optional.
	Design  string   `json:"design,omitempty"`
	Context []string `json:"context,omitempty"`
	Calls   []Call   `json:"calls,omitempty"`
}

// DecodeDesign strictly decodes a design input.
func DecodeDesign(data []byte) (*Design, error) {
	var d Design
	if err := decodeStrict(data, &d); err != nil {
		return nil, err
	}
	if err := checkFormat(d.Format); err != nil {
		return nil, err
	}
	if len(d.Passes) == 0 {
		return nil, errors.New(`missing required field "passes"`)
	}
	if err := checkRoundNotes(d.Context, d.Calls); err != nil {
		return nil, err
	}
	return &d, nil
}

// DesignRound adds the design's passes and questions to the round and
// marks its design written. It refuses a round whose design is written.
func DesignRound(project, round string, d *Design) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.liveRelease(project)
		if err != nil {
			return err
		}
		i := slices.IndexFunc(rf.Rounds, func(r Round) bool { return r.Name == round })
		if i < 0 {
			return fmt.Errorf("round %q not found in project %q", round, project)
		}
		if rf.Rounds[i].DesignWritten {
			return Refuse(fmt.Sprintf("round %q of project %q already has its design written", round, project))
		}

		next := rf.Release
		next.Passes = slices.Clone(rf.Passes)
		next.Questions = slices.Clone(rf.Questions)
		next.Rounds = slices.Clone(rf.Rounds)
		for _, p := range d.Passes {
			if p.Round == "" {
				p.Round = round
			}
			if p.Round != round {
				return fmt.Errorf("pass %q is in round %q, not %q", p.Name, p.Round, round)
			}
			if p.State != NotStarted || p.Started != nil || p.Finished != nil || len(p.Stops) > 0 ||
				p.StopReason != "" || p.ActualMinutes != nil || p.Wait != nil || p.Guess != nil ||
				p.Session != nil || p.PullRequest != nil {
				return fmt.Errorf("pass %q: a designed pass must be not_started with no events", p.Name)
			}
			next.Passes = append(next.Passes, p)
		}
		for _, q := range d.Questions {
			if q.Answer != nil {
				return fmt.Errorf("question %q: a design carries no answers", q.ID)
			}
			next.Questions = append(next.Questions, q)
		}
		next.Rounds[i].DesignWritten = true
		next.Rounds[i].Design = d.Design
		next.Rounds[i].Context = slices.Clone(d.Context)
		next.Rounds[i].Calls = slices.Clone(d.Calls)
		if err := validateRelease(&next); err != nil {
			return fmt.Errorf("design for round %q: %w", round, err)
		}
		rf.Release = next
		rf.Dirty = true
		if n := len(s.Desks.Conversations); n > 0 {
			s.Desks.Conversations = slices.DeleteFunc(s.Desks.Conversations, func(c Conversation) bool {
				return c.Project == project && c.Round == round
			})
			if len(s.Desks.Conversations) != n {
				s.desksDirty = true
			}
		}
		return nil
	}
}

func findPass(s *State, project, pass string) (*ReleaseFile, *Pass, error) {
	rf, err := s.liveRelease(project)
	if err != nil {
		return nil, nil, err
	}
	p, err := rf.pass(pass)
	if err != nil {
		return nil, nil, err
	}
	return rf, p, nil
}

// StartPass starts a not-started pass that is free to start, or resumes a
// stopped one and closes its open stop interval. A pass that breaks a rule
// is refused. A free pass that a limit holds is recorded as waiting.
func StartPass(project, pass string) Event { return start(project, pass, "") }

// StartOnGuess is StartPass for a pass resting on an unanswered question:
// it records the guess, taking the question's recommended option, and
// leaves the question out of the rules a plain start would break.
func StartOnGuess(project, pass, question string) Event { return start(project, pass, question) }

func start(project, pass, guess string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		var q *Question
		if guess != "" {
			if q = rf.question(guess); q == nil {
				return fmt.Errorf("question %q not found in project %q", guess, project)
			}
			switch {
			case !slices.Contains(p.Questions, guess):
				return Refuse(fmt.Sprintf("cannot start %s/%s: pass %q does not rest on question %q", project, pass, pass, guess))
			case q.Answer != nil:
				return Refuse(fmt.Sprintf("cannot start %s/%s: question %q is already answered", project, pass, guess))
			case q.NeverGuess:
				return Refuse(fmt.Sprintf("cannot start %s/%s: question %q is never guessed; it waits for a ruling", project, pass, guess))
			}
		}
		t := now.UTC()
		switch p.State {
		case Stopped:
			last := &p.Stops[len(p.Stops)-1]
			if t.Before(last.Start) {
				return fmt.Errorf("the clock %s is before pass %q stopped at %s", t.Format(time.RFC3339), pass, last.Start.Format(time.RFC3339))
			}
			last.End = &t
			p.StopReason, p.StopDetail = "", ""
			p.State = Building
			rf.Dirty = true
			s.passEvent = true
			return nil
		case NotStarted:
		default:
			text := fmt.Sprintf("pass %q is %s and cannot start", pass, p.State)
			if p.Session != nil && p.Session.Ended == nil {
				text += ": session " + p.Session.Name + " is live"
			}
			return Refuse(text)
		}

		if broken := s.Breaks(rf, p, guess); len(broken) > 0 {
			parts := make([]string, len(broken))
			for i, b := range broken {
				parts[i] = b.Name + " (" + b.Detail + ")"
			}
			return Refuse(fmt.Sprintf("cannot start %s/%s: %s", project, pass, strings.Join(parts, "; ")))
		}
		if q != nil {
			p.Guess = &Guess{Question: q.ID, Recommendation: q.Recommended, At: t}
		}
		rf.Dirty = true
		if limits := s.Limits(); len(limits) > 0 {
			p.State, p.Wait = Waiting, &Wait{Since: t, Limits: limits}
			return nil
		}
		p.State, p.Started = Building, &t
		s.passEvent = true
		return nil
	}
}

// FinishPass finishes a building pass and records its actual minutes: the
// time from start to finish less the time spent stopped, rounded to the
// nearest minute.
func FinishPass(project, pass string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		switch p.State {
		case Building:
		case Done:
			return Refuse(fmt.Sprintf("pass %q is already done", pass))
		case Stopped:
			return Refuse(fmt.Sprintf("pass %q is stopped; start it before finishing it", pass))
		default:
			return Refuse(fmt.Sprintf("pass %q has not started", pass))
		}
		t := now.UTC()
		if t.Before(*p.Started) {
			return fmt.Errorf("the clock %s is before pass %q started at %s", t.Format(time.RFC3339), pass, p.Started.Format(time.RFC3339))
		}
		elapsed := t.Sub(*p.Started)
		for _, st := range p.Stops {
			elapsed -= st.End.Sub(st.Start)
		}
		minutes := int(math.Round(elapsed.Minutes()))
		p.Finished = &t
		p.ActualMinutes = &minutes
		p.State = Done
		rf.Dirty = true
		s.passEvent = true
		return nil
	}
}

// ClosePass records a building or stopped pass as closed by a ruling: done,
// with its finish time and actual minutes, any open stop closed, and no merge.
// A pass that has not started, is waiting, or is done has no pull request to
// close and is refused. Only a unit test calls it; the verb is pass 3's.
func ClosePass(project, pass, rulingID string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		switch p.State {
		case Building, Stopped:
		case Done:
			return Refuse(fmt.Sprintf("pass %q is already done", pass))
		default:
			return Refuse(fmt.Sprintf("pass %q is %s and has no pull request to close", pass, p.State))
		}
		if !slices.ContainsFunc(s.Rulings, func(g Ruling) bool { return g.ID == rulingID }) {
			return fmt.Errorf("ruling %q not found", rulingID)
		}
		t := now.UTC()
		if t.Before(*p.Started) {
			return fmt.Errorf("the clock %s is before pass %q started at %s", t.Format(time.RFC3339), pass, p.Started.Format(time.RFC3339))
		}
		if p.State == Stopped {
			last := &p.Stops[len(p.Stops)-1]
			if t.Before(last.Start) {
				return fmt.Errorf("the clock %s is before pass %q stopped at %s", t.Format(time.RFC3339), pass, last.Start.Format(time.RFC3339))
			}
			last.End = &t
			p.StopReason, p.StopDetail = "", ""
		}
		elapsed := t.Sub(*p.Started)
		for _, st := range p.Stops {
			elapsed -= st.End.Sub(st.Start)
		}
		minutes := int(math.Round(elapsed.Minutes()))
		p.Finished, p.ActualMinutes, p.State = &t, &minutes, Done
		p.Closed = &Closed{Ruling: rulingID, Time: t}
		rf.Dirty = true
		s.passEvent = true
		return nil
	}
}

// StopPass stops a building pass for one of the stop reasons and opens a
// stop interval.
func StopPass(project, pass string, reason StopReason) Event {
	return func(s *State, now time.Time) error {
		if !ValidStopReason(reason) {
			return fmt.Errorf("stop reason %q is not one of %v", reason, StopReasons)
		}
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.State != Building {
			return Refuse(fmt.Sprintf("pass %q is %s and cannot stop", pass, p.State))
		}
		if now.Before(*p.Started) {
			return fmt.Errorf("the clock %s is before pass %q started at %s", now.UTC().Format(time.RFC3339), pass, p.Started.Format(time.RFC3339))
		}
		p.Stops = append(p.Stops, StopInterval{Start: now.UTC()})
		p.StopReason = reason
		p.State = Stopped
		rf.Dirty = true
		s.passEvent = true
		return nil
	}
}

// RestopPass changes the reason and detail of a pass that is already stopped.
// It opens no stop interval, so the stretch the pass has spent stopped keeps
// running from the stop that began it.
func RestopPass(project, pass string, reason StopReason, detail string) Event {
	return func(s *State, now time.Time) error {
		if !ValidStopReason(reason) {
			return fmt.Errorf("stop reason %q is not one of %v", reason, StopReasons)
		}
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.State != Stopped {
			return Refuse(fmt.Sprintf("pass %q is %s, not stopped", pass, p.State))
		}
		p.StopReason = reason
		if p.Session == nil {
			p.StopDetail = detail
		}
		rf.Dirty = true
		s.passEvent = true
		return nil
	}
}

// SetCheckout records the repo's main checkout and default branch on the
// project's live release. The path must be absolute and an existing folder.
func SetCheckout(project, path, branch string) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.liveRelease(project)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(path) {
			return Refuse(fmt.Sprintf("checkout %q is not an absolute path", path))
		}
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return Refuse(fmt.Sprintf("checkout %q is not a folder that exists", path))
		}
		if branch == "" {
			return errors.New("the default branch is empty")
		}
		rf.Checkout, rf.DefaultBranch, rf.Dirty = path, branch, true
		return nil
	}
}
