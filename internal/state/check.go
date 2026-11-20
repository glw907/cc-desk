package state

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Efforts are the effort levels a session can launch at.
var Efforts = []string{"low", "medium", "high", "xhigh", "max"}

// ValidModel reports whether m is a model a launch accepts: an alias or a
// full claude-* id.
func ValidModel(m string) bool {
	return slices.Contains([]string{"opus", "sonnet", "fable", "haiku"}, m) || strings.HasPrefix(m, "claude-")
}

func validPassState(s PassState) bool {
	switch s {
	case NotStarted, Waiting, Building, Stopped, Done:
		return true
	}
	return false
}

// ValidStopReason reports whether r is a stop reason.
func ValidStopReason(r StopReason) bool { return slices.Contains(StopReasons, r) }

// validateRelease checks one release on its own: required fields, names
// unique and resolvable, answers well formed, pass state consistent with
// its times, and no loop in what rests on what. The error carries no file
// name.
func validateRelease(r *Release) error {
	if r.Project == "" {
		return errors.New(`missing required field "project"`)
	}
	if r.Name == "" {
		return errors.New(`missing required field "name"`)
	}
	if strings.ContainsAny(r.Name, "/\\\x00") {
		return fmt.Errorf("release name %q holds a path separator", r.Name)
	}
	if r.Repo == "" {
		return errors.New(`missing required field "repo"`)
	}
	for i, rule := range r.AskRules {
		if rule == "" {
			return fmt.Errorf(`ask_rules %d: the rule is empty`, i+1)
		}
		if slices.Contains(r.AllowRules, rule) {
			return fmt.Errorf("rule %q is in both allow_rules and ask_rules", rule)
		}
	}

	rounds := map[string]*Round{}
	for i := range r.Rounds {
		rd := &r.Rounds[i]
		if rd.Name == "" {
			return fmt.Errorf(`round %d: missing required field "name"`, i+1)
		}
		if rounds[rd.Name] != nil {
			return fmt.Errorf("round %q is defined twice", rd.Name)
		}
		rounds[rd.Name] = rd
		if rd.Summary == "" {
			return fmt.Errorf(`round %q: missing required field "summary"`, rd.Name)
		}
		if err := checkRoundNotes(rd.Context, rd.Calls); err != nil {
			return fmt.Errorf("round %q: %w", rd.Name, err)
		}
		if !rd.NeedsGeoff && rd.NoNeedReason == "" {
			return fmt.Errorf(`round %q needs nothing from Geoff but gives no "no_need_reason"`, rd.Name)
		}
	}
	for _, rd := range r.Rounds {
		for _, a := range rd.After {
			if rounds[a] == nil {
				return fmt.Errorf("round %q comes after round %q, which does not exist", rd.Name, a)
			}
		}
	}

	questions := map[string]bool{}
	for i := range r.Questions {
		q := &r.Questions[i]
		if q.ID == "" {
			return fmt.Errorf(`question %d: missing required field "id"`, i+1)
		}
		if questions[q.ID] {
			return fmt.Errorf("question %q is defined twice", q.ID)
		}
		questions[q.ID] = true
		if q.Text == "" || len(q.Options) == 0 || q.Recommended == "" {
			return fmt.Errorf(`question %q: missing required field "text", "options", or "recommended"`, q.ID)
		}
		if !slices.Contains(q.Options, q.Recommended) {
			return fmt.Errorf("question %q recommends %q, which is not one of its options", q.ID, q.Recommended)
		}
		if err := validateCauses(q); err != nil {
			return err
		}
		if err := validateAnswer(q); err != nil {
			return err
		}
	}

	passes := map[string]*Pass{}
	for i := range r.Passes {
		p := &r.Passes[i]
		if p.Name == "" {
			return fmt.Errorf(`pass %d: missing required field "name"`, i+1)
		}
		if passes[p.Name] != nil {
			return fmt.Errorf("pass %q is defined twice", p.Name)
		}
		passes[p.Name] = p
	}
	for i := range r.Passes {
		if err := validatePass(&r.Passes[i], rounds, passes, questions); err != nil {
			return err
		}
	}
	if err := validateReleaseSessions(r); err != nil {
		return err
	}
	return findLoop(r)
}

