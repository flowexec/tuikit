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
	Args string
	Time time.Time
	Path string
}

func (e ArchiveEntry) Title() string {
	return e.Args
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

func NewArchiveLogFile(archiveDir string) *os.File {
	if dir, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.MkdirAll(archiveDir, 0750)
		if err != nil {
			panic(fmt.Errorf("failed to create archive directory: %w", err))
		}
	} else if !dir.IsDir() {
		panic(fmt.Errorf("archive directory is not a directory"))
	}
	writer, err := os.Create(filepath.Clean(filepath.Join(archiveDir, NewArchiveFileName())))
	if err != nil {
		panic(fmt.Errorf("failed to create archive log file: %w", err))
	}
	return writer
}

func RotateArchive(logger *Logger) {
	if logger.archiveDir == "" {
		return
	}
	files, err := os.ReadDir(logger.archiveDir)
	if err != nil {
		logger.Fatalf("failed to read archive directory: %s", err)
	}
	if len(files) < MaxArchiveSize {
		return
	}
	slices.SortFunc(files, func(i, j os.DirEntry) int {
		iInfo, err := i.Info()
		if err != nil {
			logger.Fatalf("failed to get info for archive file: %s", err)
		}
		jInfo, err := j.Info()
		if err != nil {
			logger.Fatalf("failed to get info for archive file: %s", err)
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
		err := os.Remove(filepath.Clean(filepath.Join(logger.archiveDir, oldest.Name())))
		if err != nil {
			logger.Fatalf("failed to remove oldest archive file: %s", err)
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
		args, ts, err := ParseArchiveFileMetadata(file.Name())
		if err != nil {
			continue
		}
		entries = append(entries, ArchiveEntry{
			Args: args,
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

func NewArchiveFileName() string {
	args := os.Args
	argsStr := url.QueryEscape(strings.Join(args[1:], " "))
	return fmt.Sprintf("%s__%s.log", argsStr, time.Now().Local().Format(LogEntryTimeFormat))
}

func ParseArchiveFileMetadata(name string) (args string, ts time.Time, err error) {
	parts := strings.Split(strings.TrimSuffix(filepath.Base(name), ".log"), "__")
	if len(parts) != 2 {
		return "", time.Time{}, errors.New("invalid archive file name")
	}
	args, err = url.QueryUnescape(parts[0])
	if err != nil {
		return "", time.Time{}, err
	}
	ts, err = time.Parse(LogEntryTimeFormat, parts[1])
	if err != nil {
		return "", time.Time{}, err
	}
	return args, ts, nil
}
