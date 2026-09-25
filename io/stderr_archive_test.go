package io_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/flowexec/tuikit/io"
)

// archiveLogger returns a logger that writes terminal output to a temp file and archives to a
// temp directory, plus readers for both.
func archiveLogger(t *testing.T, mode io.LogMode) (*io.StandardLogger, func() string, func() string) {
	t.Helper()
	dir := t.TempDir()
	out, err := os.CreateTemp(dir, "out")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { _ = out.Close() })

	archiveDir := filepath.Join(dir, "archive")
	logger := io.NewLogger(
		io.WithMode(mode),
		io.WithOutput(out),
		io.WithArchiveDirectory(archiveDir),
		io.WithArchiveID("test"),
	)

	readOut := func() string {
		data, rErr := os.ReadFile(out.Name())
		if rErr != nil {
			t.Fatalf("read output: %v", rErr)
		}
		return string(data)
	}
	readArchive := func() string {
		if fErr := logger.Flush(); fErr != nil {
			t.Fatalf("flush: %v", fErr)
		}
		entries, rErr := os.ReadDir(archiveDir)
		if rErr != nil {
			t.Fatalf("read archive dir: %v", rErr)
		}
		if len(entries) != 1 {
			t.Fatalf("expected 1 archive file, got %d", len(entries))
		}
		data, rErr := os.ReadFile(filepath.Join(archiveDir, entries[0].Name()))
		if rErr != nil {
			t.Fatalf("read archive: %v", rErr)
		}
		return string(data)
	}
	return logger, readOut, readArchive
}

var kitchenTime = regexp.MustCompile(`\d{1,2}:\d{2}(AM|PM)`)

func TestStdErrWriter_ArchivesStderrAsInfoWithTask(t *testing.T) {
	logger, readOut, readArchive := archiveLogger(t, io.Logfmt)
	writer := io.StdErrWriter{Logger: logger, Task: &io.TaskContext{Name: "build"}}

	if _, err := writer.Write([]byte("warn: deprecated option\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	out := readOut()
	if !strings.Contains(out, "[build]") || !strings.Contains(out, "warn: deprecated option") {
		t.Errorf("expected terminal output to keep the task prefix and line, got: %q", out)
	}

	archive := readArchive()
	for _, want := range []string{"level=info", `msg="warn: deprecated option"`, "task=build", "stream=stderr"} {
		if !strings.Contains(archive, want) {
			t.Errorf("expected archive to contain %q, got: %q", want, archive)
		}
	}
	if strings.Contains(archive, "level=error") {
		t.Errorf("expected no level=error in archive, got: %q", archive)
	}
}

func TestStdErrWriter_ArchivesStderrAsInfoWithoutTask(t *testing.T) {
	for _, mode := range []io.LogMode{io.Logfmt, io.JSON} {
		t.Run(string(mode), func(t *testing.T) {
			logger, readOut, readArchive := archiveLogger(t, mode)
			writer := io.StdErrWriter{Logger: logger, LogFields: []any{"exec", "lint"}}

			if _, err := writer.Write([]byte("warn: deprecated option\n")); err != nil {
				t.Fatalf("write: %v", err)
			}

			// Terminal output must match what Notice rendered before this change.
			ref, refOut, _ := archiveLogger(t, mode)
			// Called through the interface: vet infers StandardLogger.Notice is printf-like
			// because it archives via Errorf.
			var refLogger io.Logger = ref
			refLogger.Notice("warn: deprecated option", "exec", "lint")
			got := kitchenTime.ReplaceAllString(readOut(), "")
			want := kitchenTime.ReplaceAllString(refOut(), "")
			if mode == io.JSON {
				got = regexp.MustCompile(`"time":"[^"]*"`).ReplaceAllString(got, "")
				want = regexp.MustCompile(`"time":"[^"]*"`).ReplaceAllString(want, "")
			}
			if got != want {
				t.Errorf("terminal output changed:\n got: %q\nwant: %q", got, want)
			}

			archive := readArchive()
			for _, want := range []string{"level=info", "exec=lint", "stream=stderr"} {
				if !strings.Contains(archive, want) {
					t.Errorf("expected archive to contain %q, got: %q", want, archive)
				}
			}
			if strings.Contains(archive, "level=error") {
				t.Errorf("expected no level=error in archive, got: %q", archive)
			}
		})
	}
}

func TestStdOutWriter_ArchiveHasNoStreamField(t *testing.T) {
	logger, _, readArchive := archiveLogger(t, io.Logfmt)
	writer := io.StdOutWriter{Logger: logger, Task: &io.TaskContext{Name: "build"}}

	if _, err := writer.Write([]byte("compiling...\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	archive := readArchive()
	if !strings.Contains(archive, "level=info") || !strings.Contains(archive, "task=build") {
		t.Errorf("expected info line with task, got: %q", archive)
	}
	if strings.Contains(archive, "stream=") {
		t.Errorf("expected no stream field on stdout lines, got: %q", archive)
	}
}

func TestLoggerErrors_StillArchiveAsError(t *testing.T) {
	logger, _, readArchive := archiveLogger(t, io.Logfmt)
	logger.Error("run failed", "exec", "lint")
	logger.Errorf("exit status %d", 1)

	archive := readArchive()
	lines := strings.Split(strings.TrimSpace(archive), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 archive lines, got %d: %q", len(lines), archive)
	}
	for _, line := range lines {
		if !strings.Contains(line, "level=error") {
			t.Errorf("expected level=error, got: %q", line)
		}
		if strings.Contains(line, "stream=") {
			t.Errorf("expected no stream field on logger errors, got: %q", line)
		}
	}
}
