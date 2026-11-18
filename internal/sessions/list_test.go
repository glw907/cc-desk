package sessions

import (
	"strings"
	"testing"
)

func TestParseList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		rows int
		want string // empty means no error
	}{
		{"an unknown key", `[{"id":"ab12","state":"working","newKey":{"x":1}}]`, 1, ""},
		{"an empty list", `[]`, 0, ""},
		{"an object", `{"id":"ab12"}`, 0, "not a JSON list"},
		{"null", `null`, 0, "null"},
		{"not JSON", `nope`, 0, "not a JSON list"},
		{"a row that is not an object", `[1]`, 0, "row 0"},
		{"two rows with one id", `[{"id":"ab12","state":"done"},{"id":"ab12","state":"working"}]`, 2, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := ParseList([]byte(tt.in))
			switch {
			case tt.want == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tt.want != "" && (err == nil || !strings.Contains(err.Error(), tt.want)):
				t.Fatalf("error = %v, want it to hold %q", err, tt.want)
			case len(rows) != tt.rows:
				t.Fatalf("got %d rows, want %d", len(rows), tt.rows)
			}
		})
	}
}

func TestFind(t *testing.T) {
	rows, err := ParseList([]byte(`[
		{"id":"aa","state":"done","status":"idle"},
		{"id":"aa","state":"working"},
		{"id":"bb","state":"stopped"},
		{"id":"bb","state":"failed"},
		{"id":"cc","state":"mystery"}]`))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id    string
		found bool
		state string
		class Class
		live  bool
	}{
		{"aa", true, "working", ClassWorking, true},
		{"bb", true, "stopped", ClassNone, false},
		{"cc", true, "mystery", ClassNone, false},
		{"zz", false, "", ClassNone, false},
		{"", false, "", ClassNone, false},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			r, ok := Find(rows, tt.id)
			if ok != tt.found || r.State != tt.state || r.Class() != tt.class || r.Live() != tt.live {
				t.Fatalf("Find = %+v %v", r, ok)
			}
		})
	}
	if r := (Row{State: "done"}); r.Class() != ClassDone || r.Live() {
		t.Error("a done row is the idle case and is not live")
	}
	if r := (Row{State: "blocked"}); r.Class() != ClassBlocked || !r.Live() {
		t.Error("a blocked row is live")
	}
}
