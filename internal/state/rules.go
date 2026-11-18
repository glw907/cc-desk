package state

import (
	"fmt"
	"slices"
	"strings"
)

// The names a start can break, in the order a refusal lists them. They
// are the design's words and appear in a refusal exactly as written.
const (
	BreakUnfinished = "unfinished work before it"
	BreakClosed     = "a pass before it was closed"
	BreakQuestion   = "an unanswered question"
	BreakOnePass    = "the one-pass rule"
	BreakMinutes    = "the 300-minute rule"
)

// UsageUnreadable is the reason the snapshot gives when the usage figure
// can't be read.
const UsageUnreadable = "the usage figure can't be read"

// Broken is one rule a start breaks, with the plain-words detail a
// refusal prints after its name.
type Broken struct {
	Name   string
	Detail string
}

// Before returns the passes before p and the rounds before it that hold no
// passes. A pass is before p when it is in p's after list or in a round
// p's round comes after, and so on through each one's own lists.
func (rf *ReleaseFile) Before(p *Pass) (passes []*Pass, empty []string) {
	seenPass := map[string]bool{p.Name: true}
	seenRound := map[string]bool{}
	var visitPass func(q *Pass)
	var visitRound func(name string)
	visitRound = func(name string) {
		if seenRound[name] {
			return
		}
		seenRound[name] = true
		any := false
		for i := range rf.Passes {
			if q := &rf.Passes[i]; q.Round == name {
				any = true
				visitPass(q)
			}
		}
		if !any {
			empty = append(empty, name)
		}
	}
	visitPass = func(q *Pass) {
		if seenPass[q.Name] {
			return
		}
		seenPass[q.Name] = true
		passes = append(passes, q)
		for _, a := range q.After {
			if r, err := rf.pass(a); err == nil {
				visitPass(r)
			}
		}
		rf.eachRoundAfter(q.Round, visitRound)
	}
	for _, a := range p.After {
		if r, err := rf.pass(a); err == nil {
			visitPass(r)
		}
	}
	rf.eachRoundAfter(p.Round, visitRound)
	return passes, empty
}

func (rf *ReleaseFile) eachRoundAfter(round string, f func(string)) {
	for i := range rf.Rounds {
		if rf.Rounds[i].Name == round {
			for _, a := range rf.Rounds[i].After {
				f(a)
			}
		}
	}
}

func (rf *ReleaseFile) question(id string) *Question {
	for i := range rf.Questions {
		if rf.Questions[i].ID == id {
			return &rf.Questions[i]
		}
	}
	return nil
}

// OpenGuess reports whether p carries a guess whose question has no answer.
func (rf *ReleaseFile) OpenGuess(p *Pass) bool {
	if p.Guess == nil {
		return false
	}
	q := rf.question(p.Guess.Question)
	return q != nil && q.Answer == nil
}

// Unanswered lists the unanswered questions p rests on, leaving out skip.
func (rf *ReleaseFile) Unanswered(p *Pass, skip string) []string {
	var out []string
	for _, id := range p.Questions {
		if id == skip {
			continue
		}
		if q := rf.question(id); q != nil && q.Answer == nil {
			out = append(out, id)
		}
	}
	return out
}

// Breaks lists the rules a start of p breaks, in table order. guess is the
// question a guess start names, or "" for a plain start. Only a guess start
// can break the 300-minute rule.
func (s *State) Breaks(rf *ReleaseFile, p *Pass, guess string) []Broken {
	var out []Broken
	before, empty := rf.Before(p)

	var unfinished []string
	for i := range rf.Rounds {
		if r := &rf.Rounds[i]; r.Name == p.Round && !r.DesignWritten {
			unfinished = append(unfinished, "round "+r.Name+" has no design written")
		}
	}
	var closed []string
	for _, q := range before {
		switch {
		case q.Closed != nil:
			closed = append(closed, "pass "+q.Name+" was closed")
		case q.State != Done:
			unfinished = append(unfinished, "pass "+q.Name+" is not done")
		}
	}
	for _, r := range empty {
		unfinished = append(unfinished, "round "+r+" has no passes")
	}
	if len(unfinished) > 0 {
		out = append(out, Broken{BreakUnfinished, strings.Join(unfinished, ", ")})
	}
	if len(closed) > 0 {
		out = append(out, Broken{BreakClosed, strings.Join(closed, ", ")})
	}

	if ids := rf.Unanswered(p, guess); len(ids) > 0 {
		out = append(out, Broken{BreakQuestion, strings.Join(ids, ", ")})
	}

	var guessed []string
	for _, q := range before {
		if rf.OpenGuess(q) {
			guessed = append(guessed, q.Name)
		}
	}
	if len(guessed) > 0 {
		out = append(out, Broken{BreakOnePass, "pass " + strings.Join(guessed, ", ") + " is still on a guess"})
	}

	if guess != "" {
		if open := s.OpenGuessMinutes(); open+p.PlannedMinutes > GuessMinutesCap {
			out = append(out, Broken{BreakMinutes, fmt.Sprintf("%d minutes already rest on guesses and this pass adds %d, past %d", open, p.PlannedMinutes, GuessMinutesCap)})
		}
	}
	return out
}

// Free reports whether a plain start of p breaks none of the first three
// rules.
func (s *State) Free(rf *ReleaseFile, p *Pass) bool {
	return len(s.Breaks(rf, p, "")) == 0
}

// OpenGuessMinutes sums the planned minutes of every pass in a live
// release that carries an open guess, in any state.
func (s *State) OpenGuessMinutes() int {
	n := 0
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			if rf.OpenGuess(&rf.Passes[i]) {
				n += rf.Passes[i].PlannedMinutes
			}
		}
	}
	return n
}

// Building counts the building passes in live releases.
func (s *State) Building() int {
	n := 0
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			if rf.Passes[i].State == Building {
				n++
			}
		}
	}
	return n
}

// UsageStop reports whether the usage limit holds: the reading is missing,
// carries an error, or is at or over the stop.
func (s *State) UsageStop() bool {
	stop := s.Settings.UsageStopPercent
	if stop == 0 {
		stop = DefaultUsageStop
	}
	u := s.Usage
	return u == nil || u.Error != "" || u.Percent >= stop
}

// Limits lists the limits that hold a free pass now, the two-pass limit
// first.
func (s *State) Limits() []string {
	var out []string
	if s.Building() >= 2 {
		out = append(out, LimitTwoPass)
	}
	if s.UsageStop() {
		out = append(out, LimitUsage)
	}
	return out
}

func knownLimit(l string) bool { return slices.Contains([]string{LimitTwoPass, LimitUsage}, l) }
