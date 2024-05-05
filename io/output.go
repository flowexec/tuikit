package io

import (
	"fmt"
	"strings"
)

type StdOutWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   LogMode
}

func (w StdOutWriter) Write(p []byte) (n int, err error) {
	if strings.TrimSpace(string(p)) == "" {
		return len(p), nil
	}
	splitP := strings.Split(string(p), "\n")

	curMode := w.Logger.LogMode()
	if w.LogMode != curMode || w.LogMode == "" {
		w.Logger.SetMode(w.LogMode)
	}
	for _, line := range splitP {
		switch w.Logger.LogMode() {
		case Hidden:
			return len(p), nil
		case Text:
			w.Logger.Println(line)
		case JSON, Logfmt:
			if len(w.LogFields) > 0 {
				w.Logger.Infox(line, w.LogFields...)
			} else {
				w.Logger.Infof(line)
			}
		default:
			return len(p), fmt.Errorf("unknown log mode %s", w.LogMode)
		}
	}
	if w.LogMode != curMode {
		w.Logger.SetMode(curMode)
	}

	return len(p), nil
}

type StdErrWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   LogMode
}

func (w StdErrWriter) Write(p []byte) (n int, err error) {
	trimmedP := strings.TrimSpace(string(p))
	if trimmedP == "" {
		return len(p), nil
	}

	curMode := w.Logger.LogMode()
	if w.LogMode != curMode || w.LogMode == "" {
		w.Logger.SetMode(w.LogMode)
	}
	switch w.Logger.LogMode() {
	case Hidden:
		return len(p), nil
	case Text:
		w.Logger.PlainTextError(trimmedP)
	case JSON, Logfmt:
		if len(w.LogFields) > 0 {
			w.Logger.Errorx(trimmedP, w.LogFields...)
		} else {
			w.Logger.Errorf(trimmedP)
		}
	default:
		return len(p), fmt.Errorf("unknown log mode %s", w.LogMode)
	}
	if w.LogMode != curMode {
		w.Logger.SetMode(curMode)
	}

	return len(p), nil
}
