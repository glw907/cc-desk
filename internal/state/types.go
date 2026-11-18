// Package state holds the cc-desk state files: their types, the strict
// reader and checks, the lock, and the one commit path every verb runs.
package state

import "time"

// Format is the version every state file and JSONL line carries.
const Format = 1

// PassState is where a pass is in its life.
type PassState string

const (
	NotStarted PassState = "not_started"
	Waiting    PassState = "waiting"
	Building   PassState = "building"
	Stopped    PassState = "stopped"
	Done       PassState = "done"
)

// StopReason is why a pass stopped. Each one needs Geoff.
type StopReason string

const (
	StopUsage        StopReason = "usage"
	StopIrreversible StopReason = "irreversible"
	StopSecret       StopReason = "secret"
	StopAuthData     StopReason = "auth_data"
	StopScope        StopReason = "scope"

	// The reasons the session check records. The page shows each as words.
	StopPermissionPrompt StopReason = "permission_prompt"
	StopSignInExpired    StopReason = "sign_in_expired"
	StopUnreadable       StopReason = "unreadable_stop"
	StopRelaunchLimit    StopReason = "relaunch_limit"
	StopLaunchFailed     StopReason = "launch_failed"
)

// StopReasons lists every stop reason: the design's, then the check's.
var StopReasons = []StopReason{
	StopUsage, StopIrreversible, StopSecret, StopAuthData, StopScope,
	StopPermissionPrompt, StopSignInExpired, StopUnreadable, StopRelaunchLimit, StopLaunchFailed,
}

// SessionListUnreadable is the reason the snapshot gives after three failed
// reads of the session list in a row. It is not a stop reason and sits on
// no record.
const SessionListUnreadable = "session_list_unreadable"

// ListFailureLimit is how many failed list reads in a row make the snapshot
// say so.
const ListFailureLimit = 3

// SessionKind says what a session record's session works on.
type SessionKind string

const (
	KindPass      SessionKind = "pass"
	KindDesign    SessionKind = "design"
	KindConductor SessionKind = "conductor"
)

// The states the check records in a session's last_state.
const (
	SeenWorking = "working"
	SeenBlocked = "blocked"
	SeenIdle    = "idle"
	SeenMissing = "missing"
)

// Session is one background session and what the check knows of it. It sits
// on the pass, round, or release it works on, and is absent until a launch
// records it. A record without Ended is live.
type Session struct {
	Name        string      `json:"name"`
	Kind        SessionKind `json:"kind"`
	Model       string      `json:"model,omitempty"`
	Effort      string      `json:"effort,omitempty"`
	ShortID     string      `json:"short_id"`
	SessionID   string      `json:"session_id"`
	Worktree    string      `json:"worktree"`
	Branch      string      `json:"branch"`
	Launched    time.Time   `json:"launched"`
	LaunchTries int         `json:"launch_tries"`
	Relaunches  []Relaunch  `json:"relaunches,omitempty"`
	LastSeen    *time.Time  `json:"last_seen,omitempty"`
	LastState   string      `json:"last_state,omitempty"`
	LastWorking *time.Time  `json:"last_working,omitempty"`
	Prompts     []Prompt    `json:"prompts,omitempty"`
	Finished    *time.Time  `json:"finished,omitempty"`
	// StopReason and Stopped are set on a conductor's or design writer's
	// record only. A pass's stop stays on the pass.
	StopReason StopReason `json:"stop_reason,omitempty"`
	Stopped    *time.Time `json:"stopped,omitempty"`
	// StopDetail is the cause the check read at the session's last stop.
	StopDetail string     `json:"stop_detail,omitempty"`
	Ended      *time.Time `json:"ended,omitempty"`
}

// Counted reports how many of the session's relaunches count toward the
// limit of two.
func (s *Session) Counted() int {
	n := 0
	for _, r := range s.Relaunches {
		if r.Counted {
			n++
		}
	}
	return n
}

// Relaunch is one relaunch of a session. One after a prompt's answer or a
// fix does not count.
type Relaunch struct {
	Time    time.Time `json:"time"`
	Cause   string    `json:"cause"`
	Counted bool      `json:"counted"`
}

// Prompt is one permission prompt a session stopped on.
type Prompt struct {
	Time          time.Time `json:"time"`
	Tool          string    `json:"tool"`
	Command       string    `json:"command,omitempty"`
	SuggestedRule string    `json:"suggested_rule,omitempty"`
	Answer        string    `json:"answer,omitempty"`
}

