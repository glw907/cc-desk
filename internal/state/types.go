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
)

// StopReasons lists the five reasons in the order the design gives them.
var StopReasons = []StopReason{StopUsage, StopIrreversible, StopSecret, StopAuthData, StopScope}

// Release is one live release of a project: its rounds, passes, and the
// questions they rest on. It is the content of releases/<project>.json.
type Release struct {
	Format    int        `json:"format"`
	Project   string     `json:"project"`
	Name      string     `json:"name"`
	Repo      string     `json:"repo"`
	Rounds    []Round    `json:"rounds,omitempty"`
	Passes    []Pass     `json:"passes,omitempty"`
	Questions []Question `json:"questions,omitempty"`
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
}

// Pass is one unit of work in a round.
type Pass struct {
	Name           string         `json:"name"`
	Summary        string         `json:"summary"`
	Round          string         `json:"round"`
	PlannedMinutes int            `json:"planned_minutes"`
	After          []string       `json:"after,omitempty"`
	Questions      []string       `json:"questions,omitempty"`
	State          PassState      `json:"state"`
	Started        *time.Time     `json:"started,omitempty"`
	Finished       *time.Time     `json:"finished,omitempty"`
	Stops          []StopInterval `json:"stops,omitempty"`
	StopReason     StopReason     `json:"stop_reason,omitempty"`
	ActualMinutes  *int           `json:"actual_minutes,omitempty"`
	// Limit says which limit a waiting pass waits on. Pass 3 writes it.
	Limit *Limit `json:"limit,omitempty"`
	// Guess records a pass started on an unanswered question. Pass 3
	// writes it.
	Guess *Guess `json:"guess,omitempty"`
}

// StopInterval is one stretch a pass spent stopped. End is nil while the
// pass is still stopped.
type StopInterval struct {
	Start time.Time  `json:"start"`
	End   *time.Time `json:"end,omitempty"`
}

// Limit names the limit a waiting pass waits on and since when.
type Limit struct {
	Kind  string    `json:"kind"`
	Since time.Time `json:"since"`
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

// Forecast is one line of forecasts.jsonl.
type Forecast struct {
	Format  int       `json:"format"`
	Made    time.Time `json:"made"`
	Project string    `json:"project"`
	Round   string    `json:"round"`
	Time    time.Time `json:"time"`
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

// Usage is usage.json: the last usage reading.
type Usage struct {
	Format    int       `json:"format"`
	Percent   int       `json:"percent"`
	ResetText string    `json:"reset_text"`
	ReadAt    time.Time `json:"read_at"`
}

// Settings is settings.json, edited by hand.
type Settings struct {
	Format           int `json:"format"`
	UsageStopPercent int `json:"usage_stop_percent"`
}

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
}
