package systemd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestUnits(t *testing.T) {
	service, err := os.ReadFile("cc-desk-check.service")
	if err != nil {
		t.Fatal(err)
	}
	timer, err := os.ReadFile("cc-desk-check.timer")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		file []byte
		line string
	}{
		{"path", service, "Environment=PATH=%h/.local/share/cc-desk/bin:%h/.local/bin:/home/linuxbrew/.linuxbrew/bin:/usr/local/bin:/usr/bin:/bin"},
		{"kill mode", service, "KillMode=process"},
		{"type", service, "Type=oneshot"},
		{"exec", service, "ExecStart=%h/.local/share/cc-desk/bin/cc-desk --state-dir %h/.local/state/cc-desk check"},
		{"startup", timer, "OnStartupSec=1min"},
		{"interval", timer, "OnUnitActiveSec=5min"},
		{"accuracy", timer, "AccuracySec=1s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !hasLine(c.file, c.line) {
				t.Errorf("no line %q", c.line)
			}
		})
	}
	t.Run("embedded service", func(t *testing.T) {
		if !bytes.Equal(Service, service) {
			t.Error("embedded service differs from the file")
		}
	})
	t.Run("embedded timer", func(t *testing.T) {
		if !bytes.Equal(Timer, timer) {
			t.Error("embedded timer differs from the file")
		}
	})
}

func hasLine(file []byte, line string) bool {
	return strings.Contains("\n"+string(file), "\n"+line+"\n")
}
