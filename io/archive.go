package io

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const (
	MaxArchiveSize     = 50
	LogEntryTimeFormat = "2006-01-02-15-04-05"
)

type ArchiveEntry struct {
	ID   string
	Time time.Time
	Path string
}

func (e ArchiveEntry) Title() string {
	return e.ID
}
func (e ArchiveEntry) Description() string {
	return e.Time.Format("03:04PM 01/02/2006")
}
func (e ArchiveEntry) FilterValue() string {
	return e.Title() + " " + e.Description()
}

func (e ArchiveEntry) Read() (string, error) {
	if e.Path == "" {
		return "", errors.New("archive entry path is empty")
	}
	data, err := os.ReadFile(filepath.Clean(e.Path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func NewArchiveLogFile(archiveDir, id string) *os.File {
	if dir, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.MkdirAll(archiveDir, 0750)
		if err != nil {
			panic(fmt.Errorf("failed to create archive directory: %w", err))
		}
	} else if !dir.IsDir() {
		panic(fmt.Errorf("archive directory is not a directory"))
	}
	writer, err := os.Create(filepath.Clean(filepath.Join(archiveDir, NewArchiveFileName(id))))
	if err != nil {
		panic(fmt.Errorf("failed to create archive log file: %w", err))
	}
	return writer
}

// RetentionPolicy bounds how many archived log files are kept. Each limit applies on its own and
// a zero (or negative) value means that dimension is unlimited. Files are pruned oldest first.
type RetentionPolicy struct {
	// MaxFiles is the maximum number of archive files to keep.
	MaxFiles int
	// MaxAge removes archive files last modified longer ago than this.
	MaxAge time.Duration
	// MaxTotalSize is the maximum combined size of the archive files, in bytes.
	MaxTotalSize int64
}

// DefaultRetentionPolicy keeps the newest MaxArchiveSize files, with no age or size limit.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{MaxFiles: MaxArchiveSize}
}

// RotateArchive prunes archiveDir with the DefaultRetentionPolicy.
func RotateArchive(archiveDir string) {
	if err := PruneArchive(archiveDir, DefaultRetentionPolicy()); err != nil {
		fmt.Fprintf(os.Stderr, "tuikit/io: %v\n", err)
	}
}

type archiveFile struct {
	path    string
	modTime time.Time
	size    int64
}

// PruneArchive removes files from archiveDir until it satisfies policy. Subdirectories are left
// alone. Removal continues past individual failures; all of them are returned together.
func PruneArchive(archiveDir string, policy RetentionPolicy) error {
	if archiveDir == "" {
		return nil
	}
	dirEntries, err := os.ReadDir(archiveDir)
	if err != nil {
		return fmt.Errorf("failed to read archive directory: %w", err)
	}

	var errs []error
	files := make([]archiveFile, 0, len(dirEntries))
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get info for archive file: %w", err))
			continue
		}
		files = append(files, archiveFile{
			path:    filepath.Clean(filepath.Join(archiveDir, entry.Name())),
			modTime: info.ModTime(),
			size:    info.Size(),
		})
	}
	slices.SortStableFunc(files, func(i, j archiveFile) int {
		return i.modTime.Compare(j.modTime)
	})

	var totalSize int64
	for _, f := range files {
		totalSize += f.size
	}
	cutoff := time.Time{}
	if policy.MaxAge > 0 {
		cutoff = time.Now().Add(-policy.MaxAge)
	}

	remaining := len(files)
	for _, f := range files {
		overCount := policy.MaxFiles > 0 && remaining > policy.MaxFiles
		overSize := policy.MaxTotalSize > 0 && totalSize > policy.MaxTotalSize
		expired := !cutoff.IsZero() && f.modTime.Before(cutoff)
		if !overCount && !overSize && !expired {
			// Files are sorted oldest first, so nothing newer can be expired either.
			break
		}
		if err := os.Remove(f.path); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove archive file: %w", err))
			continue
		}
		remaining--
		totalSize -= f.size
	}
	return errors.Join(errs...)
}

func ListArchiveEntries(archiveDir string) ([]ArchiveEntry, error) {
	_, err := os.Stat(archiveDir)
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, err
	}

	var entries []ArchiveEntry
	for _, file := range files {
		if file.IsDir() {
			continue
		} else if filepath.Ext(file.Name()) != ".log" {
			continue
		}
		if info, err := file.Info(); err != nil {
			return nil, err
		} else if info.Size() == 0 {
			continue
		}
		id, ts, err := ParseArchiveFileMetadata(file.Name())
		if err != nil {
			continue
		}
		entries = append(entries, ArchiveEntry{
			ID:   id,
			Time: ts,
			Path: filepath.Clean(filepath.Join(archiveDir, file.Name())),
		})
	}
	slices.SortFunc(entries, func(i, j ArchiveEntry) int {
		if i.Time.Before(j.Time) {
			return -1
		} else if i.Time.After(j.Time) {
			return 1
		}
		return 0
	})

	return entries, nil
}

func DeleteArchiveEntry(entryPath string) error {
	return os.Remove(filepath.Clean(entryPath))
}

func NewArchiveFileName(id string) string {
	safeID := sanitizeArchiveID(id)
	if safeID == "" {
		safeID = "session"
	}
	return fmt.Sprintf("%s__%s.log", safeID, time.Now().Local().Format(LogEntryTimeFormat))
}

func ParseArchiveFileMetadata(name string) (id string, ts time.Time, err error) {
	parts := strings.Split(strings.TrimSuffix(filepath.Base(name), ".log"), "__")
	if len(parts) != 2 {
		return "", time.Time{}, errors.New("invalid archive file name")
	}
	rawID := parts[0]
	// Legacy: old filenames used URL-encoded os.Args; decode them gracefully
	if strings.Contains(rawID, "%") {
		if decoded, decErr := url.QueryUnescape(rawID); decErr == nil {
			rawID = decoded
		}
	}
	id = rawID
	ts, err = time.Parse(LogEntryTimeFormat, parts[1])
	if err != nil {
		return "", time.Time{}, err
	}
	return id, ts, nil
}

func sanitizeArchiveID(id string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\x00':
			return '_'
		}
		return r
	}, id)
}
