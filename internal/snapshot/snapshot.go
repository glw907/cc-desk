// Package snapshot builds snapshot.json: everything the page draws,
// computed fresh from the state files.
//
// Every field is always present. An empty list encodes as [], an unset
// time or number as null, and an object a later pass fills is present
// with null or empty members. Passes 2 to 4 add their builders in their
// own files and call them from Build.
package snapshot

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// GuessMinutesLimit is the cap on work resting on open guesses.
const GuessMinutesLimit = 300

// Snapshot is the content of snapshot.json.
type Snapshot struct {
	Format            int            `json:"format"`
	GeneratedAt       time.Time      `json:"generated_at"`
	Releases          []Release      `json:"releases"`
	Rulings           []state.Ruling `json:"rulings"`
	NextSession       NextSession    `json:"next_session"`
	WaitsForGeoff     []WaitEntry    `json:"waits_for_geoff"`
	Waiting           []WaitingPass  `json:"waiting"`
	OpenGuesses       []OpenGuess    `json:"open_guesses"`
	GuessMinutesTotal int            `json:"guess_minutes_total"`
	GuessMinutesLimit int            `json:"guess_minutes_limit"`
}

// Release is one live release.
type Release struct {
	Project  string   `json:"project"`
	Name     string   `json:"name"`
	Repo     string   `json:"repo"`
	Pace     *float64 `json:"pace"`
	PaceLow  *float64 `json:"pace_low"`
	PaceHigh *float64 `json:"pace_high"`
	Rough    *bool    `json:"rough"`
	Rounds   []Round  `json:"rounds"`
}

// Round is one round of a release. State is not_designed, not_started,
// building, or done.
type Round struct {
	Name           string     `json:"name"`
	Summary        string     `json:"summary"`
	State          string     `json:"state"`
	After          []string   `json:"after"`
	PlannedMinutes int        `json:"planned_minutes"`
	ActualMinutes  *int       `json:"actual_minutes"`
	Started        *time.Time `json:"started"`
	Finished       *time.Time `json:"finished"`
	Needs          Needs      `json:"needs"`
	MinutesBefore  *int       `json:"minutes_before"`
	NeededTime     *time.Time `json:"needed_time"`
	ForecastTime   *time.Time `json:"forecast_time"`
	Passes         []Pass     `json:"passes"`
}

// Needs says what a round needs from Geoff. Kind is questions, nothing,
// or not_designed. Reason is set for nothing.
type Needs struct {
	Kind      string     `json:"kind"`
	Reason    *string    `json:"reason"`
	Questions []Question `json:"questions"`
}

// Pass is one pass of a round.
type Pass struct {
	Name           string     `json:"name"`
	Summary        string     `json:"summary"`
	State          string     `json:"state"`
	PlannedMinutes int        `json:"planned_minutes"`
	ActualMinutes  *int       `json:"actual_minutes"`
	Started        *time.Time `json:"started"`
	Finished       *time.Time `json:"finished"`
	StopReason     *string    `json:"stop_reason"`
	Questions      []Question `json:"questions"`
	MinutesBefore  *int       `json:"minutes_before"`
	NeededTime     *time.Time `json:"needed_time"`
	Free           *bool      `json:"free"`
	WhyNot         []string   `json:"why_not"`
	GuessOutcome   *string    `json:"guess_outcome"`
}

// Question is a question with whether it is answered. AnswerKind is
// ruling or carried when Answered.
type Question struct {
	ID          string   `json:"id"`
	Text        string   `json:"text"`
	Options     []string `json:"options"`
	Recommended string   `json:"recommended"`
	Answered    bool     `json:"answered"`
	AnswerKind  *string  `json:"answer_kind"`
}

// NextSession is the next decision session. Pass 2 fills it.
type NextSession struct {
	State         *string    `json:"state"`
	Time          *time.Time `json:"time"`
	Range         *Range     `json:"range"`
	Rough         *bool      `json:"rough"`
	QuestionCount *int       `json:"question_count"`
	Rounds        []string   `json:"rounds"`
	Projects      []string   `json:"projects"`
	NotDesigned   []string   `json:"not_designed"`
}

// Range is the low and high ends of a forecast time.
type Range struct {
	Low  time.Time `json:"low"`
	High time.Time `json:"high"`
}

// WaitEntry is a round with needed written questions. Pass 2 fills it.
type WaitEntry struct {
	Project   string `json:"project"`
	Round     string `json:"round"`
	Questions int    `json:"questions"`
}

// WaitingPass is a pass waiting on a limit. Pass 3 fills it.
type WaitingPass struct {
	Project   string `json:"project"`
	Pass      string `json:"pass"`
	Limit     string `json:"limit"`
	ResetText string `json:"reset_text"`
}

// OpenGuess is a pass started on a guess. Pass 3 fills it.
type OpenGuess struct {
	Project  string `json:"project"`
	Pass     string `json:"pass"`
	Question string `json:"question"`
	Minutes  int    `json:"minutes"`
}

