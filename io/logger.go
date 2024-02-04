package io

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"

	"github.com/jahvon/tuikit/styles"
)

const (
	MaxArchiveSize = 50
)

type Logger struct {
	stdOutHandler  *log.Logger
	archiveHandler *log.Logger
	style          styles.Theme
	archiveDir     string
	archiveFile    *os.File
}

func NewLogger(style styles.Theme, archiveDir string) *Logger {
	logger := &Logger{style: style}
	stdOutHandler := log.NewWithOptions(
		os.Stdout,
		log.Options{
			ReportTimestamp: true,
			ReportCaller:    false,
			Level:           log.InfoLevel,
		},
	)
	applyHumanReadableFormat(stdOutHandler, style)
	logger.stdOutHandler = stdOutHandler

	if archiveDir != "" {
		archiveFile := newArchiveLogFile(archiveDir)
		archiveHandler := log.NewWithOptions(
			archiveFile,
			log.Options{
				ReportTimestamp: true,
				ReportCaller:    false,
				Level:           log.DebugLevel,
			},
		)
		applyStorageFormat(archiveHandler)
		logger.archiveFile = archiveFile
		logger.archiveHandler = archiveHandler
		rotateArchive(logger)
	}

	return logger
}

func applyHumanReadableFormat(handler *log.Logger, style styles.Theme) {
	handler.SetFormatter(log.TextFormatter)
	handler.SetTimeFormat(time.Kitchen)
	handler.SetStyles(style.LoggerStyles())
	handler.SetColorProfile(termenv.ColorProfile())
}

func applyStorageFormat(handler *log.Logger) {
	handler.SetFormatter(log.JSONFormatter)
	handler.SetTimeFormat(time.RFC822)
	handler.SetStyles(log.DefaultStyles())
}

func newArchiveLogFile(archiveDir string) *os.File {
	if dir, err := os.Stat(archiveDir); os.IsNotExist(err) {
		err := os.MkdirAll(archiveDir, 0755)
		if err != nil {
			panic(fmt.Errorf("failed to create archive directory: %w", err))
		}
	} else if !dir.IsDir() {
		panic(fmt.Errorf("archive directory is not a directory"))
	}
	writer, err := os.Create(filepath.Clean(
		fmt.Sprintf("%s/%s.log", archiveDir, time.Now().Format("2006-01-02-15-04-05")),
	))
	if err != nil {
		panic(fmt.Errorf("failed to create archive log file: %w", err))
	}
	return writer
}

func rotateArchive(logger *Logger) {
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

// SetLevel sets the log level for the logger.
// -1 = Fatal
// 0 = Info
// 1 = Debug
// Default is Info.
func (l *Logger) SetLevel(level int) {
	switch level {
	case -1:
		l.stdOutHandler.SetLevel(log.FatalLevel)
	case 0:
		l.stdOutHandler.SetLevel(log.InfoLevel)
	case 1:
		l.stdOutHandler.SetLevel(log.DebugLevel)
	default:
		l.stdOutHandler.SetLevel(log.InfoLevel)
	}
}

func (l *Logger) Println(data string) {
	_, err := fmt.Fprintln(os.Stdout, data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, data)
	}
}

func (l *Logger) AsPlainText(exec func()) {
	l.stdOutHandler.SetFormatter(log.TextFormatter)
	if l.archiveHandler != nil {
		l.archiveHandler.SetFormatter(log.TextFormatter)
	}

	exec()

	l.stdOutHandler.SetFormatter(log.LogfmtFormatter)
	if l.archiveHandler != nil {
		l.archiveHandler.SetFormatter(log.LogfmtFormatter)
	}
}

func (l *Logger) Infof(msg string, args ...any) {
	l.stdOutHandler.Infof(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Infof(msg, args...)
	}
}

func (l *Logger) Debugf(msg string, args ...any) {
	l.stdOutHandler.Debugf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debugf(msg, args...)
	}
}

func (l *Logger) Error(err error, msg string) {
	if msg == "" {
		l.Errorf(err.Error())
		return
	}
	l.Errorx(err.Error(), "err", err)
}

func (l *Logger) Errorf(msg string, args ...any) {
	l.stdOutHandler.Errorf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
}

func (l *Logger) Warnf(msg string, args ...any) {
	l.stdOutHandler.Warnf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warnf(msg, args...)
	}
}

func (l *Logger) FatalErr(err error) {
	l.Fatalf(err.Error())
}

func (l *Logger) Fatalf(msg string, args ...any) {
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
	l.stdOutHandler.Fatalf(msg, args...)
}

func (l *Logger) Infox(msg string, kv ...any) {
	l.stdOutHandler.Info(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg, kv...)
	}
}

func (l *Logger) Debugx(msg string, kv ...any) {
	l.stdOutHandler.Debug(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg, kv...)
	}
}

func (l *Logger) Errorx(msg string, kv ...any) {
	l.stdOutHandler.Error(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
}

func (l *Logger) Warnx(msg string, kv ...any) {
	l.stdOutHandler.Warn(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg, kv...)
	}
}

func (l *Logger) Fatalx(msg string, kv ...any) {
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
	l.stdOutHandler.Fatal(msg, kv...)
}

func (l *Logger) PlainTextInfo(msg string) {
	_, _ = fmt.Fprintln(os.Stdout, l.style.RenderInfo(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *Logger) PlainTextSuccess(msg string) {
	_, _ = fmt.Fprintln(os.Stdout, msg)
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *Logger) Close() error {
	if l.archiveFile != nil {
		return l.archiveFile.Close()
	}
	return nil
}
