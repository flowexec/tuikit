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
	mode           LogMode
	archiveDir     string
	archiveFile    *os.File
	stdOutFile     *os.File
}

func NewLogger(stdOut *os.File, style styles.Theme, mode LogMode, archiveDir string) *StandardLogger {
	logger := &StandardLogger{style: style, mode: mode, archiveDir: archiveDir, stdOutFile: stdOut}
	stdOutHandler := log.NewWithOptions(stdOut, log.Options{Level: log.InfoLevel, ReportCaller: false})
	applyHumanReadableFormat(stdOutHandler, style, mode)
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

func (l *StandardLogger) SetMode(mode LogMode) {
	if mode == "" {
		return
	}
	l.mode = mode
	applyHumanReadableFormat(l.stdOutHandler, l.style, mode)
}

func (l *StandardLogger) LogMode() LogMode {
	if l.mode == "" {
		return Text
	}
	return l.mode
}

func applyHumanReadableFormat(handler *log.Logger, style styles.Theme, mode LogMode) {
	handler.SetReportTimestamp(true)
	if mode == JSON {
		handler.SetFormatter(log.JSONFormatter)
		handler.SetTimeFormat(time.RFC822)
		return
	}

	handler.SetFormatter(log.TextFormatter)
	handler.SetTimeFormat(time.Kitchen)
	handler.SetColorProfile(termenv.ColorProfile())
	handler.SetStyles(style.LoggerStyles())
}

func applyStorageFormat(handler *log.Logger) {
	handler.SetFormatter(log.LogfmtFormatter)
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

func (l *StandardLogger) Print(data string) {
	_, err := fmt.Fprint(l.stdOutFile, ""+data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprint(l.archiveFile, data)
	}
}

func (l *StandardLogger) Println(data string) {
	_, err := fmt.Fprintln(l.stdOutFile, ""+data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, data)
	}
}

func (l *StandardLogger) Infof(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextInfo(fmt.Sprintf(msg, args...))
		return
	} else if l.mode == Hidden {
		return
	}
	l.stdOutHandler.Infof(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Infof(msg, args...)
	}
}

func (l *StandardLogger) Noticef(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextNotice(fmt.Sprintf(msg, args...))
		return
	} else if l.mode == Hidden {
		return
	}
	l.stdOutHandler.With().Log(styles.LogNoticeLevel, msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
}

func (l *StandardLogger) Debugf(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextDebug(fmt.Sprintf(msg, args...))
		return
	} else if l.mode == Hidden {
		return
	}
	l.stdOutHandler.Debugf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debugf(msg, args...)
	}
}

func (l *StandardLogger) Error(err error, msg string) {
	if msg == "" {
		l.Errorf(err.Error()) //nolint:govet
		return
	} else if l.mode == Hidden {
		return
	}
	l.Errorx(err.Error(), "err", err)
}

func (l *StandardLogger) Errorf(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextError(fmt.Sprintf(msg, args...))
		return
	} else if l.mode == Hidden {
		return
	}
	l.stdOutHandler.Errorf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
}

func (l *StandardLogger) Warnf(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextWarn(fmt.Sprintf(msg, args...))
		return
	} else if l.mode == Hidden {
		return
	}
	l.stdOutHandler.Warnf(msg, args...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warnf(msg, args...)
	}
}

func (l *StandardLogger) FatalErr(err error) {
	l.Fatalf(err.Error()) //nolint:govet
}

func (l *StandardLogger) Fatalf(msg string, args ...any) {
	l.syncLoggerFormat()
	if l.mode == Text {
		l.PlainTextError(fmt.Sprintf(msg, args...))
		os.Exit(1)
	} else if l.mode == Hidden {
		return
	}
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, args...)
	}
	l.stdOutHandler.Fatalf(msg, args...)
}

func (l *StandardLogger) Infox(msg string, kv ...any) {
	l.syncLoggerFormat()
	if l.mode == Hidden {
		return
	}
	l.stdOutHandler.Info(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg, kv...)
	}
}

func (l *StandardLogger) Noticex(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.stdOutHandler.With().Log(styles.LogNoticeLevel, msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, kv...)
	}
}

func (l *StandardLogger) Debugx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.stdOutHandler.Debug(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg, kv...)
	}
}

func (l *StandardLogger) Errorx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.stdOutHandler.Error(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
}

func (l *StandardLogger) Warnx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.stdOutHandler.Warn(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg, kv...)
	}
}

func (l *StandardLogger) Fatalx(msg string, kv ...any) {
	l.syncLoggerFormat()
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
	l.stdOutHandler.Fatal(msg, kv...)
}

func (l *StandardLogger) PlainTextInfo(msg string) {
	if l.stdOutHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderInfo(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextNotice(msg string) {
	if l.stdOutHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderNotice(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextSuccess(msg string) {
	if l.stdOutHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderSuccess(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextError(msg string) {
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderError(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextWarn(msg string) {
	if l.stdOutHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderWarning(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextDebug(msg string) {
	if l.stdOutHandler.GetLevel() > log.DebugLevel {
		return
	}
	_, _ = fmt.Fprintln(l.stdOutFile, ""+l.style.RenderEmphasis(msg))
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

func (l *StandardLogger) syncLoggerFormat() {
	switch l.mode {
	case JSON:
		l.stdOutHandler.SetFormatter(log.JSONFormatter)
	case Logfmt, Text, "":
		l.stdOutHandler.SetFormatter(log.TextFormatter)
	case Hidden:
		return
	}
}