// PullRequest is the pull request a pass's session opened.
type PullRequest struct {
	Number      int        `json:"number"`
	Commit      string     `json:"commit,omitempty"`
	Verdict     string     `json:"verdict,omitempty"`
	FixAttempts int        `json:"fix_attempts,omitempty"`
	Ready       *time.Time `json:"ready,omitempty"`
	Merged      *time.Time `json:"merged,omitempty"`
}

// Release is one live release of a project: its rounds, passes, and the
// questions they rest on. It is the content of releases/<project>.json.
type Release struct {
	Format  int    `json:"format"`
	Project string `json:"project"`
	Name    string `json:"name"`
	Repo    string `json:"repo"`
	// Checkout is the absolute path of the repo's main checkout, and
	// DefaultBranch its default branch. They are set together.
	Checkout      string `json:"checkout,omitempty"`
	DefaultBranch string `json:"default_branch,omitempty"`
	// AllowRules are exact Claude Code permission rules the launch writes
	// into each session's settings file, beside the report verbs.
	AllowRules []string   `json:"allow_rules,omitempty"`
	Conductor  *Session   `json:"conductor,omitempty"`
	Rounds     []Round    `json:"rounds,omitempty"`
	Passes     []Pass     `json:"passes,omitempty"`
	Questions  []Question `json:"questions,omitempty"`
}

// Round is a group of passes that share a design.
type Round struct {
	Name       string `json:"name"`
	Summary    string `json:"summary"`
	NeedsGeoff bool   `json:"needs_geoff"`
	// NoNeedReason says why the first read found the round needs nothing
	// from Geoff. It is required when NeedsGeoff is false.
	NoNeedReason string   `json:"no_need_reason,omitempty"`
	After        []string `json:"after,omitempty"`
	// PlannedMinutes is the round's estimate while it has no passes.
	PlannedMinutes int  `json:"planned_minutes,omitempty"`
	DesignWritten  bool `json:"design_written,omitempty"`
	// Session is the design writer's session.
	Session *Session `json:"session,omitempty"`
}

// Pass is one unit of work in a round.
type Pass struct {
	Name           string   `json:"name"`
	Summary        string   `json:"summary"`
	Round          string   `json:"round"`
	PlannedMinutes int      `json:"planned_minutes"`
	After          []string `json:"after,omitempty"`
	Questions      []string `json:"questions,omitempty"`
	// Model and Effort override the launch's default of opus at high.
	Model      string         `json:"model,omitempty"`
	Effort     string         `json:"effort,omitempty"`
	State      PassState      `json:"state"`
	Started    *time.Time     `json:"started,omitempty"`
	Finished   *time.Time     `json:"finished,omitempty"`
	Stops      []StopInterval `json:"stops,omitempty"`
	StopReason StopReason     `json:"stop_reason,omitempty"`
	// StopDetail is the cause of a stop that has no session record to hold
	// it, such as a failed first launch. It is set only while stopped.
	StopDetail    string `json:"stop_detail,omitempty"`
	ActualMinutes *int   `json:"actual_minutes,omitempty"`
	// Wait is set while the pass is waiting, and only then.
	Wait *Wait `json:"wait,omitempty"`
	// Guess records a pass started on an unanswered question. A guess
	// start that waits carries it too.
	Guess       *Guess       `json:"guess,omitempty"`
	Session     *Session     `json:"session,omitempty"`
	PullRequest *PullRequest `json:"pull_request,omitempty"`
}

// StopInterval is one stretch a pass spent stopped. End is nil while the
// pass is still stopped.
type StopInterval struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end,omitempty"`
}

// The two limits a free pass can wait on.
const (
	LimitTwoPass = "two_pass_limit"
	LimitUsage   = "usage"
)

// Wait records every limit that holds a waiting pass. Since is the time
// the start was requested; settling takes the oldest request first and
// never resets it.
type Wait struct {
	Since  time.Time `json:"since"`
	Limits []string  `json:"limits"`
}

// Guess records the question a pass started on, the recommendation taken,
// and when.
type Guess struct {
	Question       string    `json:"question"`
	Recommendation string    `json:"recommendation"`
	At             time.Time `json:"at"`
}

// Question is something a pass rests on that Geoff decides.
type Question struct {
	ID          string   `json:"id"`
	Text        string   `json:"text"`
	Options     []string `json:"options"`
	Recommended string   `json:"recommended"`
	Answer      *Answer  `json:"answer,omitempty"`
}

// Answer is a ruling reference or a carried-over answer, never both.
type Answer struct {
	Ruling  string   `json:"ruling,omitempty"`
	Carried *Carried `json:"carried,omitempty"`
}

