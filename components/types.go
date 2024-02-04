package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

type TickMsg time.Time
type NoticeLevel string
type Format string

const (
	NoticeLevelInfo    NoticeLevel = "info"
	NoticeLevelWarning NoticeLevel = "warning"
	NoticeLevelError   NoticeLevel = "error"

	FormatDocument Format = "doc"
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
)

type TerminalState struct {
	Width  int
	Height int
	Theme  styles.Theme
}

type KeyCallback struct {
	Key      string
	Label    string
	Callback func() error
}

type TeaModel interface {
	tea.Model

	Interactive() bool
	HelpMsg() string
	Type() string
}