// validateCauses holds a question's causes to its options: when it has any,
// it has one for each option and none for anything else.
func validateCauses(q *Question) error {
	if len(q.Causes) == 0 {
		return nil
	}
	for k := range q.Causes {
		if !slices.Contains(q.Options, k) {
			return fmt.Errorf("question %q: causes has %q, which is not one of its options", q.ID, k)
		}
	}
	for _, o := range q.Options {
		if q.Causes[o] == "" {
			return fmt.Errorf("question %q: causes has nothing for option %q", q.ID, o)
		}
	}
	return nil
}

func validateAnswer(q *Question) error {
	a := q.Answer
	if a == nil {
		return nil
	}
	if (a.Ruling == "") == (a.Carried == nil) {
		return fmt.Errorf("question %q: an answer holds a ruling or a carried-over answer, one of the two", q.ID)
	}
	if c := a.Carried; c != nil && (c.Words == "" || c.Date == "" || c.Source == "") {
		return fmt.Errorf(`question %q: a carried-over answer needs "words", "date", and "source"`, q.ID)
	}
	return nil
}

func validatePass(p *Pass, rounds map[string]*Round, passes map[string]*Pass, questions map[string]bool) error {
	if p.Summary == "" {
		return fmt.Errorf(`pass %q: missing required field "summary"`, p.Name)
	}
	if p.Round == "" {
		return fmt.Errorf(`pass %q: missing required field "round"`, p.Name)
	}
	if rounds[p.Round] == nil {
		return fmt.Errorf("pass %q is in round %q, which does not exist", p.Name, p.Round)
	}
	if p.PlannedMinutes <= 0 {
		return fmt.Errorf(`pass %q: missing required field "planned_minutes"`, p.Name)
	}
	for _, a := range p.After {
		if passes[a] == nil {
			return fmt.Errorf("pass %q comes after pass %q, which does not exist", p.Name, a)
		}
	}
	for _, q := range p.Questions {
		if !questions[q] {
			return fmt.Errorf("pass %q rests on question %q, which does not exist", p.Name, q)
		}
	}
	if !validPassState(p.State) {
		return fmt.Errorf("pass %q: state %q is not one of the five states", p.Name, p.State)
	}
	started := p.Started != nil
	switch p.State {
	case NotStarted, Waiting:
		if started || p.Finished != nil || p.ActualMinutes != nil {
			return fmt.Errorf("pass %q is %s but has start, finish, or actual minutes", p.Name, p.State)
		}
	case Building, Stopped:
		if !started || p.Finished != nil {
			return fmt.Errorf("pass %q is %s and needs a start and no finish", p.Name, p.State)
		}
	case Done:
		if !started || p.Finished == nil || p.ActualMinutes == nil {
			return fmt.Errorf("pass %q is done and needs start, finish, and actual minutes", p.Name)
		}
	}
	if err := validateWait(p); err != nil {
		return err
	}
	if g := p.Guess; g != nil {
		if p.State == NotStarted {
			return fmt.Errorf("pass %q has a guess but has not started", p.Name)
		}
		if !slices.Contains(p.Questions, g.Question) {
			return fmt.Errorf("pass %q guesses on question %q, which it does not rest on", p.Name, g.Question)
		}
	}
	for i, st := range p.Stops {
		open := st.End == nil
		if open != (p.State == Stopped && i == len(p.Stops)-1) {
			return fmt.Errorf("pass %q: only the last stop of a stopped pass may be open", p.Name)
		}
		if started && st.Start.Before(*p.Started) {
			return fmt.Errorf("pass %q: a stop starts before the pass did", p.Name)
		}
		if !open && st.End.Before(st.Start) {
			return fmt.Errorf("pass %q: a stop ends before it starts", p.Name)
		}
	}
	if p.State == Stopped && len(p.Stops) == 0 {
		return fmt.Errorf("pass %q is stopped and needs an open stop interval", p.Name)
	}
	if err := validatePassSession(p); err != nil {
		return err
	}
	if err := validateClosed(p); err != nil {
		return err
	}
	if err := validatePullRequest(p, questions); err != nil {
		return err
	}
	if p.Model != "" && !ValidModel(p.Model) {
		return fmt.Errorf("pass %q: model %q is not opus, sonnet, fable, haiku, or a claude-* id", p.Name, p.Model)
	}
	if p.Effort != "" && !slices.Contains(Efforts, p.Effort) {
		return fmt.Errorf("pass %q: effort %q is not one of %v", p.Name, p.Effort, Efforts)
	}
	if p.StopDetail != "" && p.State != Stopped {
		return fmt.Errorf("pass %q is %s but carries a stop_detail", p.Name, p.State)
	}
	if p.State == Stopped && !ValidStopReason(p.StopReason) {
		return fmt.Errorf("pass %q is stopped with reason %q, which is not a stop reason", p.Name, p.StopReason)
	}
	if p.StopReason != "" && !ValidStopReason(p.StopReason) {
		return fmt.Errorf("pass %q has stop reason %q, which is not a stop reason", p.Name, p.StopReason)
	}
	return nil
}

