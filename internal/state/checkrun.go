package state

import (
	"fmt"
	"strings"
	"time"
)

// Recorded is one decision of a session check with what the check saw of the
// session. RecordCheck writes both. LastState is empty when the check read
// no row for the record.
type Recorded struct {
	Decision  Decision
	LastState string
	Seen      bool
	Working   bool
	// Stop, when set, stops the record for that reason, with the decision's
	// detail as the stop detail.
	Stop StopReason
	// PRWait, when set, is the wait to write to the pass's pull request,
	// and an empty string clears it. A nil pointer leaves it alone.
	PRWait *string
}

// RecordCheck writes a run's decisions to check.json and, for each record,
// the fields the check read: last_state, last_seen, last_working, and a stop.
// It resets the count of failed list reads. It never touches launch_tries or
// relaunches, which the run that carries a decision out owns.
func RecordCheck(recs []Recorded) Event {
	return func(s *State, now time.Time) error {
		t := now.UTC()
		decisions := make([]Decision, 0, len(recs))
		for _, r := range recs {
			if err := applyRecorded(s, r, t); err != nil {
				return err
			}
			if r.Stop != "" {
				s.logCheck(t, r.Decision, "stopped: "+string(r.Stop))
			}
			decisions = append(decisions, r.Decision)
		}
		s.Check = CheckFile{Format: Format, LastRun: &t, Decisions: decisions}
		s.checkDirty = true
		return nil
	}
}

// logCheck adds one line to the check log for Commit to append, and to the
// loaded lines so the snapshot of the same commit shows it.
func (s *State) logCheck(now time.Time, d Decision, outcome string) {
	l := CheckLogLine{
		Format: Format, Time: now.UTC(), Project: d.Project, Record: d.Record,
		Decision: d.Decision, Outcome: outcome, Reason: d.Reason, Detail: d.Detail,
	}
	s.newCheckLog = append(s.newCheckLog, l)
	s.CheckLog = append(s.CheckLog, l)
}

func applyRecorded(s *State, r Recorded, now time.Time) error {
	var rf *ReleaseFile
	var sess *Session
	s.LiveSessions(func(f *ReleaseFile, record string, ss *Session) {
		if f.Project == r.Decision.Project && record == r.Decision.Record {
			rf, sess = f, ss
		}
	})
	if sess == nil {
		// A building pass with no session record has only the decision to
		// write: the launch makes the record.
		if d := r.Decision.Decision; (d == "launch" || d == "wait_usage" || d == "start") && strings.HasPrefix(r.Decision.Record, "pass:") {
			return nil
		}
		return fmt.Errorf("no session record %s in project %q", r.Decision.Record, r.Decision.Project)
	}
	if r.LastState != "" && sess.LastState != r.LastState {
		sess.LastState = r.LastState
		rf.Dirty = true
	}
	if r.Seen {
		sess.LastSeen = &now
		rf.Dirty = true
	}
	if r.Working {
		sess.LastWorking = &now
		rf.Dirty = true
	}
	if r.PRWait != nil {
		if name, ok := strings.CutPrefix(r.Decision.Record, "pass:"); ok {
			if _, p, err := findPass(s, r.Decision.Project, name); err == nil && p.PullRequest != nil && p.PullRequest.Wait != *r.PRWait {
				p.PullRequest.Wait = *r.PRWait
				rf.Dirty = true
			}
		}
	}
	if r.Stop == "" {
		return nil
	}
	if name, ok := strings.CutPrefix(r.Decision.Record, "pass:"); ok {
		ev := StopPass(r.Decision.Project, name, r.Stop)
		if _, p, err := findPass(s, r.Decision.Project, name); err == nil && p.State == Stopped {
			// A pass already stopped on usage that reaches relaunch_limit
			// changes its stop's reason and opens no second interval.
			ev = RestopPass(r.Decision.Project, name, r.Stop, r.Decision.Detail)
		}
		if err := ev(s, now); err != nil {
			return err
		}
	} else if sess.StopReason != "" {
		sess.StopReason = r.Stop
	} else {
		sess.StopReason, sess.Stopped = r.Stop, &now
	}
	sess.StopDetail = r.Decision.Detail
	rf.Dirty = true
	return nil
}

// RecordListFailure counts one failed read of the session list in
// check.json. It runs as its own commit: it skips settle, so a waiting pass
// stays waiting, and it touches no release file and appends no forecast line.
func RecordListFailure() Event {
	return func(s *State, now time.Time) error {
		s.Check.Format = Format
		s.Check.ListFailures++
		s.checkDirty = true
		s.noSettle = true
		return nil
	}
}
