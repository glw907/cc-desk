package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/cc-desk/internal/desk"
	"github.com/glw907/cc-desk/internal/launch"
	"github.com/glw907/cc-desk/internal/sessions"
	"github.com/glw907/cc-desk/internal/snapshot"
	"github.com/glw907/cc-desk/internal/state"
)

// The two relaunch causes an answer records.
const (
	causeYes = "permission_yes"
	causeNo  = "permission_no"
)

const answerLockWait = 60 * time.Second

func newAnswerCmd(f *rootFlags, answer string) *cobra.Command {
	var lf launchFlags
	var followUps, rule string
	use, short := "yes <project> <record> --follow-ups <dir> [--rule <rule>]", "Allow a session's blocked command and relaunch it"
	long := "Answer the record's last unanswered permission prompt with a yes, from a desk\n" +
		"session only. The session is ended, the rule is added to the project's allow rules\n" +
		"(the command itself, or --rule), and the session is relaunched, uncounted, with a\n" +
		"fresh settings file. A follow-up file for the prompt is written into --follow-ups.\n"
	if answer == state.AnswerNo {
		use, short = "no <project> <record> --follow-ups <dir>", "Refuse a session's blocked command and relaunch it"
		long = "Answer the record's last unanswered permission prompt with a no, from a desk\n" +
			"session only. The session is ended and relaunched, uncounted, with a prompt that\n" +
			"names the refused command. A follow-up file for the prompt is written into\n" +
			"--follow-ups.\n"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long + "The record is conductor, round:<name>, or pass:<name>.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnswer(cmd, f, &lf, answer, args[0], args[1], followUps, rule)
		},
	}
	cmd.Flags().StringVar(&followUps, "follow-ups", "", "the folder that receives the prompt's follow-up file")
	if answer == state.AnswerYes {
		cmd.Flags().StringVar(&rule, "rule", "", "the rule to add, when the command is not one simple Bash command")
	}
	lf.register(cmd)
	return cmd
}

// runAnswer carries out an answer in the design's order. Every refusal comes
// before any side effect, and the follow-up file is written last, so a failure
// there undoes nothing.
func runAnswer(cmd *cobra.Command, f *rootFlags, lf *launchFlags, answer, project, record, followUps, rule string) error {
	// The caller is read before any lock, so a wait never shifts the chain.
	caller := desk.Caller()
	l, err := newLauncher(f, lf)
	if err != nil {
		return err
	}
	// A check holds this lock through its second-read wait, so an answer
	// waits longer than the state lock's wait.
	unlock, err := state.LockFile(l.dir, state.CheckLockFile, answerLockWait)
	if err != nil {
		return state.Refuse("cannot answer the prompt: a check holds check.lock; run it again after the check: " + err.Error())
	}
	defer unlock()

	s, err := state.Load(l.dir)
	if err != nil {
		return err
	}
	if _, err := s.CheckAnswer(caller, project, record, answer, rule); err != nil {
		return err
	}
	if followUps == "" {
		return state.Refuse("cannot answer the prompt: --follow-ups is required")
	}
	if fi, err := os.Stat(followUps); err != nil || !fi.IsDir() {
		return state.Refuse(fmt.Sprintf("cannot answer the prompt: --follow-ups %s is not an existing folder", followUps))
	}
	sess := recordOf(s, project, record)

	ctx := context.Background()
	rows, err := sessions.List(ctx, l.runner.Claude)
	if err != nil {
		return err
	}
	if row, found := sessions.Find(rows, sess.ShortID); found && row.Live() {
		if err := l.runner.Stop(ctx, sess.ShortID); err != nil {
			return err
		}
		if rows, err = sessions.List(ctx, l.runner.Claude); err != nil {
			return err
		}
		if row, found := sessions.Find(rows, sess.ShortID); found && row.Live() {
			return state.Refuse(fmt.Sprintf("not relaunched: session %s is still live after claude stop", sess.ShortID))
		}
	}

	var done state.Answered
	if _, err := state.Commit(l.dir, l.now, state.LockWait, state.AnswerPrompt(caller, project, record, answer, rule, &done), snapshot.Render); err != nil {
		return err
	}

	note := &launch.AnswerNote{Yes: answer == state.AnswerYes, Command: done.Prompt.Command, Rule: done.Rule}
	if note.Command == "" {
		note.Command = done.Prompt.Tool
	}
	cause := causeNo
	if note.Yes {
		cause = causeYes
	}
	id, launchErr := l.Launch(ctx, project, record, launchRequest{
		Kind: state.LaunchRelaunch, Cause: cause, Counted: false, Replaces: sess.ShortID, Answer: note,
	})
	if launchErr != nil {
		launchErr = l.failedRelaunch(sessions.Result{Project: project, Record: record}, launchErr)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "launched %s\n", id)
	}

	path, err := writePromptFollowUp(followUps, project, record, done, answer)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), path)
	if launchErr != nil {
		return state.Refuse("the prompt was answered, and the relaunch failed: " + launchErr.Error())
	}
	return nil
}

// promptFollowUpName is the follow-up file's name for a prompt: the project,
// the record with its colon made a "-", and the prompt's time.
func promptFollowUpName(project, record string, t time.Time) string {
	return fmt.Sprintf("prompt-%s-%s-%s.md", project, strings.ReplaceAll(record, ":", "-"), t.UTC().Format("20060102T150405Z"))
}

// promptFollowUp is the follow-up file's content for an answered prompt.
func promptFollowUp(project, record string, a state.Answered, answer string) []byte {
	title := a.Prompt.Tool
	if a.Prompt.Command != "" {
		title += ": " + a.Prompt.Command
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Project: %s\nRecord: %s\nSession: %s\nTime: %s\n", project, record, a.Session, a.Prompt.Time.UTC().Format(time.RFC3339))
	if a.Prompt.SuggestedRule != "" {
		fmt.Fprintf(&b, "Suggested rule: %s\n", a.Prompt.SuggestedRule)
	}
	fmt.Fprintf(&b, "Answer: %s\n", answer)
	if a.Rule != "" {
		fmt.Fprintf(&b, "Rule added: %s\n", a.Rule)
	}
	for _, o := range a.Older {
		line := o.Tool
		if o.Command != "" {
			line += ": " + o.Command
		}
		fmt.Fprintf(&b, "Older unanswered prompt: %s (%s)\n", line, o.Time.UTC().Format(time.RFC3339))
	}
	return followUpBytes(title, b.String())
}

// writePromptFollowUp writes the follow-up file into dir and gives its path.
// A file already there with the same bytes counts as written.
func writePromptFollowUp(dir, project, record string, a state.Answered, answer string) (string, error) {
	path := filepath.Join(dir, promptFollowUpName(project, record, a.Prompt.Time))
	data := promptFollowUp(project, record, a, answer)
	err := writeNew(path, data)
	if errors.Is(err, fs.ErrExist) {
		have, rerr := os.ReadFile(path)
		if rerr == nil && bytes.Equal(have, data) {
			return path, nil
		}
		return "", fmt.Errorf("the follow-up file %s already exists with other content", path)
	}
	if err != nil {
		return "", err
	}
	return path, nil
}
