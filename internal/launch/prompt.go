package launch

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
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
	// Question is the command prefix a pass or design session adds a
	// question to its round with: the file follows it. It is empty for a
	// conductor.
	Question string
	// Report is the command prefix a conductor records a report with: the
	// file and an optional --round follow it. It is empty for a pass or
	// design.
	Report string
	// Relaunch is set on a relaunch. Count is the number of the relaunch.
	Relaunch *RelaunchNote
	// Answer is set on the relaunch that follows an answered permission
	// prompt. It takes the place of Relaunch.
	Answer *AnswerNote
	// Fix is set on the relaunch that sends a pull request back to its
	// session. It takes the place of Relaunch and Answer.
	Fix *FixNote
	// Replace is set when the pass's answer was replaced. It follows whichever
	// of the notes above is set, or none.
	Replace *ReplaceNote
}

// ReplaceNote fills the replace paragraph: the question and the words that now
// answer it, as Geoff gave them.
type ReplaceNote struct {
	Question string
	Words    string
}

// FixNote fills the fix paragraph: the pull request, the commit the check
// read, and what failed.
type FixNote struct {
	Number   int
	Commit   string
	Failures []string
}

// List gives the failures as one phrase.
func (f FixNote) List() string { return strings.Join(f.Failures, "; ") }

// AnswerNote fills the answer paragraph. On a yes, Rule is the rule added;
// on a no, Command is the command that was refused.
type AnswerNote struct {
	Yes           bool
	Command, Rule string
}

// RelaunchNote fills the relaunch paragraph. An uncounted relaunch, such as
// the one after a usage stop, gives no number, so Count is read only when
// Uncounted is false.
type RelaunchNote struct {
	Count     int
	Cause     string
	Uncounted bool
}

// PushCommand is the command a session pushes its branch with.
func (p Prompt) PushCommand() string { return "git push origin " + p.Branch }

// Fill renders the prompt for the kind, with the relaunch paragraph on a
// relaunch, the answer paragraph after an answered prompt, or the fix
// paragraph on a pull request sent back. It refuses a prompt that sets more
// than one of those. The replace paragraph composes with any of them.
func Fill(p Prompt) (string, error) {
	files := []string{p.Kind + ".txt"}
	notes := 0
	for _, n := range []struct {
		set  bool
		file string
	}{{p.Relaunch != nil, "relaunch.txt"}, {p.Answer != nil, "answer.txt"}, {p.Fix != nil, "fix.txt"}} {
		if n.set {
			notes++
			files = append(files, n.file)
		}
	}
	if notes > 1 {
		return "", fmt.Errorf("prompt for %s sets more than one of a relaunch, an answer, and a fix", p.Name)
	}
	if p.Replace != nil {
		files = append(files, "replace.txt")
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
