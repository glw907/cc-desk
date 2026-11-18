// Package brief prints the questions Geoff's desk session works through,
// in the order of what each one holds up. It reads a loaded state and the
// snapshot built from it, so its forecast is the one the page shows.
package brief

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
)

// ComeBackWindow is how far ahead the "could come back" section looks.
const ComeBackWindow = 6 * time.Hour

// NoDeskWindow is how far back the log section reaches when no desk
// session is recorded.
const NoDeskWindow = 24 * time.Hour

// mergeCauses is what each merge answer sets going, in the order the brief
// prints them. %s is the commit asked about.
var mergeCauses = []struct{ answer, cause string }{
	{"merge", "Merges on the check's next run if every check has passed and the last commit is still %s. Otherwise it asks again."},
	{"fix again", "Sends the pass back with your words for one more attempt. If it comes back the same way, this is asked again."},
	{"close", "Closes the pull request, ends the session, and records the pass as closed by your ruling."},
}

// entry is one waiting question with what rests on it.
type entry struct {
	rel     int
	rf      *state.ReleaseFile
	q       *state.Question
	round   string
	roundAt int
	merge   *state.Pass
	waits   []*state.Pass
	minutes int
}

// Build returns the whole brief at now.
func Build(s *state.State, now time.Time) string {
	snap := snapshot.Build(s, now)
	var b strings.Builder
	fmt.Fprintf(&b, "Brief at %s\n", now.UTC().Format(time.RFC3339))
	if t := snap.NextSession.Time; t != nil {
		fmt.Fprintf(&b, "Next question forecast: %s\n", t.UTC().Format(time.RFC3339))
	} else {
		b.WriteString("Next question forecast: none forecast\n")
	}
	b.WriteString("\n")
	writeWaiting(&b, waiting(s, ""))
	b.WriteString("\nCould come back in the next 6 hours:\n")
	writeComeBack(&b, snap, now)
	b.WriteString("\n")
	writeSince(&b, s, now)
	return b.String()
}

