package io

import (
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/muesli/termenv"

	"github.com/flowexec/tuikit/themes"
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
	archiveID      string
	archiveFile    *os.File
	outFile        *os.File
	exitFunc       func(msg string, args ...any)
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
		logger.archiveDir = path
	}
}

func WithArchiveID(id string) LoggerOptions {
	return func(logger *StandardLogger) {
		logger.archiveID = id
	}
}

func setupArchive(logger *StandardLogger) {
	archiveFile := NewArchiveLogFile(logger.archiveDir, logger.archiveID)
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
}

func WithExitFunc(exit func(msg string, args ...any)) LoggerOptions {
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

	if logger.archiveDir != "" {
		setupArchive(logger)
		RotateArchive(logger.archiveDir)
	}

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
	handler.SetColorProfile(colorprofile.Profile(termenv.ColorProfile()))
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
		l.PlainTextInfo(safeSprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Info(safeSprintf(msg, args...))
		if l.archiveHandler != nil {
			l.archiveHandler.Info(safeSprintf(msg, args...))
		}
	}
}

func (l *StandardLogger) Noticef(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextNotice(safeSprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.With().Log(themes.LogNoticeLevel, safeSprintf(msg, args...))
		if l.archiveHandler != nil {
			l.archiveHandler.Error(safeSprintf(msg, args...))
		}
	}
}

func (l *StandardLogger) Debugf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextDebug(safeSprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Debug(safeSprintf(msg, args...))
		if l.archiveHandler != nil {
			l.archiveHandler.Debug(safeSprintf(msg, args...))
		}
	}
}

func (l *StandardLogger) WrapError(err error, msg string) {
	if msg == "" {
		l.Error(err.Error())
		return
	} else if l.mode == Hidden {
		return
	}
	l.Error(err.Error(), "err", err)
}

func (l *StandardLogger) Errorf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextError(safeSprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Error(safeSprintf(msg, args...))
		if l.archiveHandler != nil {
			l.archiveHandler.Error(safeSprintf(msg, args...))
		}
	}
}

func (l *StandardLogger) Warnf(msg string, args ...any) {
	l.syncLoggerFormat()
	switch l.mode {
	case Text:
		l.PlainTextWarn(safeSprintf(msg, args...))
		return
	case Hidden:
		return
	case JSON, Logfmt:
		l.outHandler.Warn(safeSprintf(msg, args...))
		if l.archiveHandler != nil {
			l.archiveHandler.Warn(safeSprintf(msg, args...))
		}
	}
}

func (l *StandardLogger) FatalErr(err error) {
	l.Fatalf("%s", err.Error())
}

func (l *StandardLogger) Fatalf(msg string, args ...any) {
	l.syncLoggerFormat()
	formatted := safeSprintf(msg, args...)
	switch l.mode {
	case Text:
		l.PlainTextError(formatted)
		l.exitFunc(formatted)
		return
	case Hidden:
		return
	case JSON, Logfmt:
		if l.archiveHandler != nil {
			l.archiveHandler.Error(formatted)
		}
		l.outHandler.Fatal(formatted)
	}
}

func (l *StandardLogger) Info(msg string, kv ...any) {
	l.syncLoggerFormat()
	if l.mode == Hidden {
		return
	}
	l.outHandler.Info(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg, kv...)
	}
}

func (l *StandardLogger) Notice(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.With().Log(themes.LogNoticeLevel, msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Errorf(msg, kv...)
	}
}

func (l *StandardLogger) Debug(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Debug(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg, kv...)
	}
}

func (l *StandardLogger) Error(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Error(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg, kv...)
	}
}

func (l *StandardLogger) Warn(msg string, kv ...any) {
	if l.mode == Hidden {
		return
	}
	l.syncLoggerFormat()
	l.outHandler.Warn(msg, kv...)
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg, kv...)
	}
}

func (l *StandardLogger) Fatal(msg string, kv ...any) {
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
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg)
	}
}

func (l *StandardLogger) PlainTextNotice(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderNotice(msg))
	if l.archiveHandler != nil {
		l.archiveHandler.With().Log(themes.LogNoticeLevel, msg)
	}
}

func (l *StandardLogger) PlainTextSuccess(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderSuccess(msg))
	if l.archiveHandler != nil {
		l.archiveHandler.Info(msg)
	}
}

