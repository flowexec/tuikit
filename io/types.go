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
	PlainTextNotice(msg string)
	PlainTextSuccess(msg string)
	PlainTextError(msg string)
	PlainTextDebug(msg string)
	PlainTextWarn(msg string)

	Infof(msg string, args ...any)
	Noticef(msg string, args ...any)
	Debugf(msg string, args ...any)
	WrapError(err error, msg string)
	Errorf(msg string, args ...any)
	Warnf(msg string, args ...any)
	Fatalf(msg string, args ...any)

	Info(msg string, kv ...any)
	Notice(msg string, kv ...any)
	Debug(msg string, kv ...any)
	Error(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Fatal(msg string, kv ...any)

	Print(data string)
	Println(data string)
	FatalErr(err error)
}
