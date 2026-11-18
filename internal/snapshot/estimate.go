package snapshot

import (
	"slices"
	"time"

	"github.com/glw907/cc-desk/internal/state"
)

// paceWindow is how many of a project's latest finished passes set its pace.
const paceWindow = 10

// paceMinimum is the fewest finished passes that make a pace more than a
// rough guess.
const paceMinimum = 3

// ratio is actual minutes over planned minutes, kept as integers so a
// paced length lands on a whole second with no float error.
type ratio struct{ a, p int }

// seconds is the paced length of planned minutes, rounded half up.
func (r ratio) seconds(planned int) int64 {
	n := int64(planned) * 60 * int64(r.a)
	d := int64(r.p)
	return (2*n + d) / (2 * d)
}

func (r ratio) float() float64 { return float64(r.a) / float64(r.p) }

// less reports whether r is a smaller ratio than o.
func (r ratio) less(o ratio) bool { return r.a*o.p < o.a*r.p }

// pace is one project's rate: the ratio over its last ten finished passes
// and the smallest and largest single-pass ratios among them. A rough pace
// is 1.0 and has no ends.
type pace struct {
	mid, low, high ratio
	rough          bool
}

type finishedPass struct {
	at time.Time
	r  ratio
}

// paces computes every project's pace from its finished passes in live and
// done releases.
func paces(s *state.State) map[string]pace {
	byProject := map[string][]finishedPass{}
	for _, rf := range s.Releases {
		for i := range rf.Passes {
			p := &rf.Passes[i]
			if p.State != state.Done || p.ActualMinutes == nil || p.Finished == nil {
				continue
			}
			byProject[rf.Project] = append(byProject[rf.Project],
				finishedPass{*p.Finished, ratio{*p.ActualMinutes, p.PlannedMinutes}})
		}
	}
	out := map[string]pace{}
	for project, list := range byProject {
		slices.SortStableFunc(list, func(a, b finishedPass) int { return a.at.Compare(b.at) })
		if len(list) > paceWindow {
			list = list[len(list)-paceWindow:]
		}
		out[project] = paceOf(list)
	}
	return out
}

// roughPace is the pace of a project with no finished pass.
var roughPace = pace{mid: ratio{1, 1}, rough: true}

func paceOf(list []finishedPass) pace {
	if len(list) < paceMinimum {
		return roughPace
	}
	pc := pace{low: list[0].r, high: list[0].r}
	for _, f := range list {
		pc.mid.a += f.r.a
		pc.mid.p += f.r.p
		if f.r.less(pc.low) {
			pc.low = f.r
		}
		if pc.high.less(f.r) {
			pc.high = f.r
		}
	}
	return pc
}

// node is one unfinished thing ahead of another: a pass, or a round that
// has no passes yet and counts its round-level minutes.
type node struct {
	planned int
	pass    *state.Pass
}

// timing is the work ahead of something and when it ends at the pace and
// at both ends of its range.
type timing struct {
	minutes        int
	mid, low, high time.Time
	hasEnds        bool
}

// releaseView indexes one release by name.
type releaseView struct {
	rf      *state.ReleaseFile
	passes  map[string]*state.Pass
	rounds  map[string]*state.Round
	byRound map[string][]*state.Pass
}

func newReleaseView(rf *state.ReleaseFile) *releaseView {
	v := &releaseView{
		rf:      rf,
		passes:  map[string]*state.Pass{},
		rounds:  map[string]*state.Round{},
		byRound: map[string][]*state.Pass{},
	}
	for i := range rf.Rounds {
		v.rounds[rf.Rounds[i].Name] = &rf.Rounds[i]
	}
	for i := range rf.Passes {
		p := &rf.Passes[i]
		v.passes[p.Name] = p
		v.byRound[p.Round] = append(v.byRound[p.Round], p)
	}
	return v
}

// walker collects the unfinished work ahead of its roots. Each pass and
// round counts once.
type walker struct {
	v     *releaseView
	seenP map[string]bool
	seenR map[string]bool
	out   []node
}

func (v *releaseView) walk() *walker {
	return &walker{v: v, seenP: map[string]bool{}, seenR: map[string]bool{}}
}

func (w *walker) pass(name string) {
	if w.seenP[name] {
		return
	}
	w.seenP[name] = true
	p := w.v.passes[name]
	if p.State != state.Done {
		w.out = append(w.out, node{p.PlannedMinutes, p})
	}
	w.ahead(p)
}

// ahead walks what comes before p: its after passes and the rounds its
// round comes after.
func (w *walker) ahead(p *state.Pass) {
	for _, a := range p.After {
		w.pass(a)
	}
	for _, r := range w.v.rounds[p.Round].After {
		w.round(r)
	}
}

