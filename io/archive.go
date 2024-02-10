package io

import (
	"fmt"
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

func NewArchiveLogFile(archiveDir string) *os.File {
	if dir, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.MkdirAll(archiveDir, 0755)
		if err != nil {
			panic(fmt.Errorf("failed to create archive directory: %w", err))
		}
	} else if !dir.IsDir() {
		panic(fmt.Errorf("archive directory is not a directory"))
	}
	writer, err := os.Create(filepath.Clean(
		filepath.Join(archiveDir, fmt.Sprintf("%s.log", time.Now().Local().Format(LogEntryTimeFormat))),
	))
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
		err := os.Remove(filepath.Clean(fmt.Sprintf("%s/%s", logger.archiveDir, oldest.Name())))
		if err != nil {
			logger.Fatalf("failed to remove oldest archive file: %s", err)
		}
	}
}

func ReadArchiveEntry(entry string) (string, error) {
	data, err := os.ReadFile(filepath.Clean(entry))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ListArchiveEntries(archiveDir string) ([]string, error) {
	_, err := os.Stat(archiveDir)
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, err
	}

	var entries []string
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
		name := filepath.Base(file.Name())
		_, err := time.Parse(LogEntryTimeFormat, strings.TrimSuffix(name, ".log"))
		if err != nil {
			continue
		}
		entries = append(entries, filepath.Join(archiveDir, name))
	}
	slices.SortFunc(entries, func(i, j string) int {
		iTime, err := time.Parse(LogEntryTimeFormat, filepath.Base(i))
		if err != nil {
			return 0
		}
		jTime, err := time.Parse(LogEntryTimeFormat, filepath.Base(j))
		if err != nil {
			return 0
		}
		if iTime.Before(jTime) {
			return -1
		} else if iTime.After(jTime) {
			return 1
		}
		return 0
	})

	return entries, nil
}

func DeleteArchiveEntry(entry string) error {
	return os.Remove(filepath.Clean(entry))
}
