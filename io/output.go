package io

import (
	"fmt"
	"strings"
)

type StdOutWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
}

func (w StdOutWriter) Write(p []byte) (n int, err error) {
	if strings.TrimSpace(string(p)) == "" {
		return len(p), nil
	}
	endsWithNewline := strings.HasSuffix(string(p), "\n")
	splitP := strings.Split(string(p), "\n")

	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
	}
	for i, line := range splitP {
		switch w.Logger.LogMode() {
		case Hidden:
			return len(p), nil
		case Text:
			// Maintain the newline at the end of the log message
			if i+1 == len(splitP) && endsWithNewline {
				w.Logger.Println(line)
			} else {
				w.Logger.Print(line)
			}
		case JSON, Logfmt:
			if len(w.LogFields) > 0 {
				w.Logger.Infox(line, w.LogFields...)
			} else {
				w.Logger.Infof(line)
			}
		default:
			return len(p), fmt.Errorf("unknown log mode %v", w.LogMode)
		}
	}
	if w.LogMode != nil && *w.LogMode != curMode {
		w.Logger.SetMode(curMode)
	}

	return len(p), nil
}

type StdErrWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
}

func (w StdErrWriter) Write(p []byte) (n int, err error) {
	if strings.TrimSpace(string(p)) == "" {
		return len(p), nil
	}
	endsWithNewline := strings.HasSuffix(string(p), "\n")
	splitP := strings.Split(string(p), "\n")

	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
	}
	for i, line := range splitP {
		switch w.Logger.LogMode() {
		case Hidden:
			return len(p), nil
		case Text:
			// Maintain the newline at the end of the log message
			if i == len(splitP)-1 && endsWithNewline {
				w.Logger.PlainTextError(line + "\n")
			} else {
				w.Logger.PlainTextError(line)
			}
		case JSON, Logfmt:
			if len(w.LogFields) > 0 {
				w.Logger.Errorx(line, w.LogFields...)
			} else {
				w.Logger.Errorf(line)
			}
		default:
			return len(p), fmt.Errorf("unknown log mode %v", w.LogMode)
		}
	}
	if w.LogMode != nil && *w.LogMode != curMode {
		w.Logger.SetMode(curMode)
	}

	return len(p), nil
}
