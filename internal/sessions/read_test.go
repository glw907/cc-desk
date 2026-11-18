package sessions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
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
