package snapshot

import (
	"fmt"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// ListFailureLimit is the count of failed session-list reads in a row that
// puts an entry in needs_you_now.
const ListFailureLimit = 3

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

// todoFor is the one thing to do for a stop reason.
func todoFor(r state.StopReason) string {
	switch r {
	case state.StopPermissionPrompt:
		return "answer it at the desk"
	case state.StopSignInExpired:
		return "sign in to Claude Code"
	case state.StopUsage:
		return "wait for the reset, or reset early"
	case state.StopUnreadable, state.StopRelaunchLimit:
		return "read the session's log, then start the pass again"
	case state.StopLaunchFailed:
		return "read the launch error, then start the pass again"
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
func fillSessions(snap *Snapshot, s *state.State) {
	snap.Sessions = []SessionEntry{}
	snap.NeedsYouNow = []NeedEntry{}
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		add := func(round, pass string, sess *state.Session, stop state.StopReason) {
			e := SessionEntry{
				Project: rf.Project, Round: strp(round), Pass: strp(pass),
				Kind: string(sess.Kind), Name: sess.Name, ShortID: sess.ShortID,
				LastState: strp(sess.LastState), LastSeen: utc(sess.LastSeen), LastWorking: utc(sess.LastWorking),
				Relaunches: sess.Counted(), StopReason: strp(string(stop)),
				Prompts: append([]state.Prompt{}, sess.Prompts...),
			}
			snap.Sessions = append(snap.Sessions, e)
			if stop != "" {
				snap.NeedsYouNow = append(snap.NeedsYouNow, NeedEntry{
					Project: rf.Project, Round: e.Round, Session: strp(sess.Name),
					Reason: string(stop), Cause: causeOf(sess.StopDetail, stop), Todo: todoFor(stop),
				})
			}
		}
		if c := rf.Conductor; c != nil {
			add("", "", c, c.StopReason)
		}
		for i := range rf.Rounds {
			if sess := rf.Rounds[i].Session; sess != nil {
				add(rf.Rounds[i].Name, "", sess, sess.StopReason)
			}
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.Session != nil {
				add(p.Round, p.Name, p.Session, p.StopReason)
				continue
			}
			if p.State == state.Stopped {
				round := p.Round
				snap.NeedsYouNow = append(snap.NeedsYouNow, NeedEntry{
					Project: rf.Project, Round: strp(round),
					Reason: string(p.StopReason), Cause: causeOf(p.StopDetail, p.StopReason), Todo: todoFor(p.StopReason),
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
