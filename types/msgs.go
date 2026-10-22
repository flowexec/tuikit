package types

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit/themes"
)

var TickTime = time.Millisecond * 250

// Key name strings returned by KeyPressMsg.String().
const (
	KeyUp        = "up"
	KeyDown      = "down"
	KeyEnter     = "enter"
	KeyBackspace = "backspace"
	KeyEsc       = "esc"
	KeyCopy      = "c"
)

type TickMsg time.Time
type SubmitMsg struct{}
type ReplaceViewMsg struct{}

type ToastDismissMsg struct {
	ID int
}

// ToastMsg asks the Container to show a toast notification.
type ToastMsg struct {
	Text  string
	Level themes.OutputLevel
}

// OffsetMouse returns msg with its coordinates shifted left by dx and up by dy,
// translating them into the coordinate space of a nested view.
// Messages that are not mouse events are returned unchanged.
func OffsetMouse(msg tea.Msg, dx, dy int) tea.Msg {
	switch m := msg.(type) {
	case tea.MouseClickMsg:
		m.X, m.Y = m.X-dx, m.Y-dy
		return m
	case tea.MouseReleaseMsg:
		m.X, m.Y = m.X-dx, m.Y-dy
		return m
	case tea.MouseWheelMsg:
		m.X, m.Y = m.X-dx, m.Y-dy
		return m
	case tea.MouseMotionMsg:
		m.X, m.Y = m.X-dx, m.Y-dy
		return m
	}
	return msg
}

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
	Theme         themes.Theme
}
