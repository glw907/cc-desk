package launch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/glw907/cc-desk/internal/sessions"
)

// Tools are the tools a launched session goes without. The flag that lists
// them takes every word after it, so the prompt comes before it.
var Tools = []string{"Artifact", "SendFeedback", "Workflow", "AskUserQuestion", "ReportFindings"}

// NeedsPrompt is what a row's needs reads when the session started with no
// prompt, which makes the launch a failed one.
const NeedsPrompt = "send a prompt to start"

// Request is one launch.
type Request struct {
	Name     string
	Model    string
	Effort   string
	Worktree string
	Settings string
	Prompt   string
	// Known are the short ids of the rows the caller read before the launch.
	// A launch whose output names no id takes the one new row with its name.
	Known []string
}

// Result is the identity a launch gave its session.
type Result struct {
	ShortID   string
	SessionID string
}

// Runner holds the commands a launch runs.
type Runner struct {
	Claude     string
	SystemdRun string
	JobsDir    string
	// RowWait is how long the launch waits for the session's row to show.
	RowWait time.Duration
}

// Args gives the arguments of the systemd-run command, which begin
// "--user --scope -q <claude> --bg".
func (r Runner) Args(q Request) []string {
	args := []string{
		"--user", "--scope", "-q", r.Claude, "--bg",
		"--name", q.Name, "--model", q.Model, "--effort", q.Effort,
		"--permission-mode", "auto", "--settings", q.Settings,
		q.Prompt, "--disallowedTools",
	}
	return append(args, Tools...)
}

var backgrounded = regexp.MustCompile(`backgrounded · ([0-9a-f]+)`)

// ParseShortID reads the short id from the launch's output.
func ParseShortID(out string) string {
	if m := backgrounded.FindStringSubmatch(out); m != nil {
		return m[1]
	}
	return ""
}

// Run starts the session from q.Worktree and gives its ids. When the session
// started but the launch cannot vouch for it, Run stops the session and
// returns the error.
func (r Runner) Run(ctx context.Context, q Request) (Result, error) {
	cmd := exec.CommandContext(ctx, r.SystemdRun, r.Args(q)...)
	cmd.Dir = q.Worktree
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		text := strings.TrimSpace(errOut.String())
		if text == "" {
			text = strings.TrimSpace(out.String())
		}
		return Result{}, fmt.Errorf("claude --bg: %v: %s", err, text)
	}
	id := ParseShortID(out.String())
	row, found, err := r.waitRow(ctx, q, id)
	if id == "" {
		if err != nil || !found {
			return Result{}, errors.New("the launch named no id and the session list holds no one new row named " +
				q.Name + "; a session it started may still be live and need a stop by hand")
		}
		id = row.ID
	}
	fail := func(cause error) (Result, error) {
		return Result{}, errors.Join(cause, r.Stop(ctx, id))
	}
	if found && (row.Needs == NeedsPrompt || sessions.ReadJob(r.JobsDir, id).Needs == NeedsPrompt) {
		return fail(fmt.Errorf("session %s started with no prompt (needs %q)", q.Name, NeedsPrompt))
	}
	sid := row.SessionID
	if !found || sid == "" {
		sid = r.jobSessionID(id)
	}
	if sid == "" {
		return fail(fmt.Errorf("session %s (%s) has no row in the session list and no session id in its job folder", q.Name, id))
	}
	return Result{ShortID: id, SessionID: sid}, nil
}

// waitRow reads the session list until a row for the launch shows or
// RowWait passes. It reads first, so a row already there costs one read.
func (r Runner) waitRow(ctx context.Context, q Request, id string) (sessions.Row, bool, error) {
	deadline := time.Now().Add(r.RowWait)
	step := min(r.RowWait/10, time.Second)
	for {
		rows, err := sessions.List(ctx, r.Claude)
		if err == nil {
			if row, ok := pick(rows, q, id); ok {
				return row, true, nil
			}
		} else if !time.Now().Before(deadline) {
			return sessions.Row{}, false, err
		}
		if !time.Now().Before(deadline) {
			return sessions.Row{}, false, nil
		}
		time.Sleep(step)
	}
}

// pick finds the row of the launch: by id when the output gave one, else the
// one row with the session's name that the caller had not seen.
func pick(rows []sessions.Row, q Request, id string) (sessions.Row, bool) {
	if id != "" {
		return sessions.Find(rows, id)
	}
	var found sessions.Row
	n := 0
	for _, row := range rows {
		if row.ID != "" && row.Name == q.Name && !slices.Contains(q.Known, row.ID) {
			found = row
			n++
		}
	}
	return found, n == 1
}

// jobSessionID reads sessionId from the job's state.json.
func (r Runner) jobSessionID(id string) string {
	if r.JobsDir == "" || id == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(r.JobsDir, id, "state.json"))
	if err != nil {
		return ""
	}
	var raw struct {
		SessionID string `json:"sessionId"`
	}
	if json.Unmarshal(data, &raw) != nil {
		return ""
	}
	return raw.SessionID
}

// Stop runs `claude stop <id>`.
func (r Runner) Stop(ctx context.Context, id string) error {
	out, err := exec.CommandContext(ctx, r.Claude, "stop", id).CombinedOutput()
	if err != nil {
		return fmt.Errorf("claude stop %s: %v: %s; stop it by hand", id, err, strings.TrimSpace(string(out)))
	}
	return nil
}
