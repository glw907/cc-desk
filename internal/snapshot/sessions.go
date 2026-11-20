package snapshot

import (
	"fmt"
	"strconv"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// ListFailureLimit is the count of failed session-list reads in a row that
// puts an entry in needs_you_now.
const ListFailureLimit = 3

// NoCheckAfter is how long a pull request reported ready may have no check
// registered before the snapshot asks Geoff to look.
const NoCheckAfter = 30 * time.Minute

// NoCheckRegistered is the reason of the needs_you_now entry for a pull
// request whose checks never started.
const NoCheckRegistered = "no_check_registered"

// ListUnreadable is the reason of the needs_you_now entry for failed list
// reads. It is not a stop reason and sits on no record.
const ListUnreadable = "session_list_unreadable"

// SessionEntry is one session record of a live release. Pass is null for a
// conductor and a design writer; Round is null for a conductor.
type SessionEntry struct {
	Project     string         `json:"project"`
	Round       *string        `json:"round"`
	Pass        *string        `json:"pass"`
	Kind        string         `json:"kind"`
	Name        string         `json:"name"`
	ShortID     string         `json:"short_id"`
	LastState   *string        `json:"last_state"`
	LastSeen    *time.Time     `json:"last_seen"`
	LastWorking *time.Time     `json:"last_working"`
	Relaunches  int            `json:"relaunches"`
	StopReason  *string        `json:"stop_reason"`
	Prompts     []state.Prompt `json:"prompts"`
	// StoppedAt is when the session's stop began: a conductor's or round's
	// stop time, or the open stop of a pass's. Null when it is not stopped.
	StoppedAt *time.Time      `json:"stopped_at"`
	Refusals  []state.Refusal `json:"refusals"`
	// Cleanup is the note a merged pass's cleanup left: what stayed in place.
	Cleanup string `json:"cleanup,omitempty"`
}

// NeedEntry is one cause that needs Geoff now. Session and Round are null
// for the list entry, and Round for a conductor.
type NeedEntry struct {
	Project string  `json:"project"`
	Round   *string `json:"round"`
	Session *string `json:"session"`
	Reason  string  `json:"reason"`
	Cause   string  `json:"cause"`
	Todo    string  `json:"todo"`
}

// LastCheck is the last run of the session check.
type LastCheck struct {
	Time         time.Time        `json:"time"`
	ListFailures int              `json:"list_failures"`
	Decisions    []state.Decision `json:"decisions"`
}

// todoFor is the one thing to do for a stop reason. For a pass, the line
// names the command that starts it again, since the desk runs it when Geoff
// says to. A conductor or design writer has no pass to start, so its line
// keeps the plain words. resumes reports whether the check resumes the
// pass's sign-in stop by itself, as state.SignInResumes says.
func todoFor(r state.StopReason, project, pass string, resumes bool) string {
	start := "start the pass again"
	if pass != "" {
		start = "run cc-desk pass start " + project + " " + pass
	}
	switch r {
	case state.StopPermissionPrompt:
		return "answer it at the desk"
	case state.StopSignInExpired:
		if pass == "" {
			return "sign in to Claude Code"
		}
		if resumes {
			return "sign in to Claude Code, and the pass resumes by itself"
		}
		return "sign in to Claude Code, then " + start
	case state.StopUsage:
		return "wait for the reset, or reset early"
	case state.StopUnreadable, state.StopRelaunchLimit:
		return "read the session's log, then " + start
	case state.StopLaunchFailed:
		return "read the launch error, then " + start
	}
	return "read the stop in the repo's STATUS"
}

func strp(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// fillSessions sets sessions, needs_you_now, and last_check from the live
// releases and check.json.
func fillSessions(snap *Snapshot, s *state.State, now time.Time) {
	snap.Sessions = []SessionEntry{}
	snap.NeedsYouNow = []NeedEntry{}
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		// add records a session. A conductor's or round's record that has
		// ended with a stop leaves needs_you_now, unless the stop is a launch
		// that failed or the limit of relaunches. A pass's stop stays, ended
		// session or not.
		add := func(round, pass string, sess *state.Session, stop state.StopReason, cleanup string, stoppedAt *time.Time, mayEnd, resumes bool) {
			e := SessionEntry{
				Project: rf.Project, Round: strp(round), Pass: strp(pass),
				Kind: string(sess.Kind), Name: sess.Name, ShortID: sess.ShortID,
				LastState: strp(sess.LastState), LastSeen: utc(sess.LastSeen), LastWorking: utc(sess.LastWorking),
				Relaunches: sess.Counted(), StopReason: strp(string(stop)),
				Prompts: append([]state.Prompt{}, sess.Prompts...), Cleanup: cleanup,
				StoppedAt: stoppedAt, Refusals: append([]state.Refusal{}, sess.Refusals...),
			}
			snap.Sessions = append(snap.Sessions, e)
			if stop != "" && !(mayEnd && sess.Ended != nil && stop != state.StopLaunchFailed && stop != state.StopRelaunchLimit) {
				snap.NeedsYouNow = append(snap.NeedsYouNow, NeedEntry{
					Project: rf.Project, Round: e.Round, Session: strp(sess.Name),
					Reason: string(stop), Cause: causeOf(sess.StopDetail, stop),
					Todo: todoFor(stop, rf.Project, pass, resumes),
				})
			}
		}
		if c := rf.Conductor; c != nil {
			add("", "", c, c.StopReason, "", utc(c.Stopped), true, false)
		}
		for i := range rf.Rounds {
			if sess := rf.Rounds[i].Session; sess != nil {
				add(rf.Rounds[i].Name, "", sess, sess.StopReason, "", utc(sess.Stopped), true, false)
			}
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if pr := p.PullRequest; pr != nil && p.State == state.Building && pr.Wait == state.WaitNoCheck && pr.Ready != nil && now.Sub(*pr.Ready) >= NoCheckAfter {
				e := NeedEntry{
					Project: rf.Project, Round: strp(p.Round), Reason: NoCheckRegistered, Cause: strconv.Itoa(pr.Number),
					Todo: fmt.Sprintf("See why no check has started on pull request %d", pr.Number),
				}
				if p.Session != nil {
					e.Session = strp(p.Session.Name)
				}
				snap.NeedsYouNow = append(snap.NeedsYouNow, e)
			}
			if p.Session != nil {
				cleanup := ""
				if p.PullRequest != nil {
					cleanup = p.PullRequest.CleanupNote
				}
				add(p.Round, p.Name, p.Session, p.StopReason, cleanup, openStop(p), false, state.SignInResumes(rf, p, p.Session))
				continue
			}
			if p.State == state.Stopped {
				round := p.Round
				snap.NeedsYouNow = append(snap.NeedsYouNow, NeedEntry{
					Project: rf.Project, Round: strp(round),
					Reason: string(p.StopReason), Cause: causeOf(p.StopDetail, p.StopReason), Todo: todoFor(p.StopReason, rf.Project, p.Name, false),
				})
			}
		}
	}
	if s.Check.ListFailures >= ListFailureLimit {
		snap.NeedsYouNow = append(snap.NeedsYouNow, NeedEntry{
			Reason: ListUnreadable, Cause: fmt.Sprintf("the session list could not be read in the last %d checks", s.Check.ListFailures),
			Todo: "run claude agents in a terminal and read its error",
		})
	}
	if s.Check.LastRun != nil {
		d := append([]state.Decision{}, s.Check.Decisions...)
		snap.LastCheck = &LastCheck{Time: s.Check.LastRun.UTC(), ListFailures: s.Check.ListFailures, Decisions: d}
	}
}

func causeOf(detail string, r state.StopReason) string {
	if detail != "" {
		return detail
	}
	return string(r)
}
