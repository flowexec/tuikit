package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

type TickMsg time.Time
type SubmitMsgType struct{}
type NoticeLevel string
type Format string

const (
	FormatDocument Format = "doc"
	FormatList     Format = "list"
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

func SubmitMsg() tea.Msg {
	return SubmitMsgType{}
}
