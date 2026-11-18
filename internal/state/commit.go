package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/cc-desk/internal/atomicfile"
)

// File names in the state folder that Commit owns.
const (
	SnapshotFile      = "snapshot.json"
	SnapshotErrorFile = "snapshot-error.json"
)

// Event changes the loaded state and marks each release file it changed
// as Dirty. It returns a *RefusedError when the state doesn't allow the
// change, and any other error for a bad request.
type Event func(s *State, now time.Time) error

// Render turns the state into the bytes of snapshot.json.
type Render func(s *State, now time.Time) ([]byte, error)

type snapshotError struct {
	Format  int    `json:"format"`
	File    string `json:"file"`
	Message string `json:"message"`
}

// Commit is the one path every verb runs. It takes the lock on dir,
// waiting up to wait, then loads and checks every file. A load or check
// failure writes snapshot-error.json and leaves every other file alone.
// A clean load removes that file, so a refusal clears it too. Commit
// then applies ev, settles, writes each changed release atomically, and
// rewrites snapshot.json. It returns the snapshot's bytes. A nil ev
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

	snap, err := render(s, now)
	if err != nil {
		return nil, fmt.Errorf("build snapshot: %w", err)
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, SnapshotFile), snap, 0o644); err != nil {
		return nil, fmt.Errorf("write %s: %w", SnapshotFile, err)
	}
	return snap, nil
}

// settle starts the passes that wait on a limit and now fit. Pass 3 fills
// it in.
func settle(s *State, now time.Time) {}

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
