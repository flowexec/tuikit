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

//nolint:dupl // this is a slightly modified mirror of StdErrWriter
func (w StdOutWriter) Write(p []byte) (n int, err error) {
	if strings.TrimSpace(string(p)) == "" {
		return len(p), nil
	}
	splitP := strings.Split(string(p), "\n")

	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
	}
	for _, line := range splitP {
		switch w.Logger.LogMode() {
		case Hidden:
			return len(p), nil
		case Text:
			w.Logger.Println(string(p))
		case JSON, Logfmt:
			if strings.TrimSpace(line) == "" {
				continue
			}
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

//nolint:dupl // this is a slightly modified mirror of StdOutWriter
func (w StdErrWriter) Write(p []byte) (n int, err error) {
	if strings.TrimSpace(string(p)) == "" {
		return len(p), nil
	}
	splitP := strings.Split(string(p), "\n")

	curMode := w.Logger.LogMode()
	if w.LogMode != nil && (*w.LogMode != "" && *w.LogMode != curMode) {
		w.Logger.SetMode(*w.LogMode)
	}
	for _, line := range splitP {
		switch w.Logger.LogMode() {
		case Hidden:
			return len(p), nil
		case Text:
			w.Logger.PlainTextNotice(string(p))
		case JSON, Logfmt:
			if strings.TrimSpace(line) == "" {
				continue
			}
			if len(w.LogFields) > 0 {
				w.Logger.Noticex(line, w.LogFields...)
			} else {
				w.Logger.Noticef(line)
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
