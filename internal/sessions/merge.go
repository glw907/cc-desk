package sessions

import (
	"fmt"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// What a check decides for a pass whose pull request it read.
const (
	Merge   = "merge"
	Merged  = "merged"
	Fix     = "fix"
	Ask     = "ask"
	Cleanup = "cleanup"
	Wait    = "wait"
)

// The reasons an ask or a fix names.
const (
	ReasonOutsideDesign = "outside_design"
	ReasonFixLimit      = "fix_limit"
	ReasonFix           = "fix"
)

// FixLimit is how many times a pull request goes back to its session before
// the check asks instead.
const FixLimit = 2

// MergeInfo is what carrying out a merge decision needs beyond the decision.
type MergeInfo struct {
	// Wait is the wait to write to the pull request: nil leaves it alone and
	// an empty string clears it.
	Wait *string
	// Number and Commit name the pull request and the commit the decision read.
	Number int
	Commit string
	// Merged is when the pull request merged, set on a merged decision.
	Merged time.Time
	// Outside is set on a merged decision when the check did not call the merge.
	Outside bool
	// Failures names the checks that failed, or says the reviewer asked for a fix.
	Failures []string
}

func waitPtr(s string) *string { return &s }

// decideMerge applies the merge table to a pass whose read came back. It
// reports a decision when one of the rows that skip the row rules matched.
// Otherwise the row rules run, and the returned wait, when not nil, is the
// pull request's wait to record beside them; a non-empty detail is the text
// a held row rule carries. A pass with no read, a done pass, and a pass with
// no pull request record take no part.
func decideMerge(base state.Decision, sess *state.Session, p *state.Pass, ev Evidence, key Key) (res Result, decided bool, info MergeInfo, detail string) {
	if p == nil || p.State == state.Done {
		return
	}
	read, ok := ev.Pulls[key]
	if !ok {
		return
	}
	pr := p.PullRequest
	if read.Unreachable {
		info.Wait = waitPtr(state.WaitUnreachable)
		return
	}
	number := 0
	if pr != nil {
		number = pr.Number
	}
	pull, found := read.Choose(number)
	if found && pull.State == "MERGED" && pull.Merged != nil {
		base.Decision = Merged
		info.Number, info.Commit, info.Merged = pull.Number, pull.Head, *pull.Merged
		info.Outside = pr == nil || pr.MergeCalled == nil
		if info.Outside {
			base.Detail = fmt.Sprintf("pull request %d merged outside cc-desk", pull.Number)
		}
		return Result{Decision: base, Merge: info}, true, info, ""
	}
	if p.State != state.Building {
		return
	}
	clear := waitPtr("")
	if !found || pull.State == "CLOSED" || pr == nil || pr.Ready == nil {
		info.Wait = clear
		return
	}
	info.Number, info.Commit = pr.Number, pr.Commit
	if sess.Ended != nil {
		info.Wait = clear
		return
	}
	switch {
	case pr.Question != "":
		info.Wait = clear
		detail = "question " + pr.Question + " asked"
		return
	case !pr.InsideDesign:
		base.Decision, base.Reason = Ask, ReasonOutsideDesign
		base.Detail = fmt.Sprintf("pull request %d is outside the approved design", pr.Number)
		info.Failures = nil
		return Result{Decision: base, Merge: info}, true, info, ""
	case pull.Head != pr.Commit:
		info.Wait = waitPtr(state.WaitCommitMoved)
		detail = fmt.Sprintf("head %s is not the reported %s", short(pull.Head), short(pr.Commit))
		return
	case len(pull.Checks) == 0:
		info.Wait = waitPtr(state.WaitNoCheck)
		detail = "no check registered"
		return
	}
	var failed []string
	pending := false
	for _, c := range pull.Checks {
		switch c.Bucket {
		case Fail, Cancel:
			failed = append(failed, c.Name)
		case Pending:
			pending = true
		}
	}
	if pending {
		info.Wait = waitPtr(state.WaitCheckRunning)
		detail = "a check is running"
		return
	}
	if len(failed) > 0 || pr.Verdict == state.VerdictFix {
		if len(failed) == 0 {
			failed = []string{"the reviewer asked for a fix"}
		}
		info.Failures = failed
		base.Detail = fmt.Sprintf("pull request %d: %s", pr.Number, strings.Join(failed, ", "))
		if pr.FixAttempts >= FixLimit {
			base.Decision, base.Reason = Ask, ReasonFixLimit
			return Result{Decision: base, Merge: info}, true, info, ""
		}
		no := false
		base.Decision, base.Reason, base.Counted = Fix, ReasonFix, &no
		return Result{Decision: base, Merge: info}, true, info, ""
	}
	base.Decision = Merge
	base.Detail = fmt.Sprintf("pull request %d at %s", pr.Number, short(pr.Commit))
	return Result{Decision: base, Merge: info}, true, info, ""
}

func short(sha string) string {
	if len(sha) > 8 {
		return sha[:8]
	}
	return sha
}
