package sessions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	var calls atomic.Int32
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.NotFound(w, nil)
	}))
	defer up.Close()
	var downCalls atomic.Int32
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		downCalls.Add(1)
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		conn.Close()
	}))
	defer down.Close()

	if !Probe(context.Background(), up.URL) {
		t.Error("a 404 answer should read as the link being up")
	}
	if calls.Load() != 1 {
		t.Errorf("up server saw %d calls, want 1", calls.Load())
	}
	if Probe(context.Background(), down.URL) {
		t.Error("a closed connection should read as the link being down")
	}
	if downCalls.Load() != 1 {
		t.Errorf("down server saw %d calls, want 1", downCalls.Load())
	}
	if Probe(context.Background(), "http://127.0.0.1:1/") {
		t.Error("a refused connection should read as down")
	}
	if Probe(context.Background(), "://bad") {
		t.Error("a bad URL should read as down")
	}
}

func TestReadJob(t *testing.T) {
	dir := t.TempDir()
	write := func(id, content string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(dir, id), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, id, "state.json"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("good", `{"detail":"Login expired","needs":"you","other":1}`)
	write("bad", `{not json`)
	write("typed", `{"detail":3,"needs":["x"]}`)
	tests := []struct {
		id   string
		want Job
	}{
		{"good", Job{Detail: "Login expired", Needs: "you"}},
		{"bad", Job{}},
		{"typed", Job{}},
		{"absent", Job{}},
		{"", Job{}},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := ReadJob(dir, tt.id); got != tt.want {
				t.Fatalf("ReadJob = %+v, want %+v", got, tt.want)
			}
		})
	}
	if got := ReadJob("", "good"); got != (Job{}) {
		t.Errorf("no jobs dir: %+v", got)
	}
}

func TestListAndScreenWithAStandIn(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "claude")
	body := "#!/bin/sh\ncase \"$1\" in\nagents) echo '[{\"id\":\"ab12\",\"state\":\"working\"}]';;\nlogs) echo \"screen of $2\";;\nesac\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	rows, err := List(context.Background(), script)
	if err != nil || len(rows) != 1 || rows[0].ID != "ab12" {
		t.Fatalf("List = %+v, %v", rows, err)
	}
	if got := Screen(context.Background(), script, "ab12"); got != "screen of ab12\n" {
		t.Errorf("Screen = %q", got)
	}
	if got := Screen(context.Background(), filepath.Join(dir, "absent"), "ab12"); got != "" {
		t.Errorf("a failed logs read should be empty, got %q", got)
	}
	if _, err := List(context.Background(), filepath.Join(dir, "absent")); err == nil {
		t.Error("a failed list read should be an error")
	}
	if _, err := List(context.Background(), ""); err == nil {
		t.Error("no command should be an error")
	}
}

func TestSignedIn(t *testing.T) {
	if SignInLimit != 30*time.Second {
		t.Errorf("SignInLimit = %s, want 30s", SignInLimit)
	}
	written := time.Date(2026, 11, 20, 18, 0, 0, 0, time.UTC)
	// standIn writes a stand-in claude whose body may name @DIR@, a config
	// folder holding a .credentials.json written at the written time.
	standIn := func(t *testing.T, body string) string {
		t.Helper()
		dir := t.TempDir()
		creds := filepath.Join(dir, ".credentials.json")
		if err := os.WriteFile(creds, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(creds, written, written); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(t.TempDir(), "claude.sh")
		body = strings.ReplaceAll(body, "@DIR@", dir)
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	tests := []struct {
		name string
		body string
		want time.Time
	}{
		{"good", `[ "$*" = "auth status --json" ] || exit 9
echo '{"loggedIn": true, "configDirectory": "@DIR@", "loginExpiresAt": "2099-01-01T00:00:00Z"}'`, written},
		{"good with no expiry named", `echo '{"loggedIn": true, "configDirectory": "@DIR@"}'`, written},
		{"signed out", `echo '{"loggedIn": false, "configDirectory": "@DIR@"}'; exit 1`, time.Time{}},
		{"signed out with exit 0", `echo '{"loggedIn": false, "configDirectory": "@DIR@"}'`, time.Time{}},
		{"the login expired", `echo '{"loggedIn": true, "configDirectory": "@DIR@", "loginExpiresAt": "2020-01-01T00:00:00Z"}'`, time.Time{}},
		{"no config folder named", `echo '{"loggedIn": true}'`, time.Time{}},
		{"no credentials file", `echo '{"loggedIn": true, "configDirectory": "@DIR@/none"}'`, time.Time{}},
		{"good text but a failed exit", `echo '{"loggedIn": true, "configDirectory": "@DIR@"}'; exit 1`, time.Time{}},
		{"not JSON", `echo 'Logged in'`, time.Time{}},
		{"no output", `exit 0`, time.Time{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignedIn(context.Background(), standIn(t, tt.body), time.Second); !got.Equal(tt.want) {
				t.Errorf("SignedIn = %s, want %s", got, tt.want)
			}
		})
	}
	t.Run("no command", func(t *testing.T) {
		if got := SignedIn(context.Background(), "", time.Second); !got.IsZero() {
			t.Errorf("an empty command read as signed in at %s", got)
		}
	})
	t.Run("a child holding the output past the limit", func(t *testing.T) {
		// The sleep is a child of the shell and holds stdout, so only a
		// kill of the process group ends the read at its limit.
		pidFile := filepath.Join(t.TempDir(), "child.pid")
		claude := standIn(t, `sleep 20 & echo $! > `+pidFile+`; wait; echo '{"loggedIn": true, "configDirectory": "@DIR@"}'`)
		start := time.Now()
		got := SignedIn(context.Background(), claude, 200*time.Millisecond)
		if took := time.Since(start); took > 3*time.Second {
			t.Errorf("the read took %s, want it ended near the 200ms limit", took)
		}
		if !got.IsZero() {
			t.Errorf("a read past the limit read as signed in at %s", got)
		}
		// WaitDelay alone would end the read but leave the sleep running.
		b, err := os.ReadFile(pidFile)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
		if err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(3 * time.Second)
		for syscall.Kill(pid, 0) == nil {
			if time.Now().After(deadline) {
				_ = syscall.Kill(pid, syscall.SIGKILL)
				t.Fatalf("the sleep child %d outlived the read", pid)
			}
			time.Sleep(20 * time.Millisecond)
		}
	})
}
