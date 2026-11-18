package state

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// The prefixes of the refusals the prompt and rule verbs give.
const (
	answerPrefix = "cannot answer the prompt: "
	rulesPrefix  = "cannot change the allow rules: "
)

// The two answers to a prompt.
const (
	AnswerYes = "yes"
	AnswerNo  = "no"
)

var ruleForm = regexp.MustCompile(`^[A-Za-z0-9_]+(\(.+\))?$`)

// compoundMarks are the strings that make a command more than one simple
// command, or one that reads or writes a target a rule would also have to
// cover.
var compoundMarks = []string{"&", "|", ";", "*", ">", "<", "$(", "`", "\n", "\r"}

// deniedRules are the rules a yes can never take: the settings deny them, and
// a deny rule wins over an allow rule.
var deniedRules = []string{"Bash(gh pr merge", "Bash(approval-queue land"}

// ValidRule reports whether rule is a tool name (letters, digits, and _),
// optionally followed by a non-empty content in parentheses.
func ValidRule(rule string) bool { return ruleForm.MatchString(rule) }

// ExactRule gives the exact rule for a prompt: Bash(<command>), and only for
// a Bash prompt whose command is one simple command. It reports false for
// any other prompt.
func ExactRule(p Prompt) (string, bool) {
	if p.Tool != "Bash" || p.Command == "" {
		return "", false
	}
	for _, m := range compoundMarks {
		if strings.Contains(p.Command, m) {
			return "", false
		}
	}
	return "Bash(" + p.Command + ")", true
}

// RecordPrompt appends a prompt to the one open session record whose
// session_id is sessionID, and sets found to whether it found one. It
// refuses a prompt with no tool or no time, since a record that holds one
// fails every later load. It skips settle, so a hook's commit starts no
// waiting pass and appends no forecast line.
func RecordPrompt(sessionID string, p Prompt, found *bool) Event {
	return func(s *State, now time.Time) error {
		if p.Tool == "" || p.Time.IsZero() {
			return errors.New("a prompt needs a tool and a time")
		}
		s.noSettle = true
		*found = false
		if sessionID == "" {
			return nil
		}
		p.Time = p.Time.UTC()
		s.LiveSessions(func(rf *ReleaseFile, record string, sess *Session) {
			if *found || sess.Ended != nil || sess.SessionID != sessionID {
				return
			}
			sess.Prompts = append(sess.Prompts, p)
			rf.Dirty = true
			*found = true
		})
		return nil
	}
}

// Answered is what an answer to a prompt found and changed.
type Answered struct {
	// Prompt is the prompt answered, and Index its place on the record.
	Prompt Prompt
	Index  int
	// Session is the name of the session that raised it.
	Session string
	// Rule is the rule a yes adds, empty for a no.
	Rule string
	// Older holds the prompts before it that stay unanswered.
	Older []Prompt
}

// stoppedOnPrompt reports whether the record's stop is a permission prompt:
// the pass's for a pass, the session record's for a conductor or design
// writer.
func stoppedOnPrompt(p *Pass, sess *Session) bool {
	if p != nil {
		return p.State == Stopped && p.StopReason == StopPermissionPrompt
	}
	return sess.StopReason == StopPermissionPrompt && sess.Stopped != nil
}

