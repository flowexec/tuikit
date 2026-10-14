package views

import (
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

// wheelDelta returns -1 for a wheel-up event, 1 for wheel-down, and 0 otherwise.
func wheelDelta(msg tea.MouseWheelMsg) int {
	switch msg.Button {
	case tea.MouseWheelUp:
		return -1
	case tea.MouseWheelDown:
		return 1
	default:
		return 0
	}
}

func scrollList(m *list.Model, delta int) {
	switch {
	case delta < 0:
		m.CursorUp()
	case delta > 0:
		m.CursorDown()
	}
}
