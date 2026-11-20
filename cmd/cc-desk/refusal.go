package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/cc-desk/internal/atomicfile"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
)

func newRefusalCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "refusal",
		Short: "Handle the calls auto mode refuses in a session",
	})
	cmd.AddCommand(newRefusalHookCmd(f))
	return cmd
}

func newRefusalHookCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Record a PermissionDenied hook's refusal on its session record",
		Long: "The settings file of every launched session names this verb as its PermissionDenied\n" +
			"hook. It reads the hook's input, records the refusal on the session's record, prints\n" +
			"nothing, and exits 0, so the session carries on. A failure is one line in\n" +
			"refusal-hook.log under the state folder.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			recordRefusalHook(f, cmd.InOrStdin())
			return nil
		},
	}
}

// refusalInput is the part of a PermissionDenied hook's input the desk reads.
// Fields it does not name are ignored.
type refusalInput struct {
	SessionID string                     `json:"session_id"`
	ToolName  string                     `json:"tool_name"`
	ToolInput map[string]json.RawMessage `json:"tool_input"`
	Reason    string                     `json:"reason"`
}

// refusalTargets names the tool_input key that holds what each tool aimed at.
var refusalTargets = map[string]string{
	"Bash": "command", "Read": "file_path", "Edit": "file_path", "Write": "file_path",
	"WebFetch": "url", "Grep": "pattern", "Glob": "pattern",
}

// parseRefusal decodes a hook's input into the session id and the refusal. It
// refuses input that is not JSON or has no session_id or tool_name. A tool
// with no named key gets its tool_input as compact JSON, and input with no
// reason gets an empty one.
func parseRefusal(data []byte, now time.Time) (string, state.Refusal, error) {
	var in refusalInput
	if err := json.Unmarshal(data, &in); err != nil {
		return "", state.Refusal{}, fmt.Errorf("read the hook input: %w", err)
	}
	if in.SessionID == "" {
		return "", state.Refusal{}, errors.New("the hook input has no session_id")
	}
	if in.ToolName == "" {
		return in.SessionID, state.Refusal{}, errors.New("the hook input has no tool_name")
	}
	r := state.Refusal{Time: now, Tool: in.ToolName, Reason: in.Reason}
	if key, ok := refusalTargets[in.ToolName]; ok {
		_ = json.Unmarshal(in.ToolInput[key], &r.Target)
	}
	if r.Target == "" && in.ToolInput != nil {
		b, err := json.Marshal(in.ToolInput)
		if err != nil {
			return in.SessionID, state.Refusal{}, fmt.Errorf("read the tool input: %w", err)
		}
		r.Target = string(b)
	}
	return in.SessionID, r, nil
}

// recordRefusalHook records one hook input. It never fails the session: a
// failure is one line in refusal-hook.log, and input that matched no open
// record is kept under hook-unmatched.
func recordRefusalHook(f *rootFlags, in io.Reader) {
	dir, err := f.dir()
	if err != nil {
		return
	}
	now, err := f.clock()
	if err != nil {
		now = time.Now()
	}
	sessionID, err := keepRefusal(dir, now, in)
	if err != nil {
		logRefusalFailure(dir, now, sessionID, err)
	}
}

func keepRefusal(dir string, now time.Time, in io.Reader) (string, error) {
	data, err := io.ReadAll(in)
	if err != nil {
		return "", fmt.Errorf("read the hook input: %w", err)
	}
	sessionID, r, err := parseRefusal(data, now)
	if err != nil {
		return sessionID, err
	}
	var found bool
	if _, err := state.Commit(dir, now, state.LockWait, state.RecordRefusal(sessionID, r, &found), snapshot.Render); err != nil {
		return sessionID, err
	}
	if found {
		return sessionID, nil
	}
	folder := filepath.Join(dir, "hook-unmatched")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return sessionID, fmt.Errorf("create %s: %w", folder, err)
	}
	name := fmt.Sprintf("%s-refusal-%d.json", now.UTC().Format("20060102T150405Z"), os.Getpid())
	if err := atomicfile.WriteFile(filepath.Join(folder, name), data, 0o644); err != nil {
		return sessionID, fmt.Errorf("keep the unmatched input: %w", err)
	}
	return sessionID, fmt.Errorf("no open session record has session_id %q", sessionID)
}

// logRefusalFailure appends one line to refusal-hook.log: the time, the
// session id, and the error. Its own failure is dropped, since the hook has
// nowhere left to report it.
func logRefusalFailure(dir string, now time.Time, sessionID string, err error) {
	if sessionID == "" {
		sessionID = "-"
	}
	msg := strings.Join(strings.Fields(err.Error()), " ")
	line := now.UTC().Format(time.RFC3339) + " " + sessionID + " " + msg + "\n"
	lf, openErr := os.OpenFile(filepath.Join(dir, "refusal-hook.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return
	}
	_, _ = lf.WriteString(line)
	_ = lf.Close()
}
