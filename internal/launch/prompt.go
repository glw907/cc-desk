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
	// Relaunch is set on a relaunch. Count is the number of the relaunch.
	Relaunch *RelaunchNote
}

// RelaunchNote fills the relaunch paragraph.
type RelaunchNote struct {
	Count int
	Cause string
}

// PushCommand is the command a session pushes its branch with.
func (p Prompt) PushCommand() string { return "git push origin " + p.Branch }

// Fill renders the prompt for the kind, with the relaunch paragraph on a
// relaunch.
func Fill(p Prompt) (string, error) {
	files := []string{p.Kind + ".txt"}
	if p.Relaunch != nil {
		files = append(files, "relaunch.txt")
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
