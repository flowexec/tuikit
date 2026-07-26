package io_test

import (
	"os"
	"sync"
	"testing"

	"github.com/flowexec/tuikit/io"
)

// TestStdWriters_ConcurrentDoesNotRace drives a StdOutWriter and StdErrWriter concurrently,
// the way os/exec copies a command's two output streams. Both writers temporarily flip the
// shared logger's mode; before serialization this raced on the mode field. The writers request
// a mode different from the logger's so the flip/restore path is exercised. Run with -race.
func TestStdWriters_ConcurrentDoesNotRace(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "tuikit-log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	logger := io.NewLogger(io.WithOutput(f), io.WithMode(io.Text))
	desired := io.Logfmt
	out := io.StdOutWriter{Logger: logger, LogMode: &desired}
	errW := io.StdErrWriter{Logger: logger, LogMode: &desired}

	const writesPerStream = 100
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < writesPerStream; i++ {
			if _, wErr := out.Write([]byte("stdout line\n")); wErr != nil {
				t.Errorf("stdout write: %v", wErr)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < writesPerStream; i++ {
			if _, wErr := errW.Write([]byte("stderr line\n")); wErr != nil {
				t.Errorf("stderr write: %v", wErr)
				return
			}
		}
	}()
	wg.Wait()

	// The logger's mode must be restored to its original value after the writers finish.
	if logger.LogMode() != io.Text {
		t.Errorf("logger mode = %q after writes, want %q", logger.LogMode(), io.Text)
	}
}
