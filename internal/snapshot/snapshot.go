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
const GuessMinutesLimit = state.GuessMinutesCap

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

// NextSession is the next decision session. State is needs_you_now,
// ready_now, scheduled, or nothing_waiting. WaitingOnDesign names the
// rounds with no design whose start has arrived.
type NextSession struct {
	State           *string    `json:"state"`
	Time            *time.Time `json:"time"`
	Range           *Range     `json:"range"`
	Rough           *bool      `json:"rough"`
	QuestionCount   *int       `json:"question_count"`
	Rounds          []string   `json:"rounds"`
	Projects        []string   `json:"projects"`
	NotDesigned     []string   `json:"not_designed"`
	WaitingOnDesign []string   `json:"waiting_on_design"`
	Reasons         []string   `json:"reasons"`
}

// Range is the low and high ends of a forecast time.
type Range struct {
	Low  time.Time `json:"low"`
	High time.Time `json:"high"`
}

// WaitEntry is a round with written questions needed now.
type WaitEntry struct {
	Project   string `json:"project"`
	Round     string `json:"round"`
	Questions int    `json:"questions"`
}

// WaitingPass is a pass waiting on its limits, listed oldest request
// first. ResetText is the reading's reset text when usage is among the
// limits and the reading is readable, and null otherwise.
type WaitingPass struct {
	Project   string    `json:"project"`
	Pass      string    `json:"pass"`
	Limits    []string  `json:"limits"`
	Since     time.Time `json:"since"`
	ResetText *string   `json:"reset_text"`
}

// OpenGuess is a pass carrying a guess whose question has no answer.
type OpenGuess struct {
	Project  string `json:"project"`
	Pass     string `json:"pass"`
	Question string `json:"question"`
	Minutes  int    `json:"minutes"`
}

// Render builds the state's snapshot once and returns it as indented JSON
// with the forecast line for the same next session. The line names the
// round that sets the session's time. Without one, its time is null.
func Render(s *state.State, now time.Time) ([]byte, state.Forecast, error) {
	snap, setter := build(s, now)
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return nil, state.Forecast{}, err
	}
	var fc state.Forecast
	if setter != nil && setter.round != "" {
		t := setter.time.UTC()
		fc.Project, fc.Round, fc.Time = setter.project, setter.round, &t
	}
	return append(data, '\n'), fc, nil
}

// Build fills the fields passes 1 and 2 own and leaves the rest at their
// empty forms. Only live releases appear; done ones feed the pace.
func Build(s *state.State, now time.Time) *Snapshot {
	snap, _ := build(s, now)
	return snap
}

// build is Build plus the item that sets the next session's time.
func build(s *state.State, now time.Time) (*Snapshot, *item) {
	snap := &Snapshot{
		Format:            state.Format,
		GeneratedAt:       now.UTC(),
		Releases:          []Release{},
		Rulings:           []state.Ruling{},
		NextSession:       NextSession{Rounds: []string{}, Projects: []string{}, NotDesigned: []string{}, WaitingOnDesign: []string{}, Reasons: []string{}},
		WaitsForGeoff:     []WaitEntry{},
		Waiting:           []WaitingPass{},
		OpenGuesses:       []OpenGuess{},
		GuessMinutesLimit: GuessMinutesLimit,
	}
	snap.Rulings = append(snap.Rulings, s.Rulings...)
	snap.GuessMinutesTotal = s.OpenGuessMinutes()
	pcs := paces(s)
	var ctxs []*relCtx
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		out := buildRelease(s, rf)
		pc, ok := pcs[rf.Project]
		if !ok {
			pc = roughPace
		}
		v := newReleaseView(rf)
		times := fillRelease(&out, v, pc, now)
		ctxs = append(ctxs, &relCtx{idx: len(snap.Releases), rf: rf, v: v, pc: pc, times: times})
		snap.Releases = append(snap.Releases, out)
		snap.OpenGuesses = append(snap.OpenGuesses, openGuesses(rf)...)
	}
	snap.Waiting = waitingPasses(s)
	return snap, fillSession(snap, ctxs, now, usageUnreadable(s))
}

// usageUnreadable reports whether the kept reading carries an error. A
// missing reading is a stop in the rules but adds no reason here.
func usageUnreadable(s *state.State) bool {
	return s.Usage != nil && s.Usage.Error != ""
}

func openGuesses(rf *state.ReleaseFile) []OpenGuess {
	var out []OpenGuess
	for i := range rf.Passes {
		if p := &rf.Passes[i]; rf.OpenGuess(p) {
			out = append(out, OpenGuess{Project: rf.Project, Pass: p.Name, Question: p.Guess.Question, Minutes: p.PlannedMinutes})
		}
	}
	return out
}

// waitingPasses lists the waiting passes of live releases in the order
// settling starts them: oldest request first, ties by release then pass.
func waitingPasses(s *state.State) []WaitingPass {
	out := []WaitingPass{}
	var reset *string
	if u := s.Usage; u != nil && u.Error == "" && u.ResetText != "" {
		reset = &u.ResetText
	}
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.State != state.Waiting || p.Wait == nil {
				continue
			}
			w := WaitingPass{Project: rf.Project, Pass: p.Name, Limits: append([]string{}, p.Wait.Limits...), Since: p.Wait.Since.UTC()}
			if reset != nil && slices.Contains(w.Limits, state.LimitUsage) {
				w.ResetText = reset
			}
			out = append(out, w)
		}
	}
	slices.SortStableFunc(out, func(a, b WaitingPass) int { return a.Since.Compare(b.Since) })
	return out
}

func buildRelease(s *state.State, rf *state.ReleaseFile) Release {
	r := &rf.Release
	out := Release{Project: r.Project, Name: r.Name, Repo: r.Repo, Rounds: []Round{}}
	questions := map[string]*state.Question{}
	for i := range r.Questions {
		questions[r.Questions[i].ID] = &r.Questions[i]
	}
	for i := range r.Rounds {
		out.Rounds = append(out.Rounds, buildRound(s, rf, &r.Rounds[i], questions))
	}
	return out
}

func buildRound(s *state.State, rf *state.ReleaseFile, rd *state.Round, questions map[string]*state.Question) Round {
	r := &rf.Release
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
		out.Passes = append(out.Passes, buildPass(s, rf, p, questions))
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

func buildPass(s *state.State, rf *state.ReleaseFile, p *state.Pass, questions map[string]*state.Question) Pass {
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
	out.GuessOutcome = guessOutcome(s, p, questions)
	if p.State == state.NotStarted {
		broken := s.Breaks(rf, p, "")
		free := len(broken) == 0
		out.Free = &free
		for _, b := range broken {
			out.WhyNot = append(out.WhyNot, b.Name)
		}
	}
	return out
}

// The two outcomes of a closed guess.
const (
	Held   = "held"
	Missed = "missed"
)

// guessOutcome is held when the ruling that answered the guess's question
// chose the recommendation the pass took, and missed when it chose another
// option. It is nil with no guess, while the question is open, and when the
// answer was carried over: no ruling closed it.
func guessOutcome(s *state.State, p *state.Pass, questions map[string]*state.Question) *string {
	if p.Guess == nil {
		return nil
	}
	q := questions[p.Guess.Question]
	if q == nil || q.Answer == nil || q.Answer.Ruling == "" {
		return nil
	}
	for _, r := range s.Rulings {
		if r.ID == q.Answer.Ruling {
			out := Missed
			if r.Option == p.Guess.Recommendation {
				out = Held
			}
			return &out
		}
	}
	return nil
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
