package state

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
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
				p.StopReason != "" || p.ActualMinutes != nil || p.Wait != nil || p.Guess != nil {
				return fmt.Errorf("pass %q: a new release's passes must be not_started with no events", p.Name)
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

// Design is a round's passes and questions, as a planning session writes
// them.
type Design struct {
	Format    int        `json:"format"`
	Passes    []Pass     `json:"passes"`
	Questions []Question `json:"questions,omitempty"`
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
				p.StopReason != "" || p.ActualMinutes != nil || p.Wait != nil || p.Guess != nil {
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
		if err := validateRelease(&next); err != nil {
			return fmt.Errorf("design for round %q: %w", round, err)
		}
		rf.Release = next
		rf.Dirty = true
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

// StartPass starts a not-started pass, or resumes a stopped one and closes
// its open stop interval.
func StartPass(project, pass string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		t := now.UTC()
		switch p.State {
		case NotStarted:
			p.Started = &t
		case Stopped:
			last := &p.Stops[len(p.Stops)-1]
			if t.Before(last.Start) {
				return fmt.Errorf("the clock %s is before pass %q stopped at %s", t.Format(time.RFC3339), pass, last.Start.Format(time.RFC3339))
			}
			last.End = &t
			p.StopReason = ""
		default:
			return Refuse(fmt.Sprintf("pass %q is %s and cannot start", pass, p.State))
		}
		p.State = Building
		rf.Dirty = true
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

// StopPass stops a building pass for one of the five reasons and opens a
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