// validateClosed checks a pass's closed mark: a done pass with a ruling and
// a time, and no merged pull request.
func validateClosed(p *Pass) error {
	c := p.Closed
	switch {
	case c == nil:
		return nil
	case p.State != Done:
		return fmt.Errorf("pass %q is %s but is marked closed", p.Name, p.State)
	case c.Ruling == "" || c.Time.IsZero():
		return fmt.Errorf(`pass %q: closed needs a "ruling" and a "time"`, p.Name)
	case p.PullRequest != nil && p.PullRequest.Merged != nil:
		return fmt.Errorf("pass %q is marked closed but its pull_request is merged", p.Name)
	}
	return nil
}

func validatePassSession(p *Pass) error {
	if s := p.Session; s != nil {
		if p.State != Building && p.State != Stopped && p.State != Done {
			return fmt.Errorf("pass %q is %s but carries a session record", p.Name, p.State)
		}
		if err := validateSession(s, KindPass, fmt.Sprintf("pass %q", p.Name)); err != nil {
			return err
		}
		if s.StopReason != "" || s.Stopped != nil {
			return fmt.Errorf("pass %q: a pass's stop stays on the pass, not on its session record", p.Name)
		}
	}
	return nil
}

// validatePullRequest checks a pass's pull request record. A ready time with
// no commit or verdict is allowed: records written before the ready verb
// existed hold one.
func validatePullRequest(p *Pass, questions map[string]bool) error {
	pr := p.PullRequest
	if pr == nil {
		return nil
	}
	switch {
	case pr.Number <= 0:
		return fmt.Errorf(`pass %q: pull_request needs a "number" above 0`, p.Name)
	case pr.Verdict != "" && !validVerdict(pr.Verdict):
		return fmt.Errorf("pass %q: pull_request verdict %q is not accept, fix, or escalate", p.Name, pr.Verdict)
	case pr.Wait != "" && !slices.Contains(PRWaits, pr.Wait):
		return fmt.Errorf("pass %q: pull_request wait %q is not one of %v", p.Name, pr.Wait, PRWaits)
	case pr.Merged != nil && p.State != Done:
		return fmt.Errorf("pass %q is %s but its pull_request is merged", p.Name, p.State)
	case pr.Question != "" && !questions[pr.Question]:
		return fmt.Errorf("pass %q: pull_request asks question %q, which does not exist", p.Name, pr.Question)
	}
	return nil
}

