package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FileError is a failure in one state file. File is relative to the state
// folder, and Line is set for a JSONL line.
type FileError struct {
	File string
	Line int
	Err  error
}

func (e *FileError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s line %d: %v", e.File, e.Line, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.File, e.Err)
}

func (e *FileError) Unwrap() error { return e.Err }

// decodeStrict decodes exactly one JSON value into v. An unknown field,
// an empty input, or anything after the value is an error.
func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("syntax error: the file is empty")
		}
		return err
	}
	if err := dec.Decode(new(json.RawMessage)); !errors.Is(err, io.EOF) {
		return errors.New("syntax error: data after the first value")
	}
	return nil
}

// DecodeRelease strictly decodes release JSON and runs its checks. The
// error carries no file name.
func DecodeRelease(data []byte) (*Release, error) {
	var r Release
	if err := decodeStrict(data, &r); err != nil {
		return nil, err
	}
	if err := checkFormat(r.Format); err != nil {
		return nil, err
	}
	if err := validateRelease(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func checkFormat(f int) error {
	switch f {
	case Format:
		return nil
	case 0:
		return errors.New(`missing required field "format"`)
	}
	return fmt.Errorf("format %d is not supported, want %d", f, Format)
}

// readJSON reads rel under dir into v. It reports whether the file exists.
func readJSON(dir, rel string, v any) (bool, error) {
	data, err := os.ReadFile(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, &FileError{File: rel, Err: err}
	}
	if err := decodeStrict(data, v); err != nil {
		return true, &FileError{File: rel, Err: err}
	}
	return true, nil
}

// readJSONL decodes each line of rel into a new T. A blank line is an
// error, and the final newline is optional.
func readJSONL[T any](dir, rel string, format func(*T) int) ([]T, error) {
	data, err := os.ReadFile(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, &FileError{File: rel, Err: err}
	}
	if len(data) == 0 {
		return nil, nil
	}
	text := strings.TrimSuffix(string(data), "\n")
	var out []T
	for i, line := range strings.Split(text, "\n") {
		var v T
		if strings.TrimSpace(line) == "" {
			return nil, &FileError{File: rel, Line: i + 1, Err: errors.New("syntax error: blank line")}
		}
		if err := decodeStrict([]byte(line), &v); err != nil {
			return nil, &FileError{File: rel, Line: i + 1, Err: err}
		}
		if err := checkFormat(format(&v)); err != nil {
			return nil, &FileError{File: rel, Line: i + 1, Err: err}
		}
		out = append(out, v)
	}
	return out, nil
}

// jsonNames lists the *.json files in rel under dir by sorted name.
func jsonNames(dir, rel string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(dir, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, &FileError{File: rel, Err: err}
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func loadReleases(dir, rel string, done bool) ([]*ReleaseFile, error) {
	names, err := jsonNames(dir, rel)
	if err != nil {
		return nil, err
	}
	var out []*ReleaseFile
	for _, name := range names {
		file := path.Join(rel, name)
		var r Release
		if _, err := readJSON(dir, file, &r); err != nil {
			return nil, err
		}
		if err := checkFormat(r.Format); err != nil {
			return nil, &FileError{File: file, Err: err}
		}
		if err := validateRelease(&r); err != nil {
			return nil, &FileError{File: file, Err: err}
		}
		if !done && strings.TrimSuffix(name, ".json") != r.Project {
			return nil, &FileError{File: file, Err: fmt.Errorf("the file name does not match project %q", r.Project)}
		}
		out = append(out, &ReleaseFile{Path: file, Done: done, Release: r})
	}
	return out, nil
}

// Load reads every state file under dir, strictly, in a fixed order:
// releases, done releases, desks, settings, usage, rulings, forecasts. It
// then runs the checks across files. The first failure wins and is a
// *FileError. An absent optional file loads as empty.
func Load(dir string) (*State, error) {
	s := &State{
		Desks:    Desks{Format: Format},
		Settings: Settings{Format: Format, UsageStopPercent: DefaultUsageStop},
	}
	var err error
	if s.Releases, err = loadReleases(dir, "releases", false); err != nil {
		return nil, err
	}
	done, err := loadReleases(dir, "releases/done", true)
	if err != nil {
		return nil, err
	}
	s.Releases = append(s.Releases, done...)

	var desks Desks
	if found, err := readJSON(dir, "desks.json", &desks); err != nil {
		return nil, err
	} else if found {
		if err := checkFormat(desks.Format); err != nil {
			return nil, &FileError{File: "desks.json", Err: err}
		}
		for i, d := range desks.Desks {
			if d.SessionID == "" || d.PID <= 0 || d.Started.IsZero() {
				return nil, &FileError{File: "desks.json", Err: fmt.Errorf("desk %d is missing session_id, pid, or started", i+1)}
			}
		}
		s.Desks = desks
	}

	var settings Settings
	if found, err := readJSON(dir, "settings.json", &settings); err != nil {
		return nil, err
	} else if found {
		if err := checkFormat(settings.Format); err != nil {
			return nil, &FileError{File: "settings.json", Err: err}
		}
		if settings.UsageStopPercent == 0 {
			return nil, &FileError{File: "settings.json", Err: errors.New(`missing required field "usage_stop_percent"`)}
		}
		if settings.UsageStopPercent < 1 || settings.UsageStopPercent > 100 {
			return nil, &FileError{File: "settings.json", Err: fmt.Errorf("usage_stop_percent %d is not from 1 to 100", settings.UsageStopPercent)}
		}
		s.Settings = settings
	}

	var usage Usage
	if found, err := readJSON(dir, "usage.json", &usage); err != nil {
		return nil, err
	} else if found {
		if err := checkFormat(usage.Format); err != nil {
			return nil, &FileError{File: "usage.json", Err: err}
		}
		if usage.ReadAt.IsZero() {
			return nil, &FileError{File: "usage.json", Err: errors.New(`missing required field "read_at"`)}
		}
		s.Usage = &usage
	}

	if s.Rulings, err = readJSONL(dir, "rulings.jsonl", func(r *Ruling) int { return r.Format }); err != nil {
		return nil, err
	}
	for i := range s.Rulings {
		r := &s.Rulings[i]
		if r.ID == "" || r.Time.IsZero() || r.Desk == "" || r.Words == "" {
			return nil, &FileError{File: "rulings.jsonl", Line: i + 1, Err: errors.New("missing id, time, desk, or words")}
		}
	}
	if s.Forecasts, err = readJSONL(dir, "forecasts.jsonl", func(f *Forecast) int { return f.Format }); err != nil {
		return nil, err
	}
	for i := range s.Forecasts {
		f := &s.Forecasts[i]
		if f.Made.IsZero() || f.Project == "" || f.Round == "" || f.Time.IsZero() {
			return nil, &FileError{File: "forecasts.jsonl", Line: i + 1, Err: errors.New("missing made, project, round, or time")}
		}
	}

	if err := Check(s); err != nil {
		return nil, err
	}
	return s, nil
}
