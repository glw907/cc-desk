package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/glw907/cc-desk/internal/state"
	"github.com/spf13/cobra"
)

const followUpPrefix = "cannot file the follow-up: "

// followUpSlugMax is the longest slug, in bytes.
const followUpSlugMax = 60

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugOf is the title in lower case, with each run of characters other than
// a-z and 0-9 made one "-", trimmed of "-", and cut to 60 bytes.
func slugOf(title string) string {
	s := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if len(s) > followUpSlugMax {
		s = strings.Trim(s[:followUpSlugMax], "-")
	}
	return s
}

// followUpBytes is a follow-up file: the title as a heading, a blank line,
// the body, and one final newline.
func followUpBytes(title, body string) []byte {
	return []byte("# " + title + "\n\n" + strings.TrimRight(body, "\n") + "\n")
}

// writeNew makes the file's folder and creates the file with O_EXCL, so a
// file already there is never overwritten. The error wraps fs.ErrExist when
// the file exists.
func writeNew(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("make %s: %w", filepath.Dir(path), err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}

// within reports whether the resolved path lies inside the resolved folder.
func within(folder, path string) bool {
	rel, err := filepath.Rel(folder, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// readBody reads the body file, which must resolve (past symlinks) to a path
// inside the worktree or the caller's job folder.
func readBody(bodyFile, worktree string) ([]byte, error) {
	roots := []string{worktree}
	if job := os.Getenv("CLAUDE_JOB_DIR"); job != "" {
		roots = append(roots, job)
	}
	resolved, err := filepath.EvalSymlinks(bodyFile)
	if err != nil {
		return nil, state.Refuse(followUpPrefix + fmt.Sprintf("cannot read the body file: %v", err))
	}
	ok := false
	for _, r := range roots {
		if root, err := filepath.EvalSymlinks(r); err == nil && within(root, resolved) {
			ok = true
		}
	}
	if !ok {
		return nil, state.Refuse(followUpPrefix + fmt.Sprintf("the body file %s is outside the worktree and the job folder", resolved))
	}
	b, err := os.ReadFile(resolved)
	if err != nil {
		return nil, state.Refuse(followUpPrefix + fmt.Sprintf("cannot read the body file: %v", err))
	}
	return b, nil
}

func newFollowupCmd(f *rootFlags) *cobra.Command {
	var bodyFile string
	cmd := &cobra.Command{
		Use:   "followup <project> <record> <title> --body-file <path>",
		Short: "File a follow-up in the record's worktree",
		Long: "Write docs/follow-ups/<slug>.md in the record's worktree: the title as a heading, a\n" +
			"blank line, and the body. The body file must lie inside the worktree or the job\n" +
			"folder. A file with the same name is refused. The record is conductor, round:<name>,\n" +
			"or pass:<name>.",
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, record, title := args[0], args[1], args[2]
			slug := slugOf(title)
			if slug == "" {
				return state.Refuse(followUpPrefix + "the title has no letters or digits")
			}
			if bodyFile == "" {
				return state.Refuse(followUpPrefix + "--body-file is required")
			}
			dir, err := f.dir()
			if err != nil {
				return err
			}
			s, err := state.Load(dir)
			if err != nil {
				return err
			}
			sess := recordOf(s, project, record)
			if sess == nil || sess.Ended != nil {
				return state.Refuse(followUpPrefix + fmt.Sprintf("%s/%s has no open session record", project, record))
			}
			body, err := readBody(bodyFile, sess.Worktree)
			if err != nil {
				return err
			}
			if strings.TrimSpace(string(body)) == "" {
				return state.Refuse(followUpPrefix + "the body is empty")
			}
			path := filepath.Join(sess.Worktree, "docs", "follow-ups", slug+".md")
			if err := writeNew(path, followUpBytes(title, string(body))); err != nil {
				if errors.Is(err, fs.ErrExist) {
					return state.Refuse(followUpPrefix + fmt.Sprintf("%s already exists", path))
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "the file holding the body")
	return cmd
}
