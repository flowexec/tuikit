package io_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/flowexec/tuikit/io"
)

type testArchiveFile struct {
	name string
	age  time.Duration
	size int
}

func writeArchive(t *testing.T, files []testArchiveFile) string {
	t.Helper()
	dir := t.TempDir()
	now := time.Now()
	for _, f := range files {
		path := filepath.Join(dir, f.name)
		if err := os.WriteFile(path, make([]byte, f.size), 0600); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
		modTime := now.Add(-f.age)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("chtimes %s: %v", f.name, err)
		}
	}
	return dir
}

func remainingFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	return names
}

func TestPruneArchive(t *testing.T) {
	files := []testArchiveFile{
		{name: "a.log", age: 72 * time.Hour, size: 300},
		{name: "b.log", age: 48 * time.Hour, size: 300},
		{name: "c.log", age: 24 * time.Hour, size: 300},
		{name: "d.log", age: time.Hour, size: 300},
	}
	tests := []struct {
		name   string
		policy io.RetentionPolicy
		want   []string
	}{
		{"no limits keeps everything", io.RetentionPolicy{}, []string{"a.log", "b.log", "c.log", "d.log"}},
		{"max files keeps the newest", io.RetentionPolicy{MaxFiles: 2}, []string{"c.log", "d.log"}},
		{"max age removes older files", io.RetentionPolicy{MaxAge: 36 * time.Hour}, []string{"c.log", "d.log"}},
		{"max total size removes oldest first", io.RetentionPolicy{MaxTotalSize: 700}, []string{"c.log", "d.log"}},
		{
			"strictest limit wins",
			io.RetentionPolicy{MaxFiles: 3, MaxAge: 60 * time.Hour, MaxTotalSize: 400},
			[]string{"d.log"},
		},
		{"negative values are unlimited", io.RetentionPolicy{MaxFiles: -1, MaxAge: -1, MaxTotalSize: -1},
			[]string{"a.log", "b.log", "c.log", "d.log"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeArchive(t, files)
			if err := io.PruneArchive(dir, tt.policy); err != nil {
				t.Fatalf("prune: %v", err)
			}
			if got := remainingFiles(t, dir); !slices.Equal(got, tt.want) {
				t.Errorf("remaining files = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPruneArchive_SkipsDirectories(t *testing.T) {
	dir := writeArchive(t, []testArchiveFile{{name: "a.log", age: 72 * time.Hour}})
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := io.PruneArchive(dir, io.RetentionPolicy{MaxAge: time.Hour}); err != nil {
		t.Fatalf("prune: %v", err)
	}
	if got := remainingFiles(t, dir); !slices.Equal(got, []string{"nested"}) {
		t.Errorf("remaining files = %v, want [nested]", got)
	}
}

func TestPruneArchive_MissingDirectory(t *testing.T) {
	if err := io.PruneArchive(filepath.Join(t.TempDir(), "missing"), io.DefaultRetentionPolicy()); err == nil {
		t.Error("expected an error for a missing directory")
	}
	if err := io.PruneArchive("", io.DefaultRetentionPolicy()); err != nil {
		t.Errorf("expected no error for an empty directory path, got %v", err)
	}
}

func TestNewLogger_ArchiveRetention(t *testing.T) {
	var many []testArchiveFile
	for i := range io.MaxArchiveSize + 10 {
		many = append(many, testArchiveFile{name: fmt.Sprintf("old-%02d.log", i), age: time.Duration(i+1) * time.Hour})
	}

	t.Run("defaults to MaxArchiveSize files", func(t *testing.T) {
		dir := writeArchive(t, many)
		newArchiveLogger(t, dir)
		if got := len(remainingFiles(t, dir)); got != io.MaxArchiveSize {
			t.Errorf("remaining files = %d, want %d", got, io.MaxArchiveSize)
		}
	})

	t.Run("applies the configured policy", func(t *testing.T) {
		dir := writeArchive(t, many)
		newArchiveLogger(t, dir, io.WithArchiveRetention(io.RetentionPolicy{MaxAge: 5*time.Hour + 30*time.Minute}))
		// The five files under the age limit, plus the file the logger just created.
		if got := len(remainingFiles(t, dir)); got != 6 {
			t.Errorf("remaining files = %d, want 6", got)
		}
	})
}

func newArchiveLogger(t *testing.T, dir string, opts ...io.LoggerOptions) {
	t.Helper()
	out, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	t.Cleanup(func() { _ = out.Close() })
	logger := io.NewLogger(append([]io.LoggerOptions{
		io.WithOutput(out),
		io.WithArchiveDirectory(dir),
		io.WithArchiveID("test"),
	}, opts...)...)
	t.Cleanup(func() { _ = logger.Flush() })
}
