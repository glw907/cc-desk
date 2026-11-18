package state

import (
	"strings"
	"testing"
)

const goodReport = `{"title":"Round r1","summary":"It landed.",
"calls":[{"call":"Name","chosen":"a","other":"b","why":"Shorter."}],
"misses":[{"miss":"A gap.","cost":"10 minutes."}]}`

func TestDecodeReport(t *testing.T) {
	tests := []struct {
		name, json, want string
	}{
		{"good", goodReport, ""},
		{"empty lists", `{"title":"T","summary":"S","calls":[],"misses":[]}`, ""},
		{"no title", strings.Replace(goodReport, `"title":"Round r1",`, "", 1), `"title"`},
		{"no summary", strings.Replace(goodReport, `"summary":"It landed.",`, "", 1), `"summary"`},
		{"no calls", `{"title":"T","summary":"S","misses":[]}`, `"calls"`},
		{"null calls", `{"title":"T","summary":"S","calls":null,"misses":[]}`, `"calls"`},
		{"no misses", `{"title":"T","summary":"S","calls":[]}`, `"misses"`},
		{"call without why", strings.Replace(goodReport, `,"why":"Shorter."`, "", 1), `calls[0] is missing "why"`},
		{"call without other", strings.Replace(goodReport, `"other":"b",`, "", 1), `calls[0] is missing "other"`},
		{"miss without cost", strings.Replace(goodReport, `,"cost":"10 minutes."`, "", 1), `misses[0] is missing "cost"`},
		{"unknown key", strings.Replace(goodReport, `"title":`, `"extra":1,"title":`, 1), "extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeReport([]byte(tt.json))
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one naming %s", err, tt.want)
			}
		})
	}
}

func TestAddReport(t *testing.T) {
	rf, err := DecodeReport([]byte(goodReport))
	if err != nil {
		t.Fatal(err)
	}
	s := loaded(t)
	var id string
	if err := AddReport("alpha", "r1", rf, &id)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	if id != "rep-0001" || len(s.Reports.Reports) != 1 || !s.reportsDirty {
		t.Fatalf("id %q, reports %+v, dirty %v", id, s.Reports.Reports, s.reportsDirty)
	}
	r := s.Reports.Reports[0]
	if r.Project != "alpha" || r.Round != "r1" || r.Read != nil || !r.Sent.Equal(at(9, 0)) || len(r.Calls) != 1 {
		t.Errorf("report = %+v", r)
	}

	if err := AddReport("alpha", "r1", rf, nil)(s, at(9, 1)); !isRefusal(err) {
		t.Errorf("a second report for the round: error = %v, want a refusal", err)
	}
	if err := AddReport("alpha", "", rf, &id)(s, at(9, 2)); err != nil || id != "rep-0002" {
		t.Errorf("release report: id %q, error %v", id, err)
	}
	if err := AddReport("alpha", "", rf, nil)(s, at(9, 3)); !isRefusal(err) {
		t.Errorf("a second release report: error = %v, want a refusal", err)
	}

	for _, tt := range []struct{ name, project, round string }{
		{"no project", "nope", "r1"},
		{"no round", "alpha", "r9"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := loaded(t)
			err := AddReport(tt.project, tt.round, rf, nil)(s, at(9, 0))
			if err == nil || isRefusal(err) || s.reportsDirty {
				t.Errorf("error = %v, dirty %v, want a plain error and no write", err, s.reportsDirty)
			}
		})
	}

	t.Run("done release", func(t *testing.T) {
		s := loaded(t)
		s.Releases[0].Done = true
		if err := AddReport("alpha", "r1", rf, nil)(s, at(9, 0)); err != nil {
			t.Errorf("a done release: %v", err)
		}
	})
}

func TestReadReport(t *testing.T) {
	rf, _ := DecodeReport([]byte(goodReport))
	s := loaded(t)
	if err := AddReport("alpha", "r1", rf, nil)(s, at(9, 0)); err != nil {
		t.Fatal(err)
	}
	s.reportsDirty = false

	if err := ReadReport("rep-0001")(s, at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if got := s.Reports.Reports[0].Read; got == nil || !got.Equal(at(10, 0)) || !s.reportsDirty || !s.noSettle {
		t.Fatalf("read = %v, dirty %v, noSettle %v", got, s.reportsDirty, s.noSettle)
	}

	s.reportsDirty = false
	if err := ReadReport("rep-0001")(s, at(11, 0)); err != nil {
		t.Fatal(err)
	}
	if got := s.Reports.Reports[0].Read; !got.Equal(at(10, 0)) || s.reportsDirty {
		t.Errorf("a second read changed the time to %v or wrote", got)
	}
	if err := ReadReport("rep-0009")(s, at(11, 0)); !isRefusal(err) {
		t.Errorf("an unknown id: error = %v, want a refusal", err)
	}
}

func TestLoadReports(t *testing.T) {
	const rep = `{"id":"rep-0001","project":"alpha","title":"T","summary":"S","calls":[],"misses":[],"sent":"2026-11-20T09:00:00Z","read":null}`
	tests := []struct {
		name, json, want string
	}{
		{"good", `{"format":1,"reports":[` + rep + `]}`, ""},
		{"a missing list loads empty", `{"format":1,"reports":[` + strings.Replace(rep, `"calls":[],"misses":[],`, "", 1) + `]}`, ""},
		{"no format", `{"reports":[` + rep + `]}`, "format"},
		{"unknown key", `{"format":1,"reports":[` + strings.Replace(rep, `"id"`, `"extra":1,"id"`, 1) + `]}`, "extra"},
		{"no title", `{"format":1,"reports":[` + strings.Replace(rep, `"title":"T",`, "", 1) + `]}`, "report 1"},
		{"same id twice", `{"format":1,"reports":[` + rep + `,` + rep + `]}`, "used twice"},
		{"not json", `{`, "reports.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, ReportsFile, tt.json)
			s, err := Load(dir)
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				if r := s.Reports.Reports[0]; r.Calls == nil || r.Misses == nil {
					t.Errorf("lists = %v, %v, want empty, not nil", r.Calls, r.Misses)
				}
				return
			}
			fe, ok := err.(*FileError)
			if !ok || fe.File != ReportsFile || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want a %s error naming %q", err, ReportsFile, tt.want)
			}
		})
	}
}