func (w *walker) round(name string) {
	if w.seenR[name] {
		return
	}
	w.seenR[name] = true
	if ps := w.v.byRound[name]; len(ps) > 0 {
		for _, p := range ps {
			w.pass(p.Name)
		}
		return
	}
	rd := w.v.rounds[name]
	w.out = append(w.out, node{rd.PlannedMinutes, nil})
	for _, a := range rd.After {
		w.round(a)
	}
}

// beforePass is the unfinished work ahead of p.
func (v *releaseView) beforePass(p *state.Pass) []node {
	w := v.walk()
	w.seenP[p.Name] = true
	w.ahead(p)
	return w.out
}

// beforeRound is the unfinished work ahead of a round that hasn't started.
func (v *releaseView) beforeRound(rd *state.Round) []node {
	w := v.walk()
	w.seenR[rd.Name] = true
	for _, a := range rd.After {
		w.round(a)
	}
	return w.out
}

// activeSeconds is the time a pass has spent building: from its start to
// now, less every stop, an open one up to now.
func activeSeconds(p *state.Pass, now time.Time) int64 {
	d := now.Sub(*p.Started)
	for _, st := range p.Stops {
		end := now
		if st.End != nil {
			end = *st.End
		}
		d -= end.Sub(st.Start)
	}
	return max(0, int64(d/time.Second))
}

// seconds is what a node still counts at ratio r. A pass under way counts
// its paced length less the time it has been building, never below zero.
func seconds(n node, r ratio, now time.Time) int64 {
	sec := r.seconds(n.planned)
	if n.pass != nil && (n.pass.State == state.Building || n.pass.State == state.Stopped) {
		sec = max(0, sec-activeSeconds(n.pass, now))
	}
	return sec
}

// timeOf adds up the nodes ahead at the project's pace and, unless the
// pace is rough, at both ends of its range.
func timeOf(nodes []node, pc pace, now time.Time) timing {
	t := timing{mid: now, low: now, high: now, hasEnds: !pc.rough}
	var mid, low, high int64
	for _, n := range nodes {
		t.minutes += n.planned
		mid += seconds(n, pc.mid, now)
		if t.hasEnds {
			low += seconds(n, pc.low, now)
			high += seconds(n, pc.high, now)
		}
	}
	t.mid = now.Add(time.Duration(mid) * time.Second)
	t.low = now.Add(time.Duration(low) * time.Second)
	t.high = now.Add(time.Duration(high) * time.Second)
	return t
}

// fillRelease sets the release's pace fields and the minutes before and
// needed time of each pass and round yet to start. It returns each such
// pass's timing, by name.
func fillRelease(out *Release, v *releaseView, pc pace, now time.Time) map[string]timing {
	pm := pc.mid.float()
	out.Pace = &pm
	rough := pc.rough
	out.Rough = &rough
	if !pc.rough {
		lo, hi := pc.low.float(), pc.high.float()
		out.PaceLow, out.PaceHigh = &lo, &hi
	}

	passTimes := map[string]timing{}
	for ri := range out.Rounds {
		rd := &out.Rounds[ri]
		for pi := range rd.Passes {
			p := &rd.Passes[pi]
			if p.State != string(state.NotStarted) && p.State != string(state.Waiting) {
				continue
			}
			t := timeOf(v.beforePass(v.passes[p.Name]), pc, now)
			passTimes[p.Name] = t
			setTiming(&p.MinutesBefore, &p.NeededTime, t)
		}
		if rd.State == "not_started" || rd.State == "not_designed" {
			t := timeOf(v.beforeRound(v.rounds[rd.Name]), pc, now)
			setTiming(&rd.MinutesBefore, &rd.NeededTime, t)
			if rd.State == "not_designed" && rd.Needs.Kind == "not_designed" {
				rd.ForecastTime = rd.NeededTime
			}
		}
	}
	return passTimes
}

func setTiming(minutes **int, needed **time.Time, t timing) {
	m := t.minutes
	n := t.mid.UTC()
	*minutes, *needed = &m, &n
}

// relCtx is what the next-session step needs of one live release.
type relCtx struct {
	idx   int
	rf    *state.ReleaseFile
	v     *releaseView
	pc    pace
	times map[string]timing
}

const (
	kindQuestion = iota
	kindRound
)

// item is something Geoff's session is waiting to take up: a written,
// unanswered question, or a round with no design whose start is ahead.
type item struct {
	rel, ri        int
	kind           int
	project, round string
	time           time.Time
	low, high      time.Time
	rough          bool
}

func (it *item) name() string { return it.project + "/" + it.round }

func cmpItems(a, b item) int {
	if c := a.time.Compare(b.time); c != 0 {
		return c
	}
	if a.rel != b.rel {
		return a.rel - b.rel
	}
	if a.ri != b.ri {
		return a.ri - b.ri
	}
	return a.kind - b.kind
}

