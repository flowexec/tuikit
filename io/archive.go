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

func RotateArchive(archiveDir string) {
	if archiveDir == "" {
		return
	}
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tuikit/io: failed to read archive directory: %v\n", err)
		return
	}
	if len(files) <= MaxArchiveSize {
		return
	}
	slices.SortFunc(files, func(i, j os.DirEntry) int {
		iInfo, err := i.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tuikit/io: failed to get info for archive file: %v\n", err)
		}
		jInfo, err := j.Info()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tuikit/io: failed to get info for archive file: %v\n", err)
		}
		if iInfo.ModTime().Before(jInfo.ModTime()) {
			return -1
		} else if iInfo.ModTime().After(jInfo.ModTime()) {
			return 1
		}
		return 0
	})

	for i := 0; i < len(files)-MaxArchiveSize; i++ {
		oldest := files[i]
		err := os.Remove(filepath.Clean(filepath.Join(archiveDir, oldest.Name())))
		if err != nil {
			fmt.Fprintf(os.Stderr, "tuikit/io: failed to remove oldest archive file: %v\n", err)
		}
	}
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