// CheckAnswer runs every refusal of an answer without changing anything. In
// order, the first match wins: the guard, no open session record, no
// unanswered prompt, a record not stopped on a permission prompt, and for a
// yes a rule that is malformed, denied, or missing when the prompt's command
// has no exact rule. The rule is the one given, else the exact rule of the
// prompt's command. A no takes none.
func (s *State) CheckAnswer(c Caller, project, record, answer, rule string) (Answered, error) {
	var a Answered
	if err := s.Guard(c, answerPrefix); err != nil {
		return a, err
	}
	_, slot, p, err := sessionAt(s, project, record)
	if err != nil {
		return a, err
	}
	sess := *slot
	if sess == nil || sess.Ended != nil {
		return a, Refuse(answerPrefix + fmt.Sprintf("%s/%s has no open session record", project, record))
	}
	i := -1
	for j, pr := range slices.Backward(sess.Prompts) {
		if pr.Answer == "" {
			i = j
			break
		}
	}
	if i < 0 {
		return a, Refuse(answerPrefix + fmt.Sprintf("%s/%s has no unanswered prompt", project, record))
	}
	if !stoppedOnPrompt(p, sess) {
		return a, Refuse(answerPrefix + fmt.Sprintf("%s/%s isn't stopped on a permission prompt", project, record))
	}
	a.Prompt, a.Index, a.Session = sess.Prompts[i], i, sess.Name
	for _, o := range sess.Prompts[:i] {
		if o.Answer == "" {
			a.Older = append(a.Older, o)
		}
	}
	switch answer {
	case AnswerNo:
		return a, nil
	case AnswerYes:
	default:
		return a, fmt.Errorf("the answer is yes or no, not %q", answer)
	}
	if rule == "" {
		exact, ok := ExactRule(a.Prompt)
		if !ok {
			text := "this prompt needs --rule: its tool or command is not one simple Bash command"
			if a.Prompt.SuggestedRule != "" {
				text += "; the session suggested " + a.Prompt.SuggestedRule
			}
			return a, Refuse(answerPrefix + text)
		}
		rule = exact
	}
	if !ValidRule(rule) {
		return a, Refuse(answerPrefix + fmt.Sprintf("%q is not a rule: use a tool name, then a command in parentheses", rule))
	}
	for _, d := range deniedRules {
		if strings.HasPrefix(rule, d) {
			return a, Refuse(answerPrefix + fmt.Sprintf("a deny rule wins over %q, so a yes could never take: answer no", rule))
		}
	}
	a.Rule = rule
	return a, nil
}

// AnswerPrompt answers the record's last unanswered prompt, with the checks
// of CheckAnswer run again inside the lock. A yes adds the rule to the
// release's allow rules when it isn't there. A stopped pass is resumed; a
// conductor or design writer needs no resume, since the relaunch drops its
// stop. out receives what was answered.
func AnswerPrompt(c Caller, project, record, answer, rule string, out *Answered) Event {
	return func(s *State, now time.Time) error {
		a, err := s.CheckAnswer(c, project, record, answer, rule)
		if err != nil {
			return err
		}
		rf, slot, p, err := sessionAt(s, project, record)
		if err != nil {
			return err
		}
		(*slot).Prompts[a.Index].Answer = answer
		if a.Rule != "" && !slices.Contains(rf.AllowRules, a.Rule) {
			rf.AllowRules = append(rf.AllowRules, a.Rule)
		}
		rf.Dirty = true
		*out = a
		if p != nil {
			return StartPass(project, p.Name)(s, now)
		}
		return nil
	}
}

// AddRule adds a rule to the live release's allow rules, from a desk session
// only. A rule already there is refused. It takes effect at the next launch.
func AddRule(c Caller, project, rule string) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.ruleTarget(c, project, rule)
		if err != nil {
			return err
		}
		if slices.Contains(rf.AllowRules, rule) {
			return Refuse(rulesPrefix + fmt.Sprintf("project %q already allows %q", project, rule))
		}
		rf.AllowRules = append(rf.AllowRules, rule)
		rf.Dirty = true
		return nil
	}
}

// RemoveRule removes a rule from the live release's allow rules, from a desk
// session only. A rule that isn't there is refused.
func RemoveRule(c Caller, project, rule string) Event {
	return func(s *State, now time.Time) error {
		rf, err := s.ruleTarget(c, project, rule)
		if err != nil {
			return err
		}
		i := slices.Index(rf.AllowRules, rule)
		if i < 0 {
			return Refuse(rulesPrefix + fmt.Sprintf("project %q doesn't allow %q", project, rule))
		}
		rf.AllowRules = slices.Delete(rf.AllowRules, i, i+1)
		rf.Dirty = true
		return nil
	}
}

// ruleTarget runs the guard and the rule's form, and finds the release.
func (s *State) ruleTarget(c Caller, project, rule string) (*ReleaseFile, error) {
	if err := s.Guard(c, rulesPrefix); err != nil {
		return nil, err
	}
	if !ValidRule(rule) {
		return nil, fmt.Errorf("%q is not a rule: use a tool name, then a command in parentheses", rule)
	}
	return s.liveRelease(project)
}
