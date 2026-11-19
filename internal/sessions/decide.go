package sessions

import (
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// What a check decides for one session record. Pass 1 prints and records
// each decision and carries none out.
const (
	None           = "none"
	End            = "end"
	Relaunch       = "relaunch"
	EndAndRelaunch = "end_and_relaunch"
	LaunchAgain    = "launch_again"
	Launch         = "launch"
	Start          = "start"
	Stop           = "stop"
	WaitUsage      = "wait_usage"
	WaitLink       = "wait_link"
)

// The causes a decision names that are not stop reasons.
const (
	ReasonLostConnection = "lost_connection"
	ReasonIdle           = "idle"
	ReasonMissing        = "missing"
	ReasonReplaced       = "replaced"
)

// IdleLimit is how long a done session may sit before the check ends it.
const IdleLimit = 15 * time.Minute

// RelaunchLimit is how many counted relaunches a session gets.
const RelaunchLimit = 2

// Key names one session record: the project and the record's place, as
// pass:<name>, round:<name>, or conductor.
type Key struct {
	Project string
	Record  string
}

// RecordEvidence is what the check read for one record. HasRow is false when
// the list held no row for the record's short id.
type RecordEvidence struct {
	Row    Row
	HasRow bool
	Screen string
	Job    Job
}

// ProbeResult is the answer of the network probe, which a run sends at most once.
type ProbeResult int

const (
	ProbeNotSent ProbeResult = iota
	ProbeUp
	ProbeDown
)

// Evidence is everything a run gathered outside the state lock. A record
// with no entry in Records was not covered by the reads and is left alone,
// which is not the same as a covered record with no row.
type Evidence struct {
	Records map[Key]RecordEvidence
	Probe   ProbeResult
	// Pulls holds the pull request read of each pass that had one.
	Pulls map[Key]PullRead
}

// Observed is what a decision asks the record to remember of what the check
// saw: the state to write to last_state, and whether last_seen and
// last_working take the check's time. An empty State writes nothing.
type Observed struct {
	State   string
	Seen    bool
	Working bool
}

// Result is the decision for one record and what to write beside it.
type Result struct {
	state.Decision
	Observed Observed
	// Merge is what carrying out a merge decision reads, and the wait to
	// record beside any decision for a pass with a pull request read.
	Merge MergeInfo
}

// Target is a record whose decision needs a row from the list.
type Target struct {
	Key     Key
	Session *state.Session
}

// Watched lists the records that need a row: those the first five rows of
// the decision table leave undecided.
func Watched(s *state.State) []Target {
	var out []Target
	s.LiveSessions(func(rf *state.ReleaseFile, record string, sess *state.Session) {
		if _, decided := early(rf, record, sess, false); !decided {
			out = append(out, Target{Key{rf.Project, record}, sess})
		}
	})
	return out
}

// Decide gives the decision for every session record of every live release,
// in the order LiveSessions visits them. It reads the usage hold from s, so
// the caller must have folded the current usage reading into s first.
func Decide(s *state.State, ev Evidence, now time.Time) []Result {
	var out []Result
	hold := s.UsageStop()
	s.LiveSessions(func(rf *state.ReleaseFile, record string, sess *state.Session) {
		key := Key{rf.Project, record}
		base := state.Decision{Project: rf.Project, Record: record, Session: sess.Name}
		pass := findPass(rf, record)
		if r, decided := cleanupRow(rf, record, sess); decided {
			out = append(out, r)
			return
		}
		if r, decided := replacedRow(base, sess, pass, ev.Records[key], hold); decided {
			out = append(out, r)
			return
		}
		mr, decided, info, held := decideMerge(base, sess, pass, ev, key, answerOf(s, rf, pass))
		if decided && hold && mr.Decision.Decision == Fix {
			// A fix relaunches the session, so the usage hold waits it out
			// with no attempt spent.
			base.Decision, base.Detail = WaitUsage, mr.Decision.Detail
			out = append(out, Result{Decision: base})
			return
		}
		if decided {
			out = append(out, mr)
			return
		}
		if r, decided := early(rf, record, sess, hold); decided {
			r.Merge = info
			out = append(out, r)
			return
		}
		e, ok := ev.Records[key]
		if !ok {
			base.Decision = None
			base.Detail = "no evidence gathered this run"
			out = append(out, Result{Decision: base, Merge: info})
			return
		}
		r := decideRead(base, sess, e, readyStands(pass), ev.Probe, hold, now)
		r.Merge = info
		if held != "" && r.Decision.Decision == None && r.Detail == "" {
			r.Detail = held
		}
		out = append(out, r)
	})
	return out
}

// Plan runs the record steps of a check on s, in this order: Decide on the
// folded state before any settle, a resume through StartPass for each
// uncounted usage relaunch on a release with a checkout, settle, a start for
// each free pass the limits allow, then one launch decision for each building pass with no session record (wait_usage
// under the usage hold). The caller folds the usage reading into s first and
// records the results. The dry run calls Plan on its in-memory copy.
func Plan(s *state.State, ev Evidence, now time.Time) ([]Result, error) {
	results := Decide(s, ev, now)
	for _, r := range results {
		name, isPass := strings.CutPrefix(r.Record, "pass:")
		if r.Decision.Decision != Relaunch || r.Reason != string(state.StopUsage) || !isPass || !hasCheckout(s, r.Project) {
			continue
		}
		if err := state.StartPass(r.Project, name)(s, now); err != nil {
			return nil, err
		}
	}
	s.Settle(now)
	starts, err := startFree(s, now)
	if err != nil {
		return nil, err
	}
	results = append(results, starts...)
	hold := s.UsageStop()
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.State != state.Building || p.Session != nil {
				continue
			}
			d := state.Decision{Project: rf.Project, Record: "pass:" + p.Name, Session: rf.Project + "-" + p.Name, Decision: Launch}
			if hold {
				d.Decision = WaitUsage
			}
			results = append(results, Result{Decision: d})
		}
	}
	return results, nil
}

