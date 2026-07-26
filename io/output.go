package io

import (
	"fmt"
	"strings"
)

// serializedLogger is implemented by loggers (e.g. StandardLogger) that can serialize the
// writers' mode flip/render/restore sequence. When available, Std{Out,Err}Writer hold the
// lock for the whole Write so a command's concurrently-copied stdout and stderr streams
// cannot interleave and render each other's output in the wrong mode.
type serializedLogger interface {
	acquireWriteLock()
	releaseWriteLock()
}

func serializeWrite(logger Logger) func() {
	if sl, ok := logger.(serializedLogger); ok {
		sl.acquireWriteLock()
		return sl.releaseWriteLock
	}
	return func() {}
}

type StdOutWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
	Task      *TaskContext
}

func (w StdOutWriter) Write(p []byte) (n int, err error) {
	defer serializeWrite(w.Logger)()

	curMode := w.Logger.LogMode()
	flipped := w.LogMode != nil && *w.LogMode != "" && *w.LogMode != curMode
	if flipped {
		w.Logger.SetMode(*w.LogMode)
	}
	defer func() {
		if flipped {
			w.Logger.SetMode(curMode)
		}
	}()

	switch w.Logger.LogMode() {
	case Hidden:
		return len(p), nil
	case Text:
		w.Logger.Print(string(p))
		return len(p), nil
	case Logfmt:
		if w.Task != nil {
			if tal, ok := w.Logger.(TaskAwareLogger); ok {
				writeLines(p, func(line string) { tal.PrintWithTask(w.Task, line) })
				return len(p), nil
			}
		}
		writeLines(p, func(line string) { w.Logger.Info(line, w.LogFields...) })
	case JSON:
		writeLines(p, func(line string) { w.Logger.Info(line, w.LogFields...) })
	default:
		return len(p), fmt.Errorf("unknown log mode %v", curMode)
	}

	return len(p), nil
}

type StdErrWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
	Task      *TaskContext
}

func (w StdErrWriter) Write(p []byte) (n int, err error) {
	defer serializeWrite(w.Logger)()

	curMode := w.Logger.LogMode()
	flipped := w.LogMode != nil && *w.LogMode != "" && *w.LogMode != curMode
	if flipped {
		w.Logger.SetMode(*w.LogMode)
	}
	defer func() {
		if flipped {
			w.Logger.SetMode(curMode)
		}
	}()

	switch w.Logger.LogMode() {
	case Hidden:
		return len(p), nil
	case Text:
		w.Logger.Print(string(p))
		return len(p), nil
	case Logfmt:
		if w.Task != nil {
			if tal, ok := w.Logger.(TaskAwareLogger); ok {
				writeLines(p, func(line string) { tal.PrintErrWithTask(w.Task, line) })
				return len(p), nil
			}
		}
		writeLines(p, func(line string) { w.Logger.Notice(line, w.LogFields...) })
	case JSON:
		writeLines(p, func(line string) { w.Logger.Notice(line, w.LogFields...) })
	default:
		return len(p), fmt.Errorf("unknown log mode %v", w.LogMode)
	}

	return len(p), nil
}

// writeLines splits p into non-empty lines and calls fn for each.
func writeLines(p []byte, fn func(string)) {
	s := string(p)
	if strings.TrimSpace(s) == "" {
		return
	}
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fn(line)
	}
}
