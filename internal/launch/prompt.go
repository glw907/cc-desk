package launch

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*.txt
var templates embed.FS

// Prompt is what a template names. Kind is pass, design, or conductor.
type Prompt struct {
	Kind     string
	Name     string
	Project  string
	Round    string
	Pass     string
	Checkout string
	Worktree string
	Branch   string
	Base     string
	Verbs    []string
	// FollowUp is the command prefix that files a follow-up. It is never a
	// report verb.
	FollowUp string
	// Ready is the command prefix a pass session reports its pull request
	// ready with. It is empty for a design or conductor.
	Ready string
	// Relaunch is set on a relaunch. Count is the number of the relaunch.
	Relaunch *RelaunchNote
	// Answer is set on the relaunch that follows an answered permission
	// prompt. It takes the place of Relaunch.
	Answer *AnswerNote
}

// AnswerNote fills the answer paragraph. On a yes, Rule is the rule added;
// on a no, Command is the command that was refused.
type AnswerNote struct {
	Yes           bool
	Command, Rule string
}

// RelaunchNote fills the relaunch paragraph.
type RelaunchNote struct {
	Count int
	Cause string
}

// PushCommand is the command a session pushes its branch with.
func (p Prompt) PushCommand() string { return "git push origin " + p.Branch }

// Fill renders the prompt for the kind, with the relaunch paragraph on a
// relaunch or the answer paragraph after an answered prompt. It refuses a
// prompt that sets both.
func Fill(p Prompt) (string, error) {
	if p.Relaunch != nil && p.Answer != nil {
		return "", fmt.Errorf("prompt for %s sets both a relaunch and an answer", p.Name)
	}
	files := []string{p.Kind + ".txt"}
	if p.Relaunch != nil {
		files = append(files, "relaunch.txt")
	}
	if p.Answer != nil {
		files = append(files, "answer.txt")
	}
	var out bytes.Buffer
	for i, f := range files {
		t, err := template.New(f).Option("missingkey=error").ParseFS(templates, "templates/"+f)
		if err != nil {
			return "", fmt.Errorf("template %s: %w", f, err)
		}
		if i > 0 {
			out.WriteString("\n")
		}
		if err := t.ExecuteTemplate(&out, f, p); err != nil {
			return "", fmt.Errorf("fill %s: %w", f, err)
		}
	}
	return out.String(), nil
}
