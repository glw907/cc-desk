package state

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

// Sequence runs the events in order as one event, so one commit holds them.
// The first error ends it.
func Sequence(evs ...Event) Event {
	return func(s *State, now time.Time) error {
		for _, ev := range evs {
			if err := ev(s, now); err != nil {
				return err
			}
		}
		return nil
	}
}

// ReportReady records a pass session's report that its pull request is ready:
// the number, the last commit, the reviewer's verdict (accept, fix, or
// escalate), whether the change stays inside the approved design, and the
// count of calls it leaves for Geoff. It refuses a pass that is not building
// or has no session record. The report clears the wait, failures, question,
// and merge marker of any earlier one, and keeps the fix attempts, and the
// merge failures when the pull request's number is unchanged. A second
// report replaces the first.
func ReportReady(project, pass string, number int, commit, verdict string, inside bool, calls int) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		switch {
		case number <= 0:
			return Refuse("the pull request number must be above 0")
		case commit == "":
			return Refuse("the commit is empty")
		case !validVerdict(verdict):
			return Refuse(fmt.Sprintf("verdict %q is not accept, fix, or escalate", verdict))
		case calls < 0:
			return Refuse("the calls for Geoff cannot be below 0")
		case p.State != Building:
			return Refuse(fmt.Sprintf("pass %q is %s, so it cannot report a pull request ready", pass, p.State))
		case p.Session == nil:
			return Refuse(fmt.Sprintf("pass %q has no session record", pass))
		}
		t := now.UTC()
		attempts, failed := 0, 0
		if p.PullRequest != nil {
			attempts = p.PullRequest.FixAttempts
			if p.PullRequest.Number == number {
				failed = p.PullRequest.MergeFailures
			}
		}
		p.PullRequest = &PullRequest{
			Number: number, Commit: commit, Verdict: verdict, InsideDesign: inside, Calls: calls,
			FixAttempts: attempts, MergeFailures: failed, Ready: &t,
		}
		rf.Dirty = true
		return nil
	}
}

// MergePass records that the pass's pull request merged at the given time:
// by the check's own call, or by another hand when outside is set. It closes
// an open stop at that time, finishes the pass there with its actual
// minutes, and marks the pull request merged. A time before the pass started
// or before its open stop began is raised to that moment, so no stored
// interval runs backward. It refuses a pass that has not started or is done.
func MergePass(project, pass string, number int, at time.Time, outside bool) Event {
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
			return Refuse(fmt.Sprintf("pass %q has not started", pass))
		}
		t := at.UTC()
		if t.Before(*p.Started) {
			t = *p.Started
		}
		if p.State == Stopped {
			last := &p.Stops[len(p.Stops)-1]
			if t.Before(last.Start) {
				t = last.Start
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
		if p.PullRequest == nil {
			p.PullRequest = &PullRequest{}
		}
		pr := p.PullRequest
		pr.Number, pr.Merged, pr.MergedOutside = number, &t, outside
		pr.Wait, pr.Failures, pr.MergeFailures = "", nil, 0
		rf.Dirty = true
		s.passEvent = true
		return nil
	}
}

// MarkMerge records that the check is about to call the merge, before it
// does. A check that crashes between the call and its record still knows the
// merge was its own.
func MarkMerge(project, pass string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.PullRequest == nil {
			return Refuse(fmt.Sprintf("pass %q has no pull request record", pass))
		}
		t := now.UTC()
		p.PullRequest.MergeCalled = &t
		rf.Dirty = true
		return nil
	}
}

// RecordMergeFailure counts one failed or timed-out merge call on the pass's
// pull request. It composes with AskMerge in one commit, which clears the
// count again.
func RecordMergeFailure(project, pass string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.PullRequest == nil {
			return Refuse(fmt.Sprintf("pass %q has no pull request record", pass))
		}
		p.PullRequest.MergeFailures++
		rf.Dirty = true
		return nil
	}
}

// RecordFix records that a failed or unfit pull request is sent back to its
// pass session: one more attempt, the failures saved, and the report
// cleared, so the check decides nothing more for the pull request until a
// new report arrives. It clears an answered merge question, so the answer is
// carried out once. It composes with RecordLaunch in one commit.
func RecordFix(project, pass string, failures []string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.PullRequest == nil || p.PullRequest.Ready == nil {
			return Refuse(fmt.Sprintf("pass %q has no ready report to send back", pass))
		}
		pr := p.PullRequest
		pr.FixAttempts++
		pr.Ready, pr.Wait = nil, ""
		pr.Question, pr.AskedCommit = "", ""
		pr.Failures = slices.Clone(failures)
		rf.Dirty = true
		return nil
	}
}

// AskMerge adds a merge question for the pass and records it on the pull
// request, whose report stays, along with the commit it asks about. The id is
// merge-<pass>-<k>, with k one more than the pass's merge questions so far.
// Every later merge decision for the pull request holds until the question is
// answered or a new report clears it. It clears the merge failures, so each
// "merge" answer gets a full set of calls.
func AskMerge(project, pass, text, commit string, options []string, recommended string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.PullRequest == nil || p.PullRequest.Ready == nil {
			return Refuse(fmt.Sprintf("pass %q has no ready report to ask about", pass))
		}
		prefix := "merge-" + pass + "-"
		k := 1
		for _, q := range rf.Questions {
			if strings.HasPrefix(q.ID, prefix) {
				k++
			}
		}
		id := fmt.Sprintf("%s%d", prefix, k)
		next := rf.Release
		next.Passes = slices.Clone(rf.Passes)
		next.Questions = append(slices.Clone(rf.Questions), Question{ID: id, Text: text, Options: options, Recommended: recommended})
		for i := range next.Passes {
			if next.Passes[i].Name == pass {
				next.Passes[i].Questions = append(slices.Clone(next.Passes[i].Questions), id)
			}
		}
		if err := validateRelease(&next); err != nil {
			return fmt.Errorf("merge question for pass %q: %w", pass, err)
		}
		rf.Release = next
		p, _ = rf.pass(pass)
		p.PullRequest.Question, p.PullRequest.AskedCommit = id, commit
		p.PullRequest.MergeFailures = 0
		rf.Dirty = true
		return nil
	}
}

// RecordCleanup records the end of a merged pass's cleanup at the check's
// clock, with the notes of what it left in place.
func RecordCleanup(project, pass string, notes []string) Event {
	return func(s *State, now time.Time) error {
		rf, p, err := findPass(s, project, pass)
		if err != nil {
			return err
		}
		if p.PullRequest == nil || p.PullRequest.Merged == nil {
			return Refuse(fmt.Sprintf("pass %q has no merged pull request to clean up", pass))
		}
		t := now.UTC()
		p.PullRequest.Cleaned = &t
		p.PullRequest.CleanupNote = strings.Join(notes, "; ")
		rf.Dirty = true
		return nil
	}
}

// validVerdict reports whether v is one of the three reviewer verdicts.
func validVerdict(v string) bool {
	return v == VerdictAccept || v == VerdictFix || v == VerdictEscalate
}
