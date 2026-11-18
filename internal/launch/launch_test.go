package launch

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestArgs(t *testing.T) {
	r := Runner{Claude: "claude"}
	args := r.Args(Request{Name: "p-w", Model: "opus", Effort: "high", Settings: "/s/p-w.json", Prompt: "do it"})
	if got := strings.Join(args[:5], " "); got != "--user --scope -q claude --bg" {
		t.Errorf("first words = %q", got)
	}
	at := slices.Index(args, "--disallowedTools")
	if at < 0 || args[at-1] != "do it" {
		t.Fatalf("the prompt must come right before --disallowedTools: %v", args)
	}
	if !slices.Equal(args[at+1:], Tools) || len(Tools) != 5 {
		t.Errorf("tools after the flag = %v", args[at+1:])
	}
	joined := strings.Join(args, " ")
	for _, want := range []string{"--permission-mode auto", "--settings /s/p-w.json", "--model opus", "--effort high", "--name p-w"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args lack %q: %s", want, joined)
		}
	}
	for _, bad := range []string{"bypassPermissions", "--dangerously-skip-permissions"} {
		if strings.Contains(joined, bad) {
			t.Errorf("args hold %q", bad)
		}
	}
}

func TestWriteSettings(t *testing.T) {
	state := t.TempDir()
	verbs := Verbs("/bin/cc-desk", state, "pass", "p", "w", "r")
	followRule := "Bash(" + FollowUpCommand("/bin/cc-desk", state, "p", "pass:w") + " *)"
	readyRule := "Bash(" + ReadyCommand("/bin/cc-desk", state, "p", "w") + " *)"
	path, err := WriteSettings(state, "/bin/cc-desk", "p-w", []string{"Bash(go test *)"}, verbs, []string{followRule, readyRule})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(state, "launch", "p-w.json") {
		t.Errorf("path = %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s settings
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	if s.Permissions.Allow[0] != "Bash(go test *)" || len(s.Permissions.Allow) != 3+len(SessionReasons) {
		t.Errorf("allow = %v", s.Permissions.Allow)
	}
	if want := "Bash(/bin/cc-desk --state-dir " + state + " pass stop p w --reason usage)"; !slices.Contains(s.Permissions.Allow, want) {
		t.Errorf("allow lacks %q: %v", want, s.Permissions.Allow)
	}
	if !slices.Contains(s.Permissions.Allow, "Bash(/bin/cc-desk --state-dir "+state+" followup p pass:w *)") {
		t.Errorf("allow lacks the follow-up rule: %v", s.Permissions.Allow)
	}
	if !slices.Contains(s.Permissions.Allow, "Bash(/bin/cc-desk --state-dir "+state+" pass ready p w *)") {
		t.Errorf("allow lacks the ready rule: %v", s.Permissions.Allow)
	}
	if !slices.Equal(s.Permissions.Deny, []string{"Bash(gh pr merge *)", "Bash(approval-queue land *)"}) {
		t.Errorf("deny = %v", s.Permissions.Deny)
	}
	h := s.Hooks.PermissionRequest
	if len(h) != 1 || len(h[0].Hooks) != 1 || h[0].Hooks[0].Command != "/bin/cc-desk --state-dir "+state+" prompt hook" ||
		h[0].Hooks[0].Type != "command" || h[0].Hooks[0].Timeout != 30 {
		t.Errorf("hook = %+v", h)
	}
	if strings.Contains(string(data), "matcher") {
		t.Error("the hook has no matcher")
	}
	design, err := WriteSettings(state, "/bin/cc-desk", "p-design", nil, nil, []string{followRule})
	if err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(design); err != nil || strings.Contains(string(b), "pass ready") {
		t.Errorf("a design launch's settings hold a ready rule: %q, %v", b, err)
	}
	if _, err := WriteSettings("rel", "/bin/cc-desk", "x", nil, nil, nil); err == nil {
		t.Error("a relative state folder must be refused")
	}
}

func TestTemplates(t *testing.T) {
	for _, kind := range []string{"pass", "design", "conductor"} {
		p := Prompt{
			Kind: kind, Name: "proj-x", Project: "proj", Round: "rnd", Pass: "x", Checkout: "/c",
			Worktree: "/c/.claude/worktrees/x", Branch: "x-branch", Base: "main",
			Verbs:    Verbs("/b/cc-desk", "/st", kind, "proj", "x", "rnd"),
			FollowUp: FollowUpCommand("/b/cc-desk", "/st", "proj", "pass:x"),
			Ready:    ReadyCommand("/b/cc-desk", "/st", "proj", "x"),
		}
		got, err := Fill(p)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		for _, want := range []string{
			"proj-x", "/c/.claude/worktrees/x", "x-branch", "git push origin x-branch", "`pass-core`",
			"/b/cc-desk --state-dir /st followup proj pass:x <title> --body-file <path>", "job `tmp` folder",
			"references/round-run.md", "pass-core", "/b/cc-desk --state-dir /st", "## Stop", "approval-queue send",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("%s prompt lacks %q:\n%s", kind, want, got)
			}
		}
		if kind == "pass" {
			for _, want := range []string{"/b/cc-desk --state-dir /st pass ready proj x --pr <number>", "Do not merge it"} {
				if !strings.Contains(got, want) {
					t.Errorf("pass prompt lacks %q", want)
				}
			}
		}
		if strings.Contains(got, "relaunch") {
			t.Errorf("%s prompt names a relaunch on a first launch", kind)
		}
	}
	p := Prompt{Kind: "pass", Name: "a-b", Project: "a", Round: "r", Pass: "b", Branch: "b", Relaunch: &RelaunchNote{Count: 2, Cause: "a lost connection"}}
	got, err := Fill(p)
	if err != nil {
		t.Fatal(err)
	}
	want := "This is relaunch 2 of 2, after a lost connection. Carry on from the branch as it stands; the commits on it are yours."
	if !strings.Contains(got, want) {
		t.Errorf("relaunch paragraph missing:\n%s", got)
	}
}

func TestAnswerNotes(t *testing.T) {
	base := Prompt{Kind: "pass", Name: "a-b", Project: "a", Round: "r", Pass: "b", Branch: "b"}
	for _, tc := range []struct {
		name string
		note AnswerNote
		want []string
	}{
		{"yes", AnswerNote{Yes: true, Rule: "Bash(rm x)"}, []string{"Bash(rm x)", "doesn't count toward the two", "Carry on from the branch"}},
		{"no", AnswerNote{Command: "rm -rf x"}, []string{"rm -rf x", "Do not run it again"}},
	} {
		p := base
		p.Answer = &tc.note
		got, err := Fill(p)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		for _, w := range tc.want {
			if !strings.Contains(got, w) {
				t.Errorf("%s answer lacks %q:\n%s", tc.name, w, got)
			}
		}
		if strings.Contains(got, "This is relaunch") {
			t.Errorf("%s answer holds the relaunch wording", tc.name)
		}
	}
	p := base
	p.Answer = &AnswerNote{Yes: true}
	p.Relaunch = &RelaunchNote{Count: 1}
	if _, err := Fill(p); err == nil {
		t.Error("a prompt with both a relaunch and an answer must be refused")
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git(dir, args...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// repoWithRemote makes a checkout on main with a bare remote named origin.
func repoWithRemote(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	remote, checkout := filepath.Join(root, "remote.git"), filepath.Join(root, "checkout")
	for _, args := range [][]string{{"init", "-q", "--bare", "-b", "main", remote}, {"init", "-q", "-b", "main", checkout}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	runGit(t, checkout, "config", "user.email", "t@example.com")
	runGit(t, checkout, "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(checkout, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, checkout, "add", "f")
	runGit(t, checkout, "commit", "-q", "-m", "first")
	runGit(t, checkout, "remote", "add", "origin", remote)
	runGit(t, checkout, "push", "-q", "origin", "main")
	return checkout
}

func branchExists(t *testing.T, checkout, branch string) bool {
	t.Helper()
	_, err := git(checkout, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch)
	return err == nil
}

func TestWorktreeFreshAddAndUndo(t *testing.T) {
	checkout := repoWithRemote(t)
	w := Worktree{Checkout: checkout, Default: "main", Path: filepath.Join(checkout, ".claude", "worktrees", "w"), Branch: "w"}
	undo, err := w.Make()
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Reuse(); err != nil {
		t.Errorf("a fresh worktree fails Reuse: %v", err)
	}
	if err := undo(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(w.Path); err == nil || branchExists(t, checkout, "w") {
		t.Error("the undo left the folder or the branch")
	}
}

func TestWorktreeUndoKeepsABranchThatMoved(t *testing.T) {
	checkout := repoWithRemote(t)
	w := Worktree{Checkout: checkout, Default: "main", Path: filepath.Join(checkout, "wt"), Branch: "w"}
	undo, err := w.Make()
	if err != nil {
		t.Fatal(err)
	}
	runGit(t, w.Path, "config", "user.email", "t@example.com")
	runGit(t, w.Path, "config", "user.name", "t")
	runGit(t, w.Path, "commit", "-q", "--allow-empty", "-m", "work")
	if err := undo(); err != nil {
		t.Fatal(err)
	}
	if !branchExists(t, checkout, "w") {
		t.Error("the undo deleted a branch that carries a commit")
	}
}

func TestWorktreeRefusals(t *testing.T) {
	checkout := repoWithRemote(t)
	t.Run("an existing folder", func(t *testing.T) {
		path := filepath.Join(checkout, "taken")
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		// A path that is no git checkout proves no git command ran first.
		w := Worktree{Checkout: filepath.Join(t.TempDir(), "not-a-repo"), Default: "main", Path: path, Branch: "b"}
		_, err := w.Make()
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("err = %v, want the existing folder refused before any git command", err)
		}
	})
	t.Run("an existing branch", func(t *testing.T) {
		runGit(t, checkout, "branch", "old")
		w := Worktree{Checkout: checkout, Default: "main", Path: filepath.Join(checkout, "wt-old"), Branch: "old"}
		if _, err := w.Make(); err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Errorf("err = %v", err)
		}
		if !branchExists(t, checkout, "old") {
			t.Error("the refusal deleted the branch")
		}
	})
	t.Run("a failed add", func(t *testing.T) {
		file := filepath.Join(checkout, "blocker")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit(t, checkout, "branch", "bystander")
		w := Worktree{Checkout: checkout, Default: "main", Path: filepath.Join(file, "inside"), Branch: "fresh"}
		if _, err := w.Make(); err == nil {
			t.Fatal("an add under a file must fail")
		}
		if branchExists(t, checkout, "fresh") {
			t.Error("the failed add left the branch it made")
		}
		if !branchExists(t, checkout, "bystander") {
			t.Error("the failed add removed a branch it did not make")
		}
	})
	t.Run("reuse on the wrong branch", func(t *testing.T) {
		w := Worktree{Checkout: checkout, Default: "main", Path: checkout, Branch: "nope"}
		if err := w.Reuse(); err == nil {
			t.Error("Reuse accepted a worktree on another branch")
		}
		if err := (Worktree{Path: filepath.Join(checkout, "gone")}).Reuse(); err == nil {
			t.Error("Reuse accepted a missing folder")
		}
	})
}

func TestParseShortID(t *testing.T) {
	for in, want := range map[string]string{
		"backgrounded · 0c026eb1 · ccd-spike-2-idle\n": "0c026eb1",
		"backgrounded · e3b28ac7\n":                    "e3b28ac7",
		"something else\n":                             "",
	} {
		if got := ParseShortID(in); got != want {
			t.Errorf("ParseShortID(%q) = %q, want %q", in, got, want)
		}
	}
}

// fakeRun builds a Runner whose claude prints out, lists rows, and logs each
// stop to <dir>/stops.
func fakeRun(t *testing.T, out, rows string) (Runner, string) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, body string, perm os.FileMode) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), perm); err != nil {
			t.Fatal(err)
		}
	}
	write("rows.json", rows, 0o644)
	write("out.txt", out, 0o644)
	write("claude", "#!/bin/sh\nd="+dir+"\ncase \"$1\" in\nagents) cat \"$d/rows.json\";;\n--bg) cat \"$d/out.txt\";;\nstop) echo \"$2\" >> \"$d/stops\";;\nesac\n", 0o755)
	write("systemd-run", "#!/bin/sh\nshift 3\nexec \"$@\"\n", 0o755)
	return Runner{Claude: filepath.Join(dir, "claude"), SystemdRun: filepath.Join(dir, "systemd-run"), JobsDir: filepath.Join(dir, "jobs")}, dir
}