// startFree starts each not-started pass that breaks no start rule, in
// release and file order, while no limit holds. A release with no checkout
// starts nothing. Settle has run first, so a waiting pass takes a free slot
// before a free one does.
func startFree(s *state.State, now time.Time) ([]Result, error) {
	var out []Result
	for _, rf := range s.Releases {
		if rf.Done || rf.Checkout == "" {
			continue
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.State != state.NotStarted || len(s.Limits()) > 0 || !s.Free(rf, p) {
				continue
			}
			if err := state.StartPass(rf.Project, p.Name)(s, now); err != nil {
				return nil, err
			}
			d := state.Decision{Project: rf.Project, Record: "pass:" + p.Name, Session: rf.Project + "-" + p.Name, Decision: Start, Reason: "free"}
			out = append(out, Result{Decision: d})
		}
	}
	return out, nil
}

// hasCheckout reports whether the project's live release records a checkout.
func hasCheckout(s *state.State, project string) bool {
	for _, rf := range s.Releases {
		if !rf.Done && rf.Project == project {
			return rf.Checkout != ""
		}
	}
	return false
}

// cleanupRow decides cleanup for a done pass whose pull request merged and
// has not been cleaned up. It runs before every other row, the ended-session
// test included, so a cleanup that failed part way runs again.
func cleanupRow(rf *state.ReleaseFile, record string, sess *state.Session) (Result, bool) {
	p := findPass(rf, record)
	if p == nil || p.State != state.Done || p.PullRequest == nil || p.PullRequest.Merged == nil || p.PullRequest.Cleaned != nil {
		return Result{}, false
	}
	d := state.Decision{Project: rf.Project, Record: record, Session: sess.Name, Decision: Cleanup}
	return Result{Decision: d}, true
}

// answerOf gives the ruling that answered the pass's merge question, or nil
// while the question is open or none is asked.
func answerOf(s *state.State, rf *state.ReleaseFile, p *state.Pass) *Answered {
	if p == nil || p.PullRequest == nil || p.PullRequest.Question == "" {
		return nil
	}
	for _, q := range rf.Questions {
		if q.ID != p.PullRequest.Question || q.Answer == nil {
			continue
		}
		for _, g := range s.Rulings {
			if g.ID == q.Answer.Ruling {
				return &Answered{Ruling: g.ID, Option: g.Option, Words: g.Words}
			}
		}
	}
	return nil
}

// replacedRow decides for a building pass whose answer was replaced and whose
// session is live: end it and relaunch it with the new words, uncounted. It
// runs before the merge rows, so a report made under the old answer merges
// nothing. Under the usage hold the relaunch waits.
func replacedRow(base state.Decision, sess *state.Session, p *state.Pass, e RecordEvidence, hold bool) (Result, bool) {
	if p == nil || p.State != state.Building || p.Replaced == nil || sess.Ended != nil || !e.HasRow || !e.Row.Live() {
		return Result{}, false
	}
	if hold {
		base.Decision, base.Reason = WaitUsage, ReasonReplaced
		return Result{Decision: base}, true
	}
	no := false
	base.Decision, base.Reason, base.Counted = EndAndRelaunch, ReasonReplaced, &no
	return Result{Decision: base}, true
}

// early applies rows 1 to 5, which need no row from the list and write
// nothing to the record. It reports whether one of them decided.
func early(rf *state.ReleaseFile, record string, sess *state.Session, hold bool) (Result, bool) {
	d := state.Decision{Project: rf.Project, Record: record, Session: sess.Name}
	done := func(decision, reason string) (Result, bool) {
		d.Decision, d.Reason = decision, reason
		return Result{Decision: d}, true
	}
	if sess.Ended != nil {
		return done(None, "")
	}
	pass := findPass(rf, record)
	if pass != nil && pass.State == state.Done {
		return done(None, "")
	}
	if pass == nil && sess.Finished != nil {
		return done(End, "")
	}
	var stopped state.StopReason
	if pass != nil {
		if pass.State == state.Stopped {
			stopped = pass.StopReason
		}
	} else {
		stopped = sess.StopReason
	}
	switch {
	case stopped == state.StopUsage && hold:
		return done(WaitUsage, string(state.StopUsage))
	case stopped == state.StopUsage:
		f := false
		d.Counted = &f
		return done(Relaunch, string(state.StopUsage))
	case stopped != "":
		return done(None, string(stopped))
	}
	return Result{}, false
}

