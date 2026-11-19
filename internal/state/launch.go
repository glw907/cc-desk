package state

import (
	"fmt"
	"strings"
	"time"
)

// LaunchKind says which launch a record is: the first of a session, the
// second try of a first that never showed a row, or a relaunch.
type LaunchKind int

const (
	LaunchFirst LaunchKind = iota
	LaunchAgain
	LaunchRelaunch
)

// Launch is what a launch learned: the identity of the new session, and for
// a relaunch the cause and whether it counts toward the limit of two.
// Launched, the tries, and the relaunch time come from the event's clock.
type Launch struct {
	Kind    LaunchKind
	Session Session
	// Replaces is the short id of the record an again or a relaunch
	// replaces. The event refuses when the record now holds another.
	Replaces string
	Cause    string
	Counted  bool
}

// sessionAt finds where a record sits: the release, a pointer to the
// record's slot, and the pass for a pass record.
func sessionAt(s *State, project, record string) (*ReleaseFile, **Session, *Pass, error) {
	rf, err := s.liveRelease(project)
	if err != nil {
		return nil, nil, nil, err
	}
	switch {
	case record == "conductor":
		return rf, &rf.Conductor, nil, nil
	case strings.HasPrefix(record, "round:"):
		for i := range rf.Rounds {
			if rf.Rounds[i].Name == strings.TrimPrefix(record, "round:") {
				return rf, &rf.Rounds[i].Session, nil, nil
			}
		}
	case strings.HasPrefix(record, "pass:"):
		p, err := rf.pass(strings.TrimPrefix(record, "pass:"))
		if err != nil {
			return nil, nil, nil, err
		}
		return rf, &p.Session, p, nil
	}
	return nil, nil, nil, fmt.Errorf("no record %s in project %q", record, project)
}

// recordKind gives the session kind a record's place holds.
func recordKind(record string) SessionKind {
	switch {
	case record == "conductor":
		return KindConductor
	case strings.HasPrefix(record, "round:"):
		return KindDesign
	}
	return KindPass
}

// RecordLaunch writes the session record a launch made, in one commit. A
// first launch needs a place with no live record and makes one with
// launch_tries 1. A launch again needs the record that never showed a row
// and sets launch_tries to 2. A relaunch keeps the record, replaces its ids,
// appends its entry, and clears what the check saw of the old session, so
// the never-started rule applies to the new one.
//
// A pass's record is written only while the pass is building. The result
// must pass validateSession and checkSessionsUnique before it replaces
// anything.
func RecordLaunch(project, record string, l Launch) Event {
	return func(s *State, now time.Time) error {
		rf, slot, p, err := sessionAt(s, project, record)
		if err != nil {
			return err
		}
		if p != nil && p.State != Building {
			return Refuse(fmt.Sprintf("pass %q is %s, not building, so a launch cannot be recorded", p.Name, p.State))
		}
		t := now.UTC()
		old := *slot
		next := l.Session
		next.Launched = t
		switch l.Kind {
		case LaunchFirst:
			if old != nil && old.Ended == nil {
				return Refuse(fmt.Sprintf("%s/%s already has a live session record, %s", project, record, old.Name))
			}
			next.LaunchTries = 1
		default:
			if old == nil || old.Ended != nil || old.ShortID != l.Replaces {
				return Refuse(fmt.Sprintf("%s/%s is no longer the live record %q this launch replaces", project, record, l.Replaces))
			}
			next.Prompts, next.Finished = old.Prompts, old.Finished
			next.Relaunches = append([]Relaunch(nil), old.Relaunches...)
			next.LaunchTries = 2
			if l.Kind == LaunchRelaunch {
				next.Relaunches = append(next.Relaunches, Relaunch{Time: t, Cause: l.Cause, Counted: l.Counted})
				next.LaunchTries = 1
			}
		}
		if err := validateSession(&next, recordKind(record), project+"/"+record); err != nil {
			return err
		}
		*slot = &next
		if err := checkSessionsUnique(s); err != nil {
			*slot = old
			return err
		}
		rf.Dirty = true
		return nil
	}
}

// ClearReplaced clears a pass's replaced mark and the ready report of its
// pull request, which describes the work before the replacement. It composes
// with RecordLaunch in one commit.
func ClearReplaced(project, pass string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		p.Replaced = nil
		if p.PullRequest != nil {
			p.PullRequest.Ready = nil
		}
		rf.Dirty = true
		return nil
	}
}

// EndSession records that the check ended the record's session.
func EndSession(project, record string) Event {
	return func(s *State, now time.Time) error {
		rf, slot, _, err := sessionAt(s, project, record)
		if err != nil {
			return err
		}
		if *slot == nil {
			return fmt.Errorf("%s/%s has no session record to end", project, record)
		}
		t := now.UTC()
		(*slot).Ended = &t
		rf.Dirty = true
		return nil
	}
}

// FinishSession marks a conductor's or design writer's session finished, so
// the next check ends it and never relaunches it.
func FinishSession(project, record string) Event {
	return func(s *State, now time.Time) error {
		rf, slot, p, err := sessionAt(s, project, record)
		if err != nil {
			return err
		}
		if p != nil {
			return fmt.Errorf("%s/%s is a pass; only a conductor or a design writer is finished", project, record)
		}
		if *slot == nil || (*slot).Ended != nil {
			return Refuse(fmt.Sprintf("%s/%s has no live session to finish", project, record))
		}
		t := now.UTC()
		(*slot).Finished = &t
		rf.Dirty = true
		return nil
	}
}

// StopSessionOnLaunchFailure ends a conductor's or design writer's record
// when a launch that replaced it failed. The record carries launch_failed and
// the error, so the next check does not retry it and the snapshot names it.
// A later session start replaces the ended record.
func StopSessionOnLaunchFailure(project, record, detail string) Event {
	return func(s *State, now time.Time) error {
		rf, slot, p, err := sessionAt(s, project, record)
		if err != nil {
			return err
		}
		if p != nil {
			return fmt.Errorf("%s/%s is a pass; stop the pass instead", project, record)
		}
		if *slot == nil {
			return fmt.Errorf("%s/%s has no session record to stop", project, record)
		}
		t := now.UTC()
		sess := *slot
		sess.StopReason, sess.Stopped, sess.StopDetail, sess.Ended = StopLaunchFailed, &t, detail, &t
		rf.Dirty = true
		return nil
	}
}

// RecordOutcomes writes what the check carried out into check.json, one
// outcome for each decision of the last run, in the order they were
// recorded. A count that differs from the decisions' means another run wrote
// check.json in between, and the event refuses.
func RecordOutcomes(outcomes []string) Event {
	return func(s *State, now time.Time) error {
		if len(outcomes) != len(s.Check.Decisions) {
			return fmt.Errorf("check.json holds %d decisions, not the %d this run recorded", len(s.Check.Decisions), len(outcomes))
		}
		for i := range s.Check.Decisions {
			s.Check.Decisions[i].Outcome = outcomes[i]
			if outcomes[i] != "" {
				s.logCheck(now, s.Check.Decisions[i], outcomes[i])
			}
		}
		s.checkDirty = true
		return nil
	}
}

// StopOnLaunchFailure stops a building pass as launch_failed with the error
// as its stop_detail, and as its record's when it has one.
func StopOnLaunchFailure(project, pass, detail string) Event {
	return func(s *State, now time.Time) error {
		if err := StopPass(project, pass, StopLaunchFailed)(s, now); err != nil {
			return err
		}
		_, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		p.StopDetail = detail
		if p.Session != nil {
			p.Session.StopDetail = detail
		}
		return nil
	}
}
