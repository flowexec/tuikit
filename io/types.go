package io

import "fmt"

type LogMode string

const (
	Hidden LogMode = "hidden"
	Text   LogMode = "text"
	Logfmt LogMode = "logfmt"
	JSON   LogMode = "json"
)

func (m LogMode) Validate() error {
	switch m {
	case Hidden, Text, Logfmt, JSON, "":
		return nil
	default:
		return fmt.Errorf("invalid log mode %s", m)
	}
}

//go:generate mockgen -destination=mocks/mock_logger.go -package=mocks . Logger
type Logger interface {
	Flush() error
	SetLevel(level int)
	SetMode(mode LogMode)
	LogMode() LogMode

	PlainTextInfo(msg string)
	PlainTextSuccess(msg string)
	PlainTextError(msg string)
	PlainTextDebug(msg string)
	PlainTextWarn(msg string)

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
