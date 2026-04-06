package io

import (
	"fmt"
	"strings"
)

type StdOutWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
	Task      *TaskContext
}

func (w StdOutWriter) Write(p []byte) (n int, err error) {
	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
		curMode = w.Logger.LogMode()
	}
	defer func() {
		if w.LogMode != nil && *w.LogMode != curMode {
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
	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
	}
	defer func() {
		if w.LogMode != nil && *w.LogMode != curMode {
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
