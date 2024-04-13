package io

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"

	"github.com/jahvon/tuikit/styles"
)

type StandardLogger struct {
	stdOutHandler  *log.Logger
	archiveHandler *log.Logger
	style          styles.Theme
	archiveDir     string
	archiveFile    *os.File
}

func NewLogger(style styles.Theme, archiveDir string) *StandardLogger {
	logger := &StandardLogger{style: style}
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
		archiveFile := NewArchiveLogFile(archiveDir)
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
		RotateArchive(logger)
	}

	return logger
}

func applyHumanReadableFormat(handler *log.Logger, style styles.Theme) {
	if style.UsePlainTextLogger {
		handler.SetFormatter(log.TextFormatter)
	} else {
		handler.SetFormatter(log.LogfmtFormatter)
		handler.SetTimeFormat(time.Kitchen)
	}
	handler.SetStyles(style.LoggerStyles())
	handler.SetColorProfile(termenv.ColorProfile())
}

func applyStorageFormat(handler *log.Logger) {
	handler.SetFormatter(log.JSONFormatter)
	handler.SetTimeFormat(time.RFC822)
	handler.SetStyles(log.DefaultStyles())
}

// SetLevel sets the log level for the logger.
// -1 = Fatal
// 0 = Info
// 1 = Debug
// Default is Info.
func (l *StandardLogger) SetLevel(level int) {
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

func (l *StandardLogger) Println(data string) {
	_, err := fmt.Fprintln(os.Stdout, data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, data)
	}
}

func (l *StandardLogger) AsPlainText(exec func()) {
	if l.style.UsePlainTextLogger {
		exec()
		return
	}

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

func (l *StandardLogger) AsJSON(exec func()) {
	if !l.style.UsePlainTextLogger {
		exec()
		return
	}

	l.stdOutHandler.SetFormatter(log.JSONFormatter)
	if l.archiveHandler != nil {
		l.archiveHandler.SetFormatter(log.JSONFormatter)
	}

	exec()

	l.stdOutHandler.SetFormatter(log.LogfmtFormatter)
	if l.archiveHandler != nil {
		l.archiveHandler.SetFormatter(log.LogfmtFormatter)
	}
}

func (l *StandardLogger) Infof(msg string, args ...any) {
	l.stdOutHandler.Infof(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Infof(msg, args...)
	}
}

func (l *StandardLogger) Debugf(msg string, args ...any) {
	l.stdOutHandler.Debugf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debugf(msg, args...)
	}
}

func (l *StandardLogger) Error(err error, msg string) {
	if msg == "" {
		l.Errorf(err.Error())
		return
	}
	l.Errorx(err.Error(), "err", err)
}

func (l *StandardLogger) Errorf(msg string, args ...any) {
	l.stdOutHandler.Errorf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
}

func (l *StandardLogger) Warnf(msg string, args ...any) {
	l.stdOutHandler.Warnf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warnf(msg, args...)
	}
}

func (l *StandardLogger) FatalErr(err error) {
	l.Fatalf(err.Error())
}

func (l *StandardLogger) Fatalf(msg string, args ...any) {
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
	l.stdOutHandler.Fatalf(msg, args...)
}

func (l *StandardLogger) Infox(msg string, kv ...any) {
	l.stdOutHandler.Info(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg, kv...)
	}
}

func (l *StandardLogger) Debugx(msg string, kv ...any) {
	l.stdOutHandler.Debug(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg, kv...)
	}
}

func (l *StandardLogger) Errorx(msg string, kv ...any) {
	l.stdOutHandler.Error(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
}

func (l *StandardLogger) Warnx(msg string, kv ...any) {
	l.stdOutHandler.Warn(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg, kv...)
	}
}

func (l *StandardLogger) Fatalx(msg string, kv ...any) {
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
	l.stdOutHandler.Fatal(msg, kv...)
}

func (l *StandardLogger) PlainTextInfo(msg string) {
	_, _ = fmt.Fprintln(os.Stdout, l.style.RenderInfo(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextSuccess(msg string) {
	_, _ = fmt.Fprintln(os.Stdout, l.style.RenderSuccess(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) Flush() error {
	if l.archiveFile != nil { //nolint:nestif
		if err := l.archiveFile.Sync(); err != nil {
			return err
		}
		if err := l.archiveFile.Close(); err != nil {
			return err
		}
		if info, err := os.Stat(l.archiveFile.Name()); err == nil {
			if info.Size() == 0 {
				_ = os.Remove(l.archiveFile.Name())
			}
		}
	}
	return nil
}

//go:generate mockgen -destination=mocks/mock_logger.go -package=mocks . Logger
type Logger interface {
	Flush() error
	SetLevel(level int)

	PlainTextInfo(msg string)
	PlainTextSuccess(msg string)
	AsPlainText(exec func())
	AsJSON(exec func())

	Infof(msg string, args ...any)
	Debugf(msg string, args ...any)
	Error(err error, msg string)
	Errorf(msg string, args ...any)
	Warnf(msg string, args ...any)
	Fatalf(msg string, args ...any)

	Infox(msg string, kv ...any)
	Debugx(msg string, kv ...any)
	Errorx(msg string, kv ...any)
	Warnx(msg string, kv ...any)
	Fatalx(msg string, kv ...any)

	Println(data string)
	FatalErr(err error)
}
