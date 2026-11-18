package state

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

func validPassState(s PassState) bool {
	switch s {
	case NotStarted, Waiting, Building, Stopped, Done:
		return true
	}
	return false
}

// ValidStopReason reports whether r is one of the five stop reasons.
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
	if r.Repo == "" {
		return errors.New(`missing required field "repo"`)
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
	return findLoop(r)
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
	if p.State == Stopped && !ValidStopReason(p.StopReason) {
		return fmt.Errorf("pass %q is stopped with reason %q, which is not one of the five", p.Name, p.StopReason)
	}
	if p.StopReason != "" && !ValidStopReason(p.StopReason) {
		return fmt.Errorf("pass %q has stop reason %q, which is not one of the five", p.Name, p.StopReason)
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

// checkAnswers verifies every ruling answer in r points at a ruling that
// exists.
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
	return nil
}