// validateSession checks one session record: its kind matches where it sits,
// its identity fields are set, its counts are in range, and a stop reason is
// one of the ten and comes with its time. place names the record in an error.
func validateSession(s *Session, want SessionKind, place string) error {
	if s.Kind != want {
		return fmt.Errorf("%s: session kind %q, want %q", place, s.Kind, want)
	}
	for _, f := range []struct{ name, val string }{
		{"name", s.Name}, {"short_id", s.ShortID}, {"session_id", s.SessionID}, {"branch", s.Branch},
	} {
		if f.val == "" {
			return fmt.Errorf(`%s: session record is missing "%s"`, place, f.name)
		}
	}
	if !filepath.IsAbs(s.Worktree) {
		return fmt.Errorf(`%s: session "worktree" %q is not an absolute path`, place, s.Worktree)
	}
	if s.Launched.IsZero() {
		return fmt.Errorf(`%s: session record is missing "launched"`, place)
	}
	if s.LaunchTries < 1 || s.LaunchTries > 2 {
		return fmt.Errorf(`%s: "launch_tries" %d is not 1 or 2`, place, s.LaunchTries)
	}
	for i, r := range s.Relaunches {
		if r.Time.IsZero() || r.Cause == "" {
			return fmt.Errorf(`%s: relaunch %d needs "time" and "cause"`, place, i+1)
		}
	}
	for i, p := range s.Prompts {
		if p.Time.IsZero() || p.Tool == "" {
			return fmt.Errorf(`%s: prompt %d needs "time" and "tool"`, place, i+1)
		}
	}
	for i, r := range s.Refusals {
		if r.Time.IsZero() || r.Tool == "" {
			return fmt.Errorf(`%s: refusal %d needs "time" and "tool"`, place, i+1)
		}
	}
	if (s.StopReason == "") != (s.Stopped == nil) {
		return fmt.Errorf(`%s: "stop_reason" and "stopped" are set together`, place)
	}
	if s.StopReason != "" && !ValidStopReason(s.StopReason) {
		return fmt.Errorf("%s: session stop reason %q is not a stop reason", place, s.StopReason)
	}
	return nil
}

