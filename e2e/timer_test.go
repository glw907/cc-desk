package e2e

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/cc-desk/systemd"
)

// systemctlStandIn writes a stand-in for systemctl that logs its arguments,
// one line per call, and prints canned show output. With fail set it prints
// a message and exits 1. It returns the script's path and the log's.
func systemctlStandIn(t *testing.T, fail bool) (script, log string) {
	t.Helper()
	dir := t.TempDir()
	script = filepath.Join(dir, "systemctl.sh")
	log = filepath.Join(dir, "calls.log")
	body := "#!/bin/sh\necho \"$*\" >> " + log + "\n"
	if fail {
		body += "echo 'Failed to connect to bus' >&2\nexit 1\n"
	} else {
		body += `case "$*" in
*show*cc-desk-check.timer*) printf 'ActiveState=active\nLastTriggerUSec=Fri 2026-11-20 22:57:00 UTC\nNextElapseUSecRealtime=Fri 2026-11-20 23:02:00 UTC\n' ;;
*show*cc-desk-check.service*) printf 'Result=success\nExecMainStatus=0\n' ;;
esac
`
	}
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return script, log
}

// runHome runs the binary with HOME set to home and the given stand-in for
// systemctl, on the state folder and clock named.
func runHome(t *testing.T, home, stand, dir, now string, args ...string) result {
	t.Helper()
	full := append([]string{"--state-dir", dir, "--now", now, "--systemctl-command", stand}, args...)
	cmd := exec.Command(binary, full...)
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin", "HOME="+home)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			t.Fatalf("run %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return result{out.String(), errOut.String(), code}
}

func TestTimerInstall(t *testing.T) {
	home := t.TempDir()
	stand, log := systemctlStandIn(t, false)
	state := t.TempDir()
	for i := range 2 {
		r := runHome(t, home, stand, state, workedClock, "timer", "install")
		if r.code != 0 {
			t.Fatalf("install %d: exit %d: %s", i+1, r.code, r.stderr)
		}
		bin := filepath.Join(home, ".local", "share", "cc-desk", "bin", "cc-desk")
		out, err := exec.Command(bin, "--version").CombinedOutput()
		if err != nil || !strings.Contains(string(out), "cc-desk") {
			t.Fatalf("installed binary --version: %v: %s", err, out)
		}
		for name, want := range map[string][]byte{"cc-desk-check.service": systemd.Service, "cc-desk-check.timer": systemd.Timer} {
			got, err := os.ReadFile(filepath.Join(home, ".config", "systemd", "user", name))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s differs from the repo's bytes", name)
			}
		}
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	want := "--user daemon-reload\n--user enable --now cc-desk-check.timer\n"
	if string(b) != want+want {
		t.Errorf("systemctl log = %q, want the two calls twice", b)
	}
}

func TestTimerInstallNeedsHome(t *testing.T) {
	stand, _ := systemctlStandIn(t, false)
	r := runHome(t, "", stand, t.TempDir(), workedClock, "timer", "install")
	if r.code == 0 || !strings.Contains(r.stderr, "HOME") {
		t.Errorf("exit %d, stderr %q; want a refusal naming HOME", r.code, r.stderr)
	}
}

func TestTimerStatus(t *testing.T) {
	stand, _ := systemctlStandIn(t, false)
	home := t.TempDir()
	t.Run("a recorded run", func(t *testing.T) {
		state := t.TempDir()
		writeIn(t, state, "check.json", `{"format":1,"last_run":"2026-11-20T22:57:00Z","decisions":[]}`)
		r := runHome(t, home, stand, state, workedClock, "timer", "status")
		if r.code != 0 {
			t.Fatalf("exit %d: %s", r.code, r.stderr)
		}
		for _, want := range []string{"2026-11-20T22:57:00Z", "3 minutes ago", "ActiveState: active", "LastTriggerUSec: Fri 2026-11-20 22:57:00 UTC", "NextElapseUSecRealtime: Fri 2026-11-20 23:02:00 UTC", "Result: success", "ExecMainStatus: 0"} {
			if !strings.Contains(r.stdout, want) {
				t.Errorf("stdout lacks %q:\n%s", want, r.stdout)
			}
		}
	})
	t.Run("no run", func(t *testing.T) {
		r := runHome(t, home, stand, t.TempDir(), workedClock, "timer", "status")
		if r.code != 0 || !strings.Contains(r.stdout, "no run recorded") {
			t.Errorf("exit %d, stdout %q; want no run recorded", r.code, r.stdout)
		}
	})
	t.Run("systemctl fails", func(t *testing.T) {
		bad, _ := systemctlStandIn(t, true)
		r := runHome(t, home, bad, t.TempDir(), workedClock, "timer", "status")
		if r.code != 1 || !strings.Contains(r.stderr, "Failed to connect to bus") {
			t.Errorf("exit %d, stderr %q; want exit 1 with the stand-in's message", r.code, r.stderr)
		}
	})
}
