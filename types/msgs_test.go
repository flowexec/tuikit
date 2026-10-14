package types_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit/types"
)

func TestOffsetMouse(t *testing.T) {
	m := tea.Mouse{X: 10, Y: 8, Button: tea.MouseLeft}
	for _, msg := range []tea.MouseMsg{
		tea.MouseClickMsg(m),
		tea.MouseReleaseMsg(m),
		tea.MouseWheelMsg(m),
		tea.MouseMotionMsg(m),
	} {
		got, ok := types.OffsetMouse(msg, 3, 2).(tea.MouseMsg)
		if !ok {
			t.Fatalf("expected %T to stay a mouse message", msg)
		}
		if got.Mouse().X != 7 || got.Mouse().Y != 6 {
			t.Errorf("expected %T at (7, 6), got (%d, %d)", msg, got.Mouse().X, got.Mouse().Y)
		}
	}

	key := tea.KeyPressMsg{Code: tea.KeyEnter}
	if got := types.OffsetMouse(key, 3, 2); got != key {
		t.Errorf("expected non-mouse message unchanged, got %v", got)
	}
}