// questionItem times a question by the earliest of the passes resting on
// it that are yet to start. With none, it is needed now. Its round is the
// first round, in round order, that holds a pass resting on it.
func (c *relCtx) questionItem(q *state.Question, now time.Time) item {
	it := item{rel: c.idx, kind: kindQuestion, project: c.rf.Project, rough: c.pc.rough,
		ri: len(c.rf.Rounds), time: now, low: now, high: now}
	var best *timing
	for i := range c.rf.Passes {
		p := &c.rf.Passes[i]
		if !slices.Contains(p.Questions, q.ID) {
			continue
		}
		ri := slices.IndexFunc(c.rf.Rounds, func(r state.Round) bool { return r.Name == p.Round })
		if ri < it.ri {
			it.ri, it.round = ri, p.Round
		}
		if t, ok := c.times[p.Name]; ok && (best == nil || t.mid.Before(best.mid)) {
			best = &t
		}
	}
	if best != nil {
		it.time, it.low, it.high = best.mid, best.low, best.high
	}
	return it
}

// fillSession sets next_session and waits_for_geoff from every live
// release, and returns the item that sets the session's time, or nil.
func fillSession(snap *Snapshot, ctxs []*relCtx, now time.Time, unreadable bool) *item {
	var items, questions, rounds []item
	var onDesign []string
	stopped := false
	for _, c := range ctxs {
		for i := range c.rf.Passes {
			if c.rf.Passes[i].State == state.Stopped {
				stopped = true
			}
		}
		for i := range c.rf.Questions {
			if c.rf.Questions[i].Answer == nil {
				questions = append(questions, c.questionItem(&c.rf.Questions[i], now))
			}
		}
		for ri, rd := range snap.Releases[c.idx].Rounds {
			if rd.State != "not_designed" || rd.Needs.Kind != "not_designed" {
				continue
			}
			t := timeOf(c.v.beforeRound(c.v.rounds[rd.Name]), c.pc, now)
			it := item{rel: c.idx, ri: ri, kind: kindRound, project: c.rf.Project, round: rd.Name,
				rough: c.pc.rough, time: t.mid, low: t.low, high: t.high}
			if t.mid.After(now) {
				rounds = append(rounds, it)
			} else {
				onDesign = append(onDesign, it.name())
			}
		}
	}
	items = append(append(items, questions...), rounds...)
	slices.SortStableFunc(items, cmpItems)

	next := &snap.NextSession
	slices.SortStableFunc(questions, func(a, b item) int {
		if a.rel != b.rel {
			return a.rel - b.rel
		}
		return a.ri - b.ri
	})
	for _, q := range questions {
		if !slices.Contains(next.Projects, q.project) {
			next.Projects = append(next.Projects, q.project)
		}
		if q.round == "" {
			continue
		}
		if !slices.Contains(next.Rounds, q.name()) {
			next.Rounds = append(next.Rounds, q.name())
		}
		if q.time.After(now) {
			continue
		}
		if n := len(snap.WaitsForGeoff); n > 0 && snap.WaitsForGeoff[n-1].Project == q.project && snap.WaitsForGeoff[n-1].Round == q.round {
			snap.WaitsForGeoff[n-1].Questions++
		} else {
			snap.WaitsForGeoff = append(snap.WaitsForGeoff, WaitEntry{Project: q.project, Round: q.round, Questions: 1})
		}
	}
	for _, r := range rounds {
		next.NotDesigned = append(next.NotDesigned, r.name())
	}
	next.WaitingOnDesign = append(next.WaitingOnDesign, onDesign...)

	var setter *item
	if len(items) > 0 {
		setter = &items[0]
	}
	count := len(questions)
	next.QuestionCount = &count
	set := func(s string) { next.State = &s }
	if unreadable {
		next.Reasons = append(next.Reasons, state.UsageUnreadable)
	}
	switch {
	case stopped || unreadable:
		set("needs_you_now")
	case setter != nil && setter.kind == kindQuestion && !setter.time.After(now):
		set("ready_now")
	case setter != nil:
		set("scheduled")
	case len(onDesign) > 0:
		set("ready_now")
	default:
		set("nothing_waiting")
	}
	if setter == nil {
		return nil
	}
	t, rough := setter.time.UTC(), setter.rough
	next.Time, next.Rough = &t, &rough
	if !rough {
		low, high := setter.low.UTC(), setter.high.UTC()
		if *next.State == "ready_now" {
			// A time already reached has no spread, even when a building
			// predecessor is still inside its high-pace length.
			low, high = t, t
		}
		next.Range = &Range{Low: low, High: high}
	}
	return setter
}