var limitText = regexp.MustCompile(`You've hit your .{0,30}limit(?:\s*·\s*resets [^(\n]*(?:\([^)]*\))?)?`)

// resetText gives the limit sentence with its reset time, from the job file
// before the screen.
func resetText(e RecordEvidence) string {
	for _, text := range []string{e.Job.Detail, e.Job.Needs, e.Screen} {
		if m := limitText.FindString(text); m != "" {
			return LastLine(m, "")
		}
	}
	return "no text"
}

// decideRead applies rows 6 to 20 to a record that has evidence.
func decideRead(base state.Decision, sess *state.Session, e RecordEvidence, ready bool, probe ProbeResult, hold bool, now time.Time) Result {
	res := func(decision, reason string) Result {
		base.Decision, base.Reason = decision, reason
		return Result{Decision: base}
	}
	counted := func(r Result) Result {
		t := true
		r.Counted = &t
		return r
	}
	atLimit := sess.Counted() >= RelaunchLimit
	row := e.Row

	if !e.HasRow || row.Class() == ClassNone {
		missing := Observed{State: state.SeenMissing}
		var r Result
		switch {
		case sess.LastSeen == nil && sess.LaunchTries >= 2:
			r = res(Stop, string(state.StopLaunchFailed))
		case sess.LastSeen == nil && hold:
			r = res(WaitUsage, "")
		case sess.LastSeen == nil:
			r = res(LaunchAgain, "")
		case atLimit:
			r = res(Stop, string(state.StopRelaunchLimit))
		case hold:
			r = res(WaitUsage, "")
		default:
			r = counted(res(Relaunch, ReasonMissing))
		}
		r.Observed = missing
		return r
	}

	switch row.Class() {
	case ClassWorking:
		r := res(None, "")
		r.Observed = Observed{State: state.SeenWorking, Seen: true, Working: true}
		return r
	case ClassDone:
		r := doneRow(res, counted, sess, ready, atLimit, hold, now)
		r.Observed = Observed{State: state.SeenIdle, Seen: true}
		return r
	}

	r := blockedRow(res, counted, sess, e, row, atLimit, hold, probe)
	r.Observed = Observed{State: state.SeenBlocked, Seen: true}
	return r
}

func doneRow(res func(string, string) Result, counted func(Result) Result, sess *state.Session, ready, atLimit, hold bool, now time.Time) Result {
	since := sess.Launched
	if sess.LastWorking != nil && sess.LastWorking.After(since) {
		since = *sess.LastWorking
	}
	if now.Sub(since) < IdleLimit || ready {
		return res(None, "")
	}
	switch {
	case atLimit:
		return res(Stop, string(state.StopRelaunchLimit))
	case hold:
		return res(WaitUsage, "")
	}
	return counted(res(EndAndRelaunch, ReasonIdle))
}

// lastPrompt reads the record's last unanswered prompt as "<tool>: <command>",
// or the tool alone when it has no command. It is empty when none is open.
func lastPrompt(sess *state.Session) string {
	for _, p := range slices.Backward(sess.Prompts) {
		if p.Answer != "" {
			continue
		}
		if p.Command == "" {
			return p.Tool
		}
		return p.Tool + ": " + p.Command
	}
	return ""
}

func blockedRow(res func(string, string) Result, counted func(Result) Result, sess *state.Session, e RecordEvidence, row Row, atLimit, hold bool, probe ProbeResult) Result {
	switch Classify(row, e.Job, e.Screen) {
	case LabelPermissionPrompt:
		r := res(Stop, string(state.StopPermissionPrompt))
		r.Detail = lastPrompt(sess)
		return r
	case LabelSignInExpired:
		return res(Stop, string(state.StopSignInExpired))
	case LabelUsageLimit:
		r := res(Stop, string(state.StopUsage))
		r.Detail = resetText(e)
		return r
	case LabelLostConnection:
		switch {
		case atLimit:
			return res(Stop, string(state.StopRelaunchLimit))
		case hold:
			return res(WaitUsage, "")
		case probe == ProbeUp:
			return counted(res(EndAndRelaunch, ReasonLostConnection))
		}
		r := res(WaitLink, "")
		if probe == ProbeNotSent {
			r.Detail = "probe not sent"
		}
		return r
	}
	r := res(Stop, string(state.StopUnreadable))
	r.Detail = LastLine(e.Job.Detail, e.Screen)
	return r
}

// findPass gives the pass a pass:<name> record sits on, or nil for any other
// record.
func findPass(rf *state.ReleaseFile, record string) *state.Pass {
	name, ok := strings.CutPrefix(record, "pass:")
	if !ok {
		return nil
	}
	for i := range rf.Passes {
		if rf.Passes[i].Name == name {
			return &rf.Passes[i]
		}
	}
	return nil
}

// readyStands reports whether a ready report stands on the pass's pull
// request. A record with no pass has none.
func readyStands(p *state.Pass) bool {
	return p != nil && p.PullRequest != nil && p.PullRequest.Ready != nil
}