func stops(t *testing.T, dir string) string {
	data, _ := os.ReadFile(filepath.Join(dir, "stops"))
	return strings.TrimSpace(string(data))
}

func TestRunIds(t *testing.T) {
	q := func(dir string) Request { return Request{Name: "p-w", Worktree: dir, Known: []string{"old00000"}} }
	row := `[{"id":"ab12cd34","sessionId":"ab12cd34-0000","name":"p-w","state":"working"}]`

	t.Run("the row gives the session id", func(t *testing.T) {
		r, dir := fakeRun(t, "backgrounded · ab12cd34 · p-w\n", row)
		got, err := r.Run(context.Background(), q(dir))
		if err != nil || got.ShortID != "ab12cd34" || got.SessionID != "ab12cd34-0000" {
			t.Errorf("got %+v, %v", got, err)
		}
	})
	t.Run("a missing row falls back to the job's state.json", func(t *testing.T) {
		r, dir := fakeRun(t, "backgrounded · ab12cd34 · p-w\n", "[]")
		job := filepath.Join(dir, "jobs", "ab12cd34")
		if err := os.MkdirAll(job, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(job, "state.json"), []byte(`{"sessionId":"from-job"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := r.Run(context.Background(), q(dir))
		if err != nil || got.SessionID != "from-job" {
			t.Errorf("got %+v, %v", got, err)
		}
	})
	t.Run("neither source stops the session", func(t *testing.T) {
		r, dir := fakeRun(t, "backgrounded · ab12cd34 · p-w\n", "[]")
		if _, err := r.Run(context.Background(), q(dir)); err == nil {
			t.Fatal("a launch with no session id must fail")
		}
		if got := stops(t, dir); got != "ab12cd34" {
			t.Errorf("stops = %q, want ab12cd34", got)
		}
	})
	t.Run("a row that needs a prompt fails and is stopped", func(t *testing.T) {
		r, dir := fakeRun(t, "backgrounded · ab12cd34 · p-w\n",
			`[{"id":"ab12cd34","sessionId":"x","name":"p-w","state":"working","needs":"send a prompt to start"}]`)
		if _, err := r.Run(context.Background(), q(dir)); err == nil || stops(t, dir) != "ab12cd34" {
			t.Errorf("err = %v, stops = %q", err, stops(t, dir))
		}
	})
	t.Run("no id in the output takes the one new row with the name", func(t *testing.T) {
		rows := `[{"id":"old00000","sessionId":"o","name":"p-w","state":"stopped"},` +
			`{"id":"new11111","sessionId":"new11111-0","name":"p-w","state":"working"}]`
		r, dir := fakeRun(t, "", rows)
		got, err := r.Run(context.Background(), q(dir))
		if err != nil || got.ShortID != "new11111" {
			t.Errorf("got %+v, %v", got, err)
		}
	})
	t.Run("no id and no new row fails", func(t *testing.T) {
		r, dir := fakeRun(t, "", `[{"id":"old00000","sessionId":"o","name":"p-w","state":"stopped"}]`)
		if _, err := r.Run(context.Background(), q(dir)); err == nil {
			t.Error("want an error")
		}
	})
	t.Run("a failed launch reports claude's text", func(t *testing.T) {
		r, dir := fakeRun(t, "", "[]")
		r.Claude = "/bin/false"
		if _, err := r.Run(context.Background(), q(dir)); err == nil {
			t.Error("want an error")
		}
	})
}
