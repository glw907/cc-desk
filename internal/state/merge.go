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
// the number, the last commit, the reviewer's verdict, and whether the change
// stays inside the approved design. It refuses a pass that is not building
// or has no session record. The report clears the wait, failures, question,
// and merge marker of any earlier one, and keeps the fix attempts. A second
// report replaces the first.
func ReportReady(project, pass string, number int, commit, verdict string, inside bool) Event {
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
		case verdict != VerdictAccept && verdict != VerdictFix:
			return Refuse(fmt.Sprintf("verdict %q is not accept or fix", verdict))
		case p.State != Building:
			return Refuse(fmt.Sprintf("pass %q is %s, so it cannot report a pull request ready", pass, p.State))
		case p.Session == nil:
			return Refuse(fmt.Sprintf("pass %q has no session record", pass))
		}
		t := now.UTC()
		attempts := 0
		if p.PullRequest != nil {
			attempts = p.PullRequest.FixAttempts
		}
		p.PullRequest = &PullRequest{
			Number: number, Commit: commit, Verdict: verdict, InsideDesign: inside,
			FixAttempts: attempts, Ready: &t,
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
		pr.Wait, pr.Failures = "", nil
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

// RecordFix records that a failed or unfit pull request is sent back to its
// pass session: one more attempt, the failures saved, and the report
// cleared, so the check decides nothing more for the pull request until a
// new report arrives. It composes with RecordLaunch in one commit.
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
		pr.Failures = slices.Clone(failures)
		rf.Dirty = true
		return nil
	}
}

// AskMerge adds a merge question for the pass and records it on the pull
// request, whose report stays. The id is merge-<pass>-<k>, with k one more
// than the pass's merge questions so far. Every later merge decision for the
// pull request holds until a new report clears the question.
func AskMerge(project, pass, text string, options []string, recommended string) Event {
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
		p.PullRequest.Question = id
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