// Carried is an answer Geoff gave before cc-desk, with his words.
type Carried struct {
	Words  string `json:"words"`
	Date   string `json:"date"`
	Source string `json:"source"`
}

// Ruling is one line of rulings.jsonl.
type Ruling struct {
	Format   int       `json:"format"`
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Desk     string    `json:"desk"`
	Words    string    `json:"words"`
	Project  string    `json:"project,omitempty"`
	Round    string    `json:"round,omitempty"`
	Question string    `json:"question,omitempty"`
	Option   string    `json:"option,omitempty"`
}

// Forecast is one line of forecasts.jsonl: when the next decision session
// was forecast to start, made at a pass start, finish, or stop. Project,
// Round, and Time are set together or all absent; Time encodes as null
// when absent.
type Forecast struct {
	Format  int        `json:"format"`
	Made    time.Time  `json:"made"`
	Project string     `json:"project,omitempty"`
	Round   string     `json:"round,omitempty"`
	Time    *time.Time `json:"time"`
}

// CheckFile is check.json: what the last session check decided and how many
// reads of the session list failed in a row. An absent file reads as zero.
type CheckFile struct {
	Format       int        `json:"format"`
	ListFailures int        `json:"list_failures"`
	LastRun      *time.Time `json:"last_run,omitempty"`
	Decisions    []Decision `json:"decisions"`
}

// Decision is what the check decided for one session record. Record names
// the record as pass:<name>, round:<name>, or conductor. Counted is set
// only on a relaunch. Outcome is empty until the check carries the decision
// out and writes what happened.
type Decision struct {
	Project  string `json:"project"`
	Record   string `json:"record"`
	Session  string `json:"session"`
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Counted  *bool  `json:"counted,omitempty"`
	Outcome  string `json:"outcome,omitempty"`
}

// Desks is desks.json: the desk sessions that may record a ruling.
type Desks struct {
	Format int    `json:"format"`
	Desks  []Desk `json:"desks,omitempty"`
}

// Desk is one desk session and the process that runs it.
type Desk struct {
	SessionID string    `json:"session_id"`
	PID       int       `json:"pid"`
	Started   time.Time `json:"started"`
}

// Proc is one process as the guard sees it. TTY is tty_nr from
// /proc/<pid>/stat, which is 0 for a process with no controlling terminal.
type Proc struct {
	PID     int
	Started time.Time
	TTY     int
}

// Caller is who runs a verb: the session id from the environment and the
// calling process's chain of parents. Chain[0] is the calling process.
type Caller struct {
	SessionID string
	Chain     []Proc
}

// Usage is usage.json: the last usage reading. A reading that could not
// be parsed carries Error, with no percent and no reset text.
type Usage struct {
	Format    int       `json:"format"`
	Error     string    `json:"error,omitempty"`
	Percent   int       `json:"percent,omitempty"`
	ResetText string    `json:"reset_text,omitempty"`
	ReadAt    time.Time `json:"read_at"`
}

// Settings is settings.json, edited by hand.
type Settings struct {
	Format           int `json:"format"`
	UsageStopPercent int `json:"usage_stop_percent"`
}

// GuessMinutesCap is the most planned work that may rest on open guesses.
const GuessMinutesCap = 300

// DefaultUsageStop is the usage stop when settings.json is absent.
const DefaultUsageStop = 100

// ReleaseFile is a release with the file it lives in.
type ReleaseFile struct {
	// Path is relative to the state folder.
	Path string
	// Done marks a release under releases/done.
	Done bool
	// Dirty marks a release an event changed, so commit writes it.
	Dirty bool
	Release
}

// State is every state file, loaded.
type State struct {
	// Releases holds the live releases by sorted name, then the done ones.
	Releases  []*ReleaseFile
	Desks     Desks
	Settings  Settings
	Usage     *Usage
	Rulings   []Ruling
	Forecasts []Forecast
	Check     CheckFile

	// passEvent is set by a pass start, finish, or stop, and tells Commit
	// to append a forecast line.
	passEvent bool
	// usageDirty is set when a fresh reading replaced usage.json's.
	usageDirty bool
	// newRuling is the ruling an event recorded, which Commit appends to
	// rulings.jsonl before it writes any release file.
	newRuling *Ruling
	// desksDirty is set when an event changed desks.json.
	desksDirty bool
	// checkDirty is set when an event changed check.json.
	checkDirty bool
	// noSettle tells Commit to skip settle, so a failed list read starts no
	// waiting pass.
	noSettle bool
}
