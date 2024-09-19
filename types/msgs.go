package types

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

var TickTime = time.Millisecond * 250

type TickMsg time.Time
type SubmitMsg struct{}
type ReplaceViewMsg struct{}

func Tick() tea.Msg {
	return tea.Tick(TickTime, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func Submit() tea.Msg {
	return SubmitMsg{}
}

func ReplaceView() tea.Msg {
	return ReplaceViewMsg{}
}

type RenderState struct {
	Width         int
	Height        int
	ContentWidth  int
	ContentHeight int
	Theme         *styles.Theme
}
