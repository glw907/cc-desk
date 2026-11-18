package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/glw907/cc-desk/internal/atomicfile"
)

// File names in the state folder that Commit owns.
const (
	SnapshotFile      = "snapshot.json"
	SnapshotErrorFile = "snapshot-error.json"
	UsageFile         = "usage.json"
)

// Event changes the loaded state and marks each release file it changed
// as Dirty. It returns a *RefusedError when the state doesn't allow the
// change, and any other error for a bad request.
type Event func(s *State, now time.Time) error

// ForecastFile is the append-only log of next-session forecasts.
const ForecastFile = "forecasts.jsonl"

// Render turns the state into the bytes of snapshot.json and the forecast
// line the same build gives for the next session. Commit stamps the line's
// format and time made.
type Render func(s *State, now time.Time) ([]byte, Forecast, error)

type snapshotError struct {
	Format  int    `json:"format"`
	File    string `json:"file"`
	Message string `json:"message"`
}

// Commit is the one path every verb runs. It takes the lock on dir,
// waiting up to wait, then loads and checks every file. A load or check
// failure writes snapshot-error.json and leaves every other file alone.
// A clean load removes that file, so a refusal clears it too. Commit
// then applies ev, settles, writes each changed release and a changed
// usage.json atomically, appends a forecast line when ev started, finished, or stopped a pass,
// and rewrites snapshot.json. It returns the snapshot's bytes. A nil ev
// changes nothing.
func Commit(dir string, now time.Time, wait time.Duration, ev Event, render Render) ([]byte, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create state folder: %w", err)
	}
	unlock, err := Lock(dir, wait)
	if err != nil {
		return nil, err
	}
	defer unlock()

	s, err := Load(dir)
	if err != nil {
		if fe, ok := errors.AsType[*FileError](err); ok {
			if werr := writeSnapshotError(dir, fe); werr != nil {
				return nil, fmt.Errorf("%w (and the error file could not be written: %v)", err, werr)
			}
		}
		return nil, err
	}
	if err := os.Remove(filepath.Join(dir, SnapshotErrorFile)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("remove %s: %w", SnapshotErrorFile, err)
	}

	if ev != nil {
		if err := ev(s, now); err != nil {
			return nil, err
		}
	}
	settle(s, now)

	for _, rf := range s.Releases {
		if !rf.Dirty {
			continue
		}
		data, err := marshal(rf.Release)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", rf.Path, err)
		}
		full := filepath.Join(dir, rf.Path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, fmt.Errorf("create folder for %s: %w", rf.Path, err)
		}
		if err := atomicfile.WriteFile(full, data, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", rf.Path, err)
		}
	}

	if s.usageDirty {
		data, err := marshal(s.Usage)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", UsageFile, err)
		}
		if err := atomicfile.WriteFile(filepath.Join(dir, UsageFile), data, 0o644); err != nil {
			return nil, fmt.Errorf("write %s: %w", UsageFile, err)
		}
	}

	snap, fc, err := render(s, now)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	if s.passEvent {
		fc.Format, fc.Made = Format, now.UTC()
		if err := appendForecast(dir, fc); err != nil {
			return nil, err
		}
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, SnapshotFile), snap, 0o644); err != nil {
		return nil, fmt.Errorf("write %s: %w", SnapshotFile, err)
	}
	return snap, nil
}

// WithReading wraps ev with a fresh usage reading. Inside the lock it keeps
// one reading: the fresh one when forced, when no reading is loaded, when
// the loaded one is older, or when the loaded one is dated after the clock.
// On a tie the loaded reading stays. It then settles, so a waiting pass that
// now fits starts before ev can take its place, and runs ev, which may be
// nil. usage.json is written only when the fresh reading won and ev
// succeeded.
func WithReading(u Usage, forced bool, ev Event) Event {
	return func(s *State, now time.Time) error {
		u.Format = Format
		if forced || s.Usage == nil || s.Usage.ReadAt.Before(u.ReadAt) || s.Usage.ReadAt.After(now) {
			s.Usage = &u
			s.usageDirty = true
		}
		settle(s, now)
		if ev == nil {
			return nil
		}
		return ev(s, now)
	}
}

// settle starts the waiting passes that now fit, oldest request first.
// Ties go to release order, then pass order. Each pass sees the passes
// settled before it as building. A pass that still waits has its limit
// list updated in place. Settling never re-checks the four names: a pass
// that passed them when it was accepted keeps its place.
func settle(s *State, now time.Time) {
	type entry struct {
		rf *ReleaseFile
		p  *Pass
		at int
	}
	var waiting []entry
	n := 0
	for _, rf := range s.Releases {
		if rf.Done {
			continue
		}
		for i := range rf.Passes {
			if p := &rf.Passes[i]; p.State == Waiting {
				waiting = append(waiting, entry{rf, p, n})
				n++
			}
		}
	}
	slices.SortStableFunc(waiting, func(a, b entry) int { return a.p.Wait.Since.Compare(b.p.Wait.Since) })
	for _, e := range waiting {
		limits := s.Limits()
		if len(limits) == 0 {
			t := now.UTC()
			e.p.State, e.p.Started, e.p.Wait = Building, &t, nil
			e.rf.Dirty = true
			s.passEvent = true
			continue
		}
		if !slices.Equal(limits, e.p.Wait.Limits) {
			e.p.Wait.Limits = limits
			e.rf.Dirty = true
		}
	}
}

// appendForecast adds one line to forecasts.jsonl by rewriting the whole
// file atomically, so a crash never leaves half a line. A last line with no
// newline gets one first.
func appendForecast(dir string, f Forecast) error {
	line, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("encode forecast: %w", err)
	}
	path := filepath.Join(dir, ForecastFile)
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read %s: %w", ForecastFile, err)
	}
	if n := len(data); n > 0 && data[n-1] != '\n' {
		data = append(data, '\n')
	}
	data = append(append(data, line...), '\n')
	if err := atomicfile.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ForecastFile, err)
	}
	return nil
}

func writeSnapshotError(dir string, fe *FileError) error {
	data, err := marshal(snapshotError{Format: Format, File: fe.File, Message: fe.Error()})
	if err != nil {
		return err
	}
	return atomicfile.WriteFile(filepath.Join(dir, SnapshotErrorFile), data, 0o644)
}

func marshal(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
