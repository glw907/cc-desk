package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/glw907/cc-desk/internal/atomicfile"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
)

func newPromptCmd(f *rootFlags) *cobra.Command {
	cmd := group(&cobra.Command{
		Use:   "prompt",
		Short: "Handle the permission prompts a session raises",
	})
	cmd.AddCommand(newPromptHookCmd(f), newAnswerCmd(f, state.AnswerYes), newAnswerCmd(f, state.AnswerNo))
	return cmd
}

func newPromptHookCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Record a PermissionRequest hook's prompt on its session record",
		Long: "The settings file of every launched session names this verb as its PermissionRequest\n" +
			"hook. It reads the hook's input, records the prompt on the session's record, prints\n" +
			"nothing, and exits 0, so the session's own prompt stands. A failure is one line on\n" +
			"stderr.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := recordPromptHook(f, cmd.InOrStdin()); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "prompt hook: "+err.Error())
			}
			return nil
		},
	}
}

// hookPrompt is what a PermissionRequest hook's input gives the record.
type hookPrompt struct {
	SessionID string
	Tool      string
	Command   string
	Rule      string
}

// hookInput is the part of a PermissionRequest hook's input the desk reads.
// Fields it does not name are ignored.
type hookInput struct {
	SessionID   string                     `json:"session_id"`
	ToolName    string                     `json:"tool_name"`
	ToolInput   map[string]json.RawMessage `json:"tool_input"`
	Suggestions []struct {
		Type  string `json:"type"`
		Rules []struct {
			ToolName    string `json:"toolName"`
			RuleContent string `json:"ruleContent"`
		} `json:"rules"`
	} `json:"permission_suggestions"`
}

// parseHookPrompt decodes a hook's input. It refuses input that is not JSON
// or has no session_id or tool_name.
func parseHookPrompt(data []byte) (hookPrompt, error) {
	var in hookInput
	if err := json.Unmarshal(data, &in); err != nil {
		return hookPrompt{}, fmt.Errorf("read the hook input: %w", err)
	}
	if in.SessionID == "" {
		return hookPrompt{}, errors.New("the hook input has no session_id")
	}
	if in.ToolName == "" {
		return hookPrompt{}, errors.New("the hook input has no tool_name")
	}
	p := hookPrompt{SessionID: in.SessionID, Tool: in.ToolName}
	for _, key := range []string{"command", "file_path", "url"} {
		var v string
		if json.Unmarshal(in.ToolInput[key], &v) == nil && v != "" {
			p.Command = v
			break
		}
	}
	for _, sg := range in.Suggestions {
		if sg.Type != "addRules" || len(sg.Rules) == 0 || sg.Rules[0].ToolName == "" {
			continue
		}
		r := sg.Rules[0]
		p.Rule = r.ToolName
		if r.RuleContent != "" {
			p.Rule += "(" + r.RuleContent + ")"
		}
		break
	}
	return p, nil
}

// recordPromptHook records one hook input. Input that matched no open
// record is kept under the state folder's hook-unmatched folder.
func recordPromptHook(f *rootFlags, in io.Reader) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("read the hook input: %w", err)
	}
	p, err := parseHookPrompt(data)
	if err != nil {
		return err
	}
	dir, err := f.dir()
	if err != nil {
		return err
	}
	now, err := f.clock()
	if err != nil {
		return err
	}
	var found bool
	ev := state.RecordPrompt(p.SessionID, state.Prompt{Time: now, Tool: p.Tool, Command: p.Command, SuggestedRule: p.Rule}, &found)
	if _, err := state.Commit(dir, now, state.LockWait, ev, snapshot.Render); err != nil {
		return err
	}
	if found {
		return nil
	}
	folder := filepath.Join(dir, "hook-unmatched")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", folder, err)
	}
	name := fmt.Sprintf("%s-%d.json", now.UTC().Format("20060102T150405Z"), os.Getpid())
	if err := atomicfile.WriteFile(filepath.Join(folder, name), data, 0o644); err != nil {
		return fmt.Errorf("keep the unmatched input: %w", err)
	}
	return fmt.Errorf("no open session record has session_id %q", p.SessionID)
}
