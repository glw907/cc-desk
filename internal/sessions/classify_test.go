package sessions

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// idleCause is the plan's cause string for the unreadable idle screen: its
// last 200 runes, which read "used", so the limit pattern does not match.
const idleCause = "ng… 3 Brewing… ✶ * ✢ Brewing… · Brewing… Brewing… Brewing… Brewing… ✢ Brewing… 4 ✻ Crunched for 4s · done 6:34 PM ❯ You've used 87% of your weekly limit · resets Nov 20, 11pm (America/Anchorage) 2 s 2"

func TestLastLine(t *testing.T) {
	idle := readFile(t, "testdata/logs-idle-1.txt")
	blocked := readFile(t, "testdata/logs-blocked.txt")
	long := strings.Repeat("é", 300)
	tests := []struct {
		name           string
		detail, screen string
		want           string
	}{
		{"the idle spike screen", "", idle, idleCause},
		{"the job detail wins", "Login expired", idle, "Login expired"},
		{"a blank detail falls to the screen", "  ", "a\n\n  last   line \n\n", "last line"},
		{"escape sequences go", "", "\x1b7\x1b8\x1b[1;31mred\x1b[0m \x1b]0;title\x07ok\x1b]8;;x\x1b\\ end", "red ok end"},
		{"runes, not bytes", "", long, strings.Repeat("é", 200)},
		{"no text", "", " \n\n", "no text"},
		{"no screen and no detail", "", "", "no text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastLine(tt.detail, tt.screen)
			if got != tt.want {
				t.Fatalf("LastLine = %q, want %q", got, tt.want)
			}
			if !utf8.ValidString(got) {
				t.Fatal("result is not valid UTF-8")
			}
		})
	}
	t.Run("no escape byte survives the blocked spike screen", func(t *testing.T) {
		got := LastLine("", blocked)
		if strings.ContainsRune(got, '\x1b') || utf8.RuneCountInString(got) > 200 || got == "no text" {
			t.Fatalf("LastLine = %q", got)
		}
		if !strings.Contains(blocked, "ERR_PROXY_TUNNEL") {
			t.Fatal("fixture no longer holds ERR_PROXY_TUNNEL")
		}
	})
	t.Run("the idle screen opens with escape bytes", func(t *testing.T) {
		if !strings.Contains(idle, "\x1b7\x1b8") {
			t.Fatal("fixture lost its escape bytes, so the strip test proves nothing")
		}
		if strings.ContainsRune(LastLine("", idle), '\x1b') {
			t.Fatal("an escape byte survived")
		}
	})
}

func TestLimitPatternReadsHitNotUsed(t *testing.T) {
	row := Row{State: "blocked", Status: "idle"}
	if got := Classify(row, Job{}, "You've used 87% of your weekly limit"); got != LabelCantTell {
		t.Errorf("used: %q, want can't tell", got)
	}
	if got := Classify(row, Job{}, "You've hit your weekly limit · resets Nov 6"); got != LabelUsageLimit {
		t.Errorf("hit: %q, want usage limit", got)
	}
}

func TestClassifyFixtures(t *testing.T) {
	var cases []struct {
		Name      string            `json:"name"`
		Row       Row               `json:"row"`
		Logs      string            `json:"logs"`
		StateJSON map[string]string `json:"state_json"`
		Expect    string            `json:"expect"`
	}
	if err := json.Unmarshal([]byte(readFile(t, "testdata/fixtures.json")), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 16 {
		t.Fatalf("fixtures.json holds %d cases, want 16", len(cases))
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			job := Job{Detail: c.StateJSON["detail"], Needs: c.StateJSON["needs"]}
			if got := Classify(c.Row, job, c.Logs); string(got) != c.Expect {
				t.Fatalf("Classify = %q, want %q", got, c.Expect)
			}
		})
	}
}