// BuildRound returns the brief for one round: its design, then its
// questions. It names a project or round that does not exist.
func BuildRound(s *state.State, now time.Time, project, round string) (string, error) {
	var rf *state.ReleaseFile
	for _, r := range s.Releases {
		if r.Project == project && !r.Done {
			rf = r
		}
	}
	if rf == nil {
		return "", fmt.Errorf("no project %q", project)
	}
	if !slices.ContainsFunc(rf.Rounds, func(r state.Round) bool { return r.Name == round }) {
		return "", fmt.Errorf("project %q has no round %q", project, round)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Brief for %s / %s at %s\n\n", project, round, now.UTC().Format(time.RFC3339))
	for i := range rf.Rounds {
		if rf.Rounds[i].Name == round {
			writeDesign(&b, &rf.Rounds[i])
		}
	}
	var es []entry
	for _, e := range waiting(s, project) {
		if e.round == round {
			es = append(es, e)
		}
	}
	writeWaiting(&b, es)
	return b.String(), nil
}

// writeDesign prints a round's design path, context, and calls, or says
// none is recorded.
func writeDesign(b *strings.Builder, rd *state.Round) {
	if rd.Design == "" {
		b.WriteString("No design recorded yet.\n\n")
		return
	}
	fmt.Fprintf(b, "Design: %s\n\n", rd.Design)
	if len(rd.Context) > 0 {
		b.WriteString("Context:\n")
		for _, p := range rd.Context {
			fmt.Fprintf(b, "  %s\n", p)
		}
		b.WriteString("\n")
	}
	if len(rd.Calls) > 0 {
		b.WriteString("Calls:\n")
		for _, c := range rd.Calls {
			fmt.Fprintf(b, "  %s: chose %s over %s. %s\n", c.Call, c.Chosen, c.Other, c.Why)
		}
		b.WriteString("\n")
	}
}

// waiting lists the unanswered questions a not-started pass rests on and
// the unanswered merge questions, most planned minutes held up first. A
// non-empty project keeps one project's questions.
func waiting(s *state.State, project string) []entry {
	var es []entry
	for ri, rf := range s.Releases {
		if rf.Done || (project != "" && rf.Project != project) {
			continue
		}
		merges := map[string]*state.Pass{}
		for i := range rf.Passes {
			if p := &rf.Passes[i]; p.PullRequest != nil && p.PullRequest.Question != "" {
				merges[p.PullRequest.Question] = p
			}
		}
		for i := range rf.Questions {
			q := &rf.Questions[i]
			if q.Answer != nil {
				continue
			}
			e := entry{rel: ri, rf: rf, q: q, merge: merges[q.ID], roundAt: len(rf.Rounds)}
			rests := restingOn(rf, q.ID)
			notStarted := slices.ContainsFunc(rests, func(p *state.Pass) bool { return p.State == state.NotStarted })
			if !notStarted && e.merge == nil {
				continue
			}
			for _, p := range rests {
				if at := slices.IndexFunc(rf.Rounds, func(r state.Round) bool { return r.Name == p.Round }); at >= 0 && at < e.roundAt {
					e.roundAt, e.round = at, p.Round
				}
			}
			e.waits = holdsUp(rf, rests)
			for _, p := range e.waits {
				e.minutes += p.PlannedMinutes
			}
			es = append(es, e)
		}
	}
	slices.SortStableFunc(es, func(a, b entry) int {
		switch {
		case a.minutes != b.minutes:
			return b.minutes - a.minutes
		case a.rel != b.rel:
			return a.rel - b.rel
		case a.roundAt != b.roundAt:
			return a.roundAt - b.roundAt
		}
		return strings.Compare(a.q.ID, b.q.ID)
	})
	return es
}

// restingOn lists the passes that are not done and rest on the question.
func restingOn(rf *state.ReleaseFile, id string) []*state.Pass {
	var out []*state.Pass
	for i := range rf.Passes {
		if p := &rf.Passes[i]; p.State != state.Done && slices.Contains(p.Questions, id) {
			out = append(out, p)
		}
	}
	return out
}

// holdsUp returns the resting passes and every pass that is not done and
// comes after one of them, once each, in release order.
func holdsUp(rf *state.ReleaseFile, rests []*state.Pass) []*state.Pass {
	var out []*state.Pass
	for i := range rf.Passes {
		p := &rf.Passes[i]
		if p.State == state.Done {
			continue
		}
		if slices.Contains(rests, p) {
			out = append(out, p)
			continue
		}
		before, _ := rf.Before(p)
		if slices.ContainsFunc(before, func(q *state.Pass) bool { return slices.Contains(rests, q) }) {
			out = append(out, p)
		}
	}
	return out
}

func plural(n int, one string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %ss", n, one)
}

func writeWaiting(b *strings.Builder, es []entry) {
	if len(es) == 0 {
		b.WriteString("Waiting: nothing\n")
		return
	}
	fmt.Fprintf(b, "Waiting: %s\n", plural(len(es), "question"))
	for i, e := range es {
		b.WriteString("\n")
		writeQuestion(b, i+1, e)
	}
}

func writeQuestion(b *strings.Builder, n int, e entry) {
	q, rf := e.q, e.rf
	where := rf.Project
	if e.round != "" {
		where += " / " + e.round
	}
	if p := e.merge; p != nil {
		commit := p.PullRequest.Commit
		fmt.Fprintf(b, "%d. %s, merge question %s on pull request %d, %q, at commit %s, holds up %d planned minutes\n",
			n, where, q.ID, p.PullRequest.Number, p.Summary, commit, e.minutes)
	} else {
		fmt.Fprintf(b, "%d. %s, question %s, holds up %d planned minutes\n", n, where, q.ID, e.minutes)
	}
	fmt.Fprintf(b, "   %s\n", q.Text)
	if len(q.Background) > 0 {
		b.WriteString("   Background:\n")
		for _, para := range q.Background {
			fmt.Fprintf(b, "     %s\n", para)
		}
	}
	b.WriteString("   Answers:\n")
	for _, o := range answerOrder(q, e.merge) {
		cause := q.Causes[o]
		if cause == "" && e.merge != nil {
			for _, m := range mergeCauses {
				if m.answer == o {
					cause = strings.ReplaceAll(m.cause, "%s", e.merge.PullRequest.Commit)
				}
			}
		}
		if cause == "" {
			fmt.Fprintf(b, "     %s\n", o)
		} else {
			fmt.Fprintf(b, "     %s: %s\n", o, cause)
		}
	}
	if q.Recommended != "" {
		if q.Reason != "" {
			fmt.Fprintf(b, "   Recommended: %s. %s\n", q.Recommended, q.Reason)
		} else {
			fmt.Fprintf(b, "   Recommended: %s.\n", q.Recommended)
		}
	}
	if q.NeverGuess {
		b.WriteString("   Never guessed.\n")
	}
	if len(e.waits) > 0 {
		parts := make([]string, len(e.waits))
		for i, p := range e.waits {
			parts[i] = fmt.Sprintf("%s (%s)", p.Name, plural(p.PlannedMinutes, "minute"))
		}
		fmt.Fprintf(b, "   Waits: %s\n", strings.Join(parts, ", "))
	}
}

// answerOrder puts a merge question's three answers in the design's order
// and keeps any other option after them.
func answerOrder(q *state.Question, merge *state.Pass) []string {
	if merge == nil {
		return q.Options
	}
	var out []string
	for _, m := range mergeCauses {
		if slices.Contains(q.Options, m.answer) {
			out = append(out, m.answer)
		}
	}
	for _, o := range q.Options {
		if !slices.Contains(out, o) {
			out = append(out, o)
		}
	}
	return out
}

func writeComeBack(b *strings.Builder, snap *snapshot.Snapshot, now time.Time) {
	found := false
	if t := snap.NextSession.Time; t != nil && !t.After(now.Add(ComeBackWindow)) {
		found = true
		fmt.Fprintf(b, "  The next decision session is forecast at %s", t.UTC().Format(time.RFC3339))
		if len(snap.NextSession.Rounds) > 0 {
			fmt.Fprintf(b, " (%s)", strings.Replace(snap.NextSession.Rounds[0], "/", " / ", 1))
		}
		b.WriteString(".\n")
	}
	for _, g := range snap.OpenGuesses {
		found = true
		fmt.Fprintf(b, "  %s / %s runs on a guess for %s, %s.\n", g.Project, g.Pass, g.Question, plural(g.Minutes, "minute"))
	}
	if !found {
		b.WriteString("  Nothing.\n")
	}
}

func writeSince(b *strings.Builder, s *state.State, now time.Time) {
	var since time.Time
	for _, d := range s.Desks.Desks {
		if d.Started.After(since) {
			since = d.Started
		}
	}
	if since.IsZero() {
		since = now.Add(-NoDeskWindow)
		fmt.Fprintf(b, "No desk session is recorded, so the last 24 hours, since %s:\n", since.UTC().Format(time.RFC3339))
	} else {
		fmt.Fprintf(b, "Since the last desk session, %s:\n", since.UTC().Format(time.RFC3339))
	}
	found := false
	for _, l := range s.CheckLog {
		if !l.Time.After(since) {
			continue
		}
		found = true
		fmt.Fprintf(b, "  %s %s %s %s: %s", l.Time.UTC().Format(time.RFC3339), l.Project, l.Record, l.Decision, l.Outcome)
		if l.Reason != "" {
			fmt.Fprintf(b, " (reason: %s)", l.Reason)
		}
		b.WriteString("\n")
	}
	if !found {
		b.WriteString("  Nothing.\n")
	}
}
