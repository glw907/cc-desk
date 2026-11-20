package sessions

import (
	"testing"
	"time"
)

func TestParseReset(t *testing.T) {
	at := func(s string) time.Time {
		t.Helper()
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	tests := []struct {
		name  string
		text  string
		began string
		want  string // empty when the text is unreadable
	}{
		{"dated with minutes", "You've hit your weekly limit · resets Nov 20, 10:59pm (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-21T07:59:00Z"},
		{"dated, hour only", "You've hit your weekly limit · resets Nov 6, 11pm (America/Anchorage)", "2026-11-06T20:00:00Z", "2026-11-07T08:00:00Z"},
		{"dated, am with minutes", "resets Nov 21, 1:00am (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-21T10:00:00Z"},
		{"dated, am hour only", "x · resets Nov 21, 1am (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-21T10:00:00Z"},
		{"noon and midnight", "resets Nov 21, 12am (UTC)", "2026-11-20T20:00:00Z", "2026-11-21T00:00:00Z"},
		{"noon", "resets Nov 21, 12:30pm (UTC)", "2026-11-20T20:00:00Z", "2026-11-21T12:30:00Z"},
		{"the date wraps the year", "resets Jan 2, 1am (America/Anchorage)", "2026-12-30T20:00:00Z", "2027-01-02T10:00:00Z"},
		{"a past date in the stop's year stays", "resets Nov 6, 11pm (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-07T08:00:00Z"},
		{"a time alone later the same day", "resets 11pm (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-21T08:00:00Z"},
		{"a time alone later in the day", "resets 3:00pm (America/Anchorage)", "2026-11-20T20:00:00Z", "2026-11-21T00:00:00Z"},
		{"a time alone already passed today reads tomorrow", "resets 3:00pm (America/Anchorage)", "2026-11-21T01:00:00Z", "2026-11-22T00:00:00Z"},
		{"an unknown zone", "resets Nov 20, 10:59pm (Mars/Olympus)", "2026-11-20T20:00:00Z", ""},
		{"no resets", "You've hit your Fable limit", "2026-11-20T20:00:00Z", ""},
		{"resets with no zone", "Current session: 96% used · resets 3:00pm", "2026-11-20T20:00:00Z", ""},
		{"an hour out of range", "resets Nov 20, 13pm (UTC)", "2026-11-20T20:00:00Z", ""},
		{"minutes out of range", "resets Nov 20, 1:75pm (UTC)", "2026-11-20T20:00:00Z", ""},
		{"a day that does not exist", "resets Feb 30, 1pm (UTC)", "2026-11-20T20:00:00Z", ""},
		{"an unknown month", "resets Foo 3, 1pm (UTC)", "2026-11-20T20:00:00Z", ""},
		{"empty", "", "2026-11-20T20:00:00Z", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseReset(tt.text, at(tt.began))
			if tt.want == "" {
				if ok {
					t.Fatalf("ParseReset(%q) = %s, want unreadable", tt.text, got)
				}
				return
			}
			if !ok || !got.Equal(at(tt.want)) {
				t.Fatalf("ParseReset(%q) = %s, %v; want %s", tt.text, got, ok, tt.want)
			}
		})
	}
}
