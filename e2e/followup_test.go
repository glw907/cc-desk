//go:build unix

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFollowup(t *testing.T) {
	r := newLaunchRig(t)
	r.addRelease(t, true, "a")
	requireRun(t, r.do(t, workedClock, "pass", "start", "proj", "a"))
	requireRun(t, r.run(t, workedClock, "check"))
	worktree := filepath.Join(r.clone, ".claude", "worktrees", "a")

	body := filepath.Join(worktree, "body.md")
	if err := os.WriteFile(body, []byte("The gate is slow.\nIt takes a minute.\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := r.do(t, workedClock, "followup", "proj", "pass:a", "Speed up the gate!", "--body-file", body)
	if got.code != 0 {
		t.Fatalf("followup exit %d, stderr: %s", got.code, got.stderr)
	}
	file := filepath.Join(worktree, "docs", "follow-ups", "speed-up-the-gate.md")
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Speed up the gate!\n\nThe gate is slow.\nIt takes a minute.\n"
	if string(b) != want {
		t.Errorf("file holds %q, want %q", b, want)
	}
	if strings.TrimSpace(got.stdout) != file {
		t.Errorf("stdout %q, want the path %s", got.stdout, file)
	}

	again := r.do(t, workedClock, "followup", "proj", "pass:a", "speed up the gate", "--body-file", body)
	if again.code != 2 || !strings.Contains(again.stderr, file) {
		t.Errorf("a second file of that name: exit %d, stderr %q, want exit 2 naming %s", again.code, again.stderr, file)
	}

	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("not for a PR\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	refused := r.do(t, workedClock, "followup", "proj", "pass:a", "Leak", "--body-file", outside)
	if refused.code != 2 || !strings.Contains(refused.stderr, "outside the worktree") {
		t.Errorf("a body file outside: exit %d, stderr %q", refused.code, refused.stderr)
	}
	if _, err := os.Stat(filepath.Join(worktree, "docs", "follow-ups", "leak.md")); err == nil {
		t.Error("the refused follow-up left a file")
	}

	link := filepath.Join(worktree, "link.md")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	viaLink := r.do(t, workedClock, "followup", "proj", "pass:a", "Via a link", "--body-file", link)
	if viaLink.code != 2 || !strings.Contains(viaLink.stderr, "outside the worktree") {
		t.Errorf("a link to a file outside: exit %d, stderr %q", viaLink.code, viaLink.stderr)
	}

	empty := filepath.Join(worktree, "empty.md")
	if err := os.WriteFile(empty, []byte("\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, args := range map[string][]string{
		"an empty body":        {"proj", "pass:a", "Empty", "--body-file", empty},
		"a title with no slug": {"proj", "pass:a", "!!!", "--body-file", body},
		"no open record":       {"proj", "pass:zzz", "Nothing", "--body-file", body},
		"a missing body file":  {"proj", "pass:a", "Missing", "--body-file", filepath.Join(worktree, "none.md")},
	} {
		if got := r.do(t, workedClock, append([]string{"followup"}, args...)...); got.code != 2 {
			t.Errorf("%s: exit %d, want 2, stderr %q", name, got.code, got.stderr)
		}
	}
}