func (l *StandardLogger) PlainTextError(msg string) {
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderError(msg))
	if l.archiveHandler != nil {
		l.archiveHandler.Error(msg)
	}
}

func (l *StandardLogger) PlainTextWarn(msg string) {
	if l.outHandler.GetLevel() < log.InfoLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderWarning(msg))
	if l.archiveHandler != nil {
		l.archiveHandler.Warn(msg)
	}
}

func (l *StandardLogger) PlainTextDebug(msg string) {
	if l.outHandler.GetLevel() > log.DebugLevel {
		return
	}
	_, _ = fmt.Fprintln(l.outFile, ""+l.theme.RenderEmphasis(msg))
	if l.archiveHandler != nil {
		l.archiveHandler.Debug(msg)
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

// safeSprintf applies fmt.Sprintf only when args are provided.
// When called with no args, it returns msg as-is, avoiding
// misinterpretation of % characters in the message.
func safeSprintf(msg string, args ...any) string {
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func defaultExit(_ string, _ ...any) {
	os.Exit(1)
}

// --- TaskAwareLogger implementation ---

func (l *StandardLogger) PrintWithTask(task *TaskContext, line string) {
	if l.mode == Hidden {
		return
	}
	prefix := taskPrefix(task, l.theme.ColorPalette())
	formatted := fmt.Sprintf("%s %s", prefix, line)
	_, _ = fmt.Fprintln(l.outFile, formatted)

	if l.archiveHandler != nil {
		l.archiveHandler.Info(line, "task", task.Name)
	}
}

func (l *StandardLogger) PrintErrWithTask(task *TaskContext, line string) {
	if l.mode == Hidden {
		return
	}
	prefix := taskPrefix(task, l.theme.ColorPalette())
	errLine := l.theme.RenderError(line)
	formatted := fmt.Sprintf("%s %s", prefix, errLine)
	_, _ = fmt.Fprintln(l.outFile, formatted)

	if l.archiveHandler != nil {
		l.archiveHandler.Error(line, "task", task.Name)
	}
}

func (l *StandardLogger) PrintTaskSummary(tasks []*TaskContext) {
	if l.mode == Hidden || len(tasks) == 0 {
		return
	}

	_, _ = fmt.Fprintln(l.outFile)

	palette := l.theme.ColorPalette()
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(palette.Primary))
	_, _ = fmt.Fprintln(l.outFile, headerStyle.Render("Task Summary"))

	maxLen := 0
	for _, t := range tasks {
		if len(t.Name) > maxLen {
			maxLen = len(t.Name)
		}
	}

	for _, t := range tasks {
		var statusColor string
		switch t.Status {
		case TaskSuccess:
			statusColor = palette.Success
		case TaskFailed:
			statusColor = palette.Error
		case TaskSkipped:
			statusColor = palette.Warning
		case TaskRunning:
			statusColor = palette.Info
		}

		nameStyle := lipgloss.NewStyle().Width(maxLen + 2)
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true)
		durationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(palette.Gray))

		line := fmt.Sprintf("  %s %s %s",
			nameStyle.Render(t.Name),
			statusStyle.Render(fmt.Sprintf("%-4s", t.Status.Icon())),
			durationStyle.Render(t.Duration().Truncate(time.Millisecond).String()),
		)
		_, _ = fmt.Fprintln(l.outFile, line)
	}
	_, _ = fmt.Fprintln(l.outFile)

	if l.archiveHandler != nil {
		for _, t := range tasks {
			l.archiveHandler.Info("task_summary",
				"task", t.Name,
				"status", t.Status.String(),
				"duration", t.Duration().String(),
			)
		}
	}
}

func (l *StandardLogger) BeginGroup(name string) {
	if isCI() {
		_, _ = fmt.Fprintf(l.outFile, "::group::%s\n", name)
	} else {
		palette := l.theme.ColorPalette()
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Secondary)).
			Bold(true)
		_, _ = fmt.Fprintf(l.outFile, "\n%s\n", style.Render("--- "+name+" ---"))
	}
}

func (l *StandardLogger) EndGroup() {
	if isCI() {
		_, _ = fmt.Fprintln(l.outFile, "::endgroup::")
	}
}