// Render builds the state's snapshot and encodes it as indented JSON.
func Render(s *state.State, now time.Time) ([]byte, error) {
	data, err := json.MarshalIndent(Build(s, now), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// Build fills the fields pass 1 owns and leaves the rest at their empty
// forms. Only live releases appear; done ones feed the pace.
func Build(s *state.State, now time.Time) *Snapshot {
	snap := &Snapshot{
		Format:            state.Format,
		GeneratedAt:       now.UTC(),
		Releases:          []Release{},
		Rulings:           []state.Ruling{},
		NextSession:       NextSession{Rounds: []string{}, Projects: []string{}, NotDesigned: []string{}},
		WaitsForGeoff:     []WaitEntry{},
		Waiting:           []WaitingPass{},
		OpenGuesses:       []OpenGuess{},
		GuessMinutesLimit: GuessMinutesLimit,
	}
	snap.Rulings = append(snap.Rulings, s.Rulings...)
	for _, rf := range s.Releases {
		if !rf.Done {
			snap.Releases = append(snap.Releases, buildRelease(&rf.Release))
		}
	}
	return snap
}

func buildRelease(r *state.Release) Release {
	out := Release{Project: r.Project, Name: r.Name, Repo: r.Repo, Rounds: []Round{}}
	questions := map[string]*state.Question{}
	for i := range r.Questions {
		questions[r.Questions[i].ID] = &r.Questions[i]
	}
	for i := range r.Rounds {
		out.Rounds = append(out.Rounds, buildRound(r, &r.Rounds[i], questions))
	}
	return out
}

func buildRound(r *state.Release, rd *state.Round, questions map[string]*state.Question) Round {
	out := Round{
		Name:           rd.Name,
		Summary:        rd.Summary,
		After:          append([]string{}, rd.After...),
		PlannedMinutes: rd.PlannedMinutes,
		Passes:         []Pass{},
	}
	var ids []string
	allDone, anyStarted := true, false
	actual := 0
	for i := range r.Passes {
		p := &r.Passes[i]
		if p.Round != rd.Name {
			continue
		}
		out.Passes = append(out.Passes, buildPass(p, questions))
		for _, id := range p.Questions {
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		if p.State != state.Done {
			allDone = false
		} else {
			actual += *p.ActualMinutes
		}
		if p.Started != nil {
			anyStarted = true
			if out.Started == nil || p.Started.Before(*out.Started) {
				out.Started = utc(p.Started)
			}
		}
		if p.Finished != nil && (out.Finished == nil || p.Finished.After(*out.Finished)) {
			out.Finished = utc(p.Finished)
		}
	}
	if len(out.Passes) > 0 {
		out.PlannedMinutes = 0
		for _, p := range out.Passes {
			out.PlannedMinutes += p.PlannedMinutes
		}
	}

	switch {
	case !rd.DesignWritten && len(out.Passes) == 0:
		out.State = "not_designed"
	case allDone && len(out.Passes) > 0:
		out.State = "done"
		out.ActualMinutes = &actual
	case anyStarted:
		out.State = "building"
	default:
		out.State = "not_started"
	}
	if out.State != "done" {
		out.Finished = nil
	}

	out.Needs = Needs{Questions: []Question{}}
	switch {
	case out.State == "not_designed" && rd.NeedsGeoff:
		out.Needs.Kind = "not_designed"
	case out.State == "not_designed" || len(ids) == 0:
		out.Needs.Kind = "nothing"
		if rd.NoNeedReason != "" {
			out.Needs.Reason = &rd.NoNeedReason
		}
	default:
		out.Needs.Kind = "questions"
		for _, id := range ids {
			out.Needs.Questions = append(out.Needs.Questions, buildQuestion(questions[id]))
		}
	}
	return out
}

func buildPass(p *state.Pass, questions map[string]*state.Question) Pass {
	out := Pass{
		Name:           p.Name,
		Summary:        p.Summary,
		State:          string(p.State),
		PlannedMinutes: p.PlannedMinutes,
		ActualMinutes:  p.ActualMinutes,
		Started:        utc(p.Started),
		Finished:       utc(p.Finished),
		Questions:      []Question{},
		WhyNot:         []string{},
	}
	if p.StopReason != "" {
		reason := string(p.StopReason)
		out.StopReason = &reason
	}
	for _, id := range p.Questions {
		out.Questions = append(out.Questions, buildQuestion(questions[id]))
	}
	return out
}

func buildQuestion(q *state.Question) Question {
	out := Question{
		ID:          q.ID,
		Text:        q.Text,
		Options:     append([]string{}, q.Options...),
		Recommended: q.Recommended,
	}
	if q.Answer != nil {
		out.Answered = true
		kind := "carried"
		if q.Answer.Ruling != "" {
			kind = "ruling"
		}
		out.AnswerKind = &kind
	}
	return out
}

func utc(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	u := t.UTC()
	return &u
}
