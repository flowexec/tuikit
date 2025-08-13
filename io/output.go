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
	case JSON, Logfmt:
		if strings.TrimSpace(string(p)) == "" {
			return len(p), nil
		}
		splitP := strings.Split(string(p), "\n")
		for _, line := range splitP {
			if strings.TrimSpace(line) == "" {
				continue
			}
			w.Logger.Info(line, w.LogFields...)
		}
	default:
		return len(p), fmt.Errorf("unknown log mode %v", curMode)
	}

	return len(p), nil
}

type StdErrWriter struct {
	LogFields []any
	Logger    Logger
	LogMode   *LogMode
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
	case JSON, Logfmt:
		if strings.TrimSpace(string(p)) == "" {
			return len(p), nil
		}
		splitP := strings.Split(string(p), "\n")
		for _, line := range splitP {
			if strings.TrimSpace(line) == "" {
				continue
			}
			w.Logger.Notice(line, w.LogFields...)
		}
	default:
		return len(p), fmt.Errorf("unknown log mode %v", w.LogMode)
	}

	return len(p), nil
}