func validateReleaseSessions(r *Release) error {
	if (r.Checkout == "") != (r.DefaultBranch == "") {
		return errors.New(`"checkout" and "default_branch" are set together`)
	}
	if r.Checkout != "" && !filepath.IsAbs(r.Checkout) {
		return fmt.Errorf(`"checkout" %q is not an absolute path`, r.Checkout)
	}
	if c := r.Conductor; c != nil {
		if err := validateSession(c, KindConductor, "conductor"); err != nil {
			return err
		}
	}
	for i := range r.Rounds {
		if s := r.Rounds[i].Session; s != nil {
			if err := validateSession(s, KindDesign, fmt.Sprintf("round %q", r.Rounds[i].Name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateWait(p *Pass) error {
	if p.State != Waiting {
		if p.Wait != nil {
			return fmt.Errorf("pass %q is %s but carries a wait", p.Name, p.State)
		}
		return nil
	}
	if p.Wait == nil || p.Wait.Since.IsZero() || len(p.Wait.Limits) == 0 {
		return fmt.Errorf(`pass %q is waiting and needs a "wait" with "since" and "limits"`, p.Name)
	}
	for _, l := range p.Wait.Limits {
		if !knownLimit(l) {
			return fmt.Errorf("pass %q waits on limit %q, which is not two_pass_limit or usage", p.Name, l)
		}
	}
	return nil
}

// findLoop reports a loop in what rests on what. A pass waits on the
// passes before it and on every round before its own round. A round waits
// on its passes and the rounds before it.
func findLoop(r *Release) error {
	deps := map[string][]string{}
	roundOf := map[string]*Round{}
	for i := range r.Rounds {
		rd := &r.Rounds[i]
		roundOf[rd.Name] = rd
		for _, a := range rd.After {
			deps["round "+rd.Name] = append(deps["round "+rd.Name], "round "+a)
		}
	}
	for i := range r.Passes {
		p := &r.Passes[i]
		id := "pass " + p.Name
		for _, a := range p.After {
			deps[id] = append(deps[id], "pass "+a)
		}
		for _, a := range roundOf[p.Round].After {
			deps[id] = append(deps[id], "round "+a)
		}
		deps["round "+p.Round] = append(deps["round "+p.Round], id)
	}

	const (
		visiting = 1
		visited  = 2
	)
	mark := map[string]int{}
	var path []string
	var visit func(n string) []string
	visit = func(n string) []string {
		switch mark[n] {
		case visited:
			return nil
		case visiting:
			i := slices.Index(path, n)
			return append(slices.Clone(path[i:]), n)
		}
		mark[n] = visiting
		path = append(path, n)
		for _, d := range deps[n] {
			if loop := visit(d); loop != nil {
				return loop
			}
		}
		path = path[:len(path)-1]
		mark[n] = visited
		return nil
	}
	var nodes []string
	for n := range deps {
		nodes = append(nodes, n)
	}
	slices.Sort(nodes)
	for _, n := range nodes {
		if loop := visit(n); loop != nil {
			return fmt.Errorf("a loop in what rests on what: %s", strings.Join(loop, " -> "))
		}
	}
	return nil
}

// checkAnswers verifies every ruling answer, and every closed pass's ruling,
// in r points at a ruling that exists.
func checkAnswers(r *Release, rulings []Ruling) error {
	for i := range r.Questions {
		q := &r.Questions[i]
		if q.Answer == nil || q.Answer.Ruling == "" {
			continue
		}
		if !slices.ContainsFunc(rulings, func(g Ruling) bool { return g.ID == q.Answer.Ruling }) {
			return fmt.Errorf("question %q is answered by ruling %q, which is not in rulings.jsonl", q.ID, q.Answer.Ruling)
		}
	}
	for i := range r.Passes {
		p := &r.Passes[i]
		if p.Closed != nil && !slices.ContainsFunc(rulings, func(g Ruling) bool { return g.ID == p.Closed.Ruling }) {
			return fmt.Errorf("pass %q is closed by ruling %q, which is not in rulings.jsonl", p.Name, p.Closed.Ruling)
		}
	}
	return nil
}

// Check runs the checks that span files: each answer points at a ruling
// that exists, and ruling ids are unique. The first failure wins.
func Check(s *State) error {
	seen := map[string]bool{}
	for i, g := range s.Rulings {
		if seen[g.ID] {
			return &FileError{File: "rulings.jsonl", Line: i + 1, Err: fmt.Errorf("ruling id %q appears twice", g.ID)}
		}
		seen[g.ID] = true
	}
	for _, rf := range s.Releases {
		if err := checkAnswers(&rf.Release, s.Rulings); err != nil {
			return &FileError{File: rf.Path, Err: err}
		}
	}
	return checkSessionsUnique(s)
}

// LiveSessions calls f for each session record of each live release, ended
// ones included, with the name of the record's place: pass:<name>,
// round:<name>, or conductor.
func (s *State) LiveSessions(f func(rf *ReleaseFile, record string, sess *Session)) {
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		if c := rf.Conductor; c != nil {
			f(rf, "conductor", c)
		}
		for i := range rf.Rounds {
			if sess := rf.Rounds[i].Session; sess != nil {
				f(rf, "round:"+rf.Rounds[i].Name, sess)
			}
		}
		for i := range rf.Passes {
			if sess := rf.Passes[i].Session; sess != nil {
				f(rf, "pass:"+rf.Passes[i].Name, sess)
			}
		}
	}
}

// checkSessionsUnique refuses two live records that name one worktree or
// one session name, across every live release. The error names both.
func checkSessionsUnique(s *State) error {
	type seen struct{ file, place string }
	worktrees := map[string]seen{}
	names := map[string]seen{}
	var err error
	s.LiveSessions(func(rf *ReleaseFile, record string, sess *Session) {
		if err != nil || sess.Ended != nil {
			return
		}
		here := seen{rf.Path, rf.Project + "/" + record}
		if prev, ok := worktrees[sess.Worktree]; ok {
			err = &FileError{File: rf.Path, Err: fmt.Errorf("%s and %s name one worktree, %s", prev.place, here.place, sess.Worktree)}
			return
		}
		if prev, ok := names[sess.Name]; ok {
			err = &FileError{File: rf.Path, Err: fmt.Errorf("%s and %s name one session, %s", prev.place, here.place, sess.Name)}
			return
		}
		worktrees[sess.Worktree], names[sess.Name] = here, here
	})
	return err
}
