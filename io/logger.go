package io

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"

	"github.com/jahvon/tuikit/themes"
)

var (
	defaultMode   = Text
	defaultTheme  = themes.EverforestTheme()
	defaultOutput = os.Stdout
)

type StandardLogger struct {
	outHandler     *log.Logger
	archiveHandler *log.Logger
	theme          themes.Theme
	mode           LogMode
	archiveDir     string
	archiveFile    *os.File
	outFile        *os.File
	exitFunc       func()
}

type LoggerOptions func(*StandardLogger)

func WithTheme(theme themes.Theme) LoggerOptions {
	return func(logger *StandardLogger) {
		logger.theme = theme
	}
}

func WithMode(mode LogMode) LoggerOptions {
	return func(logger *StandardLogger) {
		logger.mode = mode
	}
}

func WithOutput(file *os.File) LoggerOptions {
	return func(logger *StandardLogger) {
		logger.outFile = file
	}
}

func WithArchiveDirectory(path string) LoggerOptions {
	return func(logger *StandardLogger) {
		if path == "" {
			return
		}
		logger.archiveDir = path
		archiveFile := NewArchiveLogFile(path)
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
}

func WithExitFunc(exit func()) LoggerOptions {
	return func(logger *StandardLogger) {
		logger.exitFunc = exit
	}
}

// NewLogger creates a new instance of StandardLogger with the provided functional options.
//
// Functional options allow you to customize the behavior of the logger:
//   - WithTheme(theme styles.Theme): Sets the theme for the logger's output.
//   - WithMode(mode LogMode): Configures the logging mode (e.g., text or JSON).
//   - WithOutput(file *os.File): Specifies the output file for the logger.
//   - WithArchiveDirectory(path string): Enables log archiving to the specified directory.
//     If the path is empty, archiving is disabled.
//   - WithExitFunc(exit func()): Sets a custom function to be called on logger exit.
//
// By default, the logger uses a standard theme, text mode, and writes to os.Stdout.
func NewLogger(opts ...LoggerOptions) *StandardLogger {
	logger := &StandardLogger{
		theme:    defaultTheme,
		mode:     defaultMode,
		outFile:  defaultOutput,
		exitFunc: defaultExit,
	}
	for _, opt := range opts {
		opt(logger)
	}
	stdOutHandler := log.NewWithOptions(logger.outFile, log.Options{Level: log.InfoLevel, ReportCaller: false})
	applyHumanReadableFormat(stdOutHandler, logger.theme, logger.mode)
	logger.outHandler = stdOutHandler

	return logger
}

func (l *StandardLogger) SetMode(mode LogMode) {
	if mode == "" {
		return
	}
	l.mode = mode
	applyHumanReadableFormat(l.outHandler, l.theme, mode)
}

func (l *StandardLogger) LogMode() LogMode {
	if l.mode == "" {
		return Text
	}
	return l.mode
}

func applyHumanReadableFormat(handler *log.Logger, style themes.Theme, mode LogMode) {
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
		l.outHandler.SetLevel(log.FatalLevel)
	case 0:
		l.outHandler.SetLevel(log.InfoLevel)
	case 1:
		l.outHandler.SetLevel(log.DebugLevel)
	default:
		l.outHandler.SetLevel(log.InfoLevel)
	}
}

func (l *StandardLogger) Print(data string) {
	_, err := fmt.Fprint(l.outFile, ""+data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprint(l.archiveFile, data)
	}
}

func (l *StandardLogger) Println(data string) {
	_, err := fmt.Fprintln(l.outFile, ""+data)
	if err != nil {
		panic(err)
	}
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, data)
	}
}

func (l *StandardLogger) Infof(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextInfo(fmt.Sprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Infof(msg, args...)
		if l.archiveHandler != nil {
			l.archiveHandler.Infof(msg, args...)
		}
	}
}

func (l *StandardLogger) Noticef(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextNotice(fmt.Sprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.With().Log(themes.LogNoticeLevel, msg, args...)
		if l.archiveHandler != nil {
			l.archiveHandler.Errorf(msg, args...)
		}
	}
}

func (l *StandardLogger) Debugf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextDebug(fmt.Sprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Debugf(msg, args...)
		if l.archiveHandler != nil {
			l.archiveHandler.Debugf(msg, args...)
		}
	}
}

func (l *StandardLogger) Error(err error, msg string) {
	if msg == "" {
		l.Errorf(err.Error())
		return
	} else if l.mode == Hidden {
		return
	}
	l.Errorx(err.Error(), "err", err)
}

func (l *StandardLogger) Errorf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextError(fmt.Sprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Errorf(msg, args...)
		if l.archiveHandler != nil {
			l.archiveHandler.Errorf(msg, args...)
		}
	}
}

func (l *StandardLogger) Warnf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextWarn(fmt.Sprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Warnf(msg, args...)
		if l.archiveHandler != nil {
			l.archiveHandler.Warnf(msg, args...)
		}
	}
}

func (l *StandardLogger) FatalErr(err error) {
	l.Fatalf(err.Error())
}

func (l *StandardLogger) Fatalf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextError(fmt.Sprintf(msg, args...))
		l.exitFunc()
		return
	case Hidden:
		return
	case JSON, Logfmt:
		if l.archiveHandler != nil {
			l.archiveHandler.Errorf(msg, args...)
		}
		l.outHandler.Fatalf(msg, args...)
	}
}

func (l *StandardLogger) Infox(msg string, kv ...any) {
	l.syncLoggerFormat()
	if l.mode == Hidden {
		return
	}
	l.outHandler.Info(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg, kv...)
	}
}

func (l *StandardLogger) Noticex(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.With().Log(themes.LogNoticeLevel, msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, kv...)
	}
}

func (l *StandardLogger) Debugx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Debug(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg, kv...)
	}
}

func (l *StandardLogger) Errorx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Error(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
}

func (l *StandardLogger) Warnx(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Warn(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg, kv...)
	}
}

func (l *StandardLogger) Fatalx(msg string, kv ...any) {
	l.syncLoggerFormat()
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
	l.outHandler.Fatal(msg, kv...)
}

func (l *StandardLogger) PlainTextInfo(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderInfo(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextNotice(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderNotice(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextSuccess(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderSuccess(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextError(msg string) {
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderError(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextWarn(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderWarning(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) PlainTextDebug(msg string) {
	if l.outHandler.GetLevel() > log.DebugLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderEmphasis(msg))
	if l.archiveFile != nil {
		_, _ = fmt.Fprintln(l.archiveFile, msg)
	}
}

func (l *StandardLogger) Flush() error {
	if l.archiveFile == nil {
		return nil
	}

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
	return nil
}

func (l *StandardLogger) syncLoggerFormat() {
	switch l.mode {
	case JSON:
		l.outHandler.SetFormatter(log.JSONFormatter)
	case Logfmt, Text, "":
		l.outHandler.SetFormatter(log.TextFormatter)
	case Hidden:
		return
	}
}

func defaultExit() {
	os.Exit(1)
}
