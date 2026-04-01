package overlay

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

const (
	defaultMaxVisible = 3
	defaultTimeout    = 3 * time.Second
)

type toast struct {
	id    int
	text  string
	level themes.OutputLevel
}

// ToastManager manages a queue of auto-dismissing toast notifications.
type ToastManager struct {
	queue      []toast
	nextID     int
	theme      themes.Theme
	maxVisible int
	timeout    time.Duration
}

// NewToastManager creates a new ToastManager with sensible defaults.
func NewToastManager(theme themes.Theme) *ToastManager {
	return &ToastManager{
		theme:      theme,
		maxVisible: defaultMaxVisible,
		timeout:    defaultTimeout,
	}
}

// Push adds a toast and returns a tea.Cmd that will fire a ToastDismissMsg after the timeout.
func (tm *ToastManager) Push(text string, lvl themes.OutputLevel) tea.Cmd {
	id := tm.nextID
	tm.nextID++
	tm.queue = append(tm.queue, toast{id: id, text: text, level: lvl})

	timeout := tm.timeout
	return tea.Tick(timeout, func(_ time.Time) tea.Msg {
		return types.ToastDismissMsg{ID: id}
	})
}

// Dismiss removes a toast by ID.
func (tm *ToastManager) Dismiss(id int) {
	for i, t := range tm.queue {
		if t.id == id {
			tm.queue = append(tm.queue[:i], tm.queue[i+1:]...)
			return
		}
	}
}

// Empty returns true if there are no active toasts.
func (tm *ToastManager) Empty() bool {
	return len(tm.queue) == 0
}

// Render produces the stacked toast string for overlay composition.
func (tm *ToastManager) Render(width, _ int) string {
	if len(tm.queue) == 0 {
		return ""
	}

	visible := tm.queue
	if len(visible) > tm.maxVisible {
		visible = visible[len(visible)-tm.maxVisible:]
	}

	rendered := make([]string, len(visible))
	for i, t := range visible {
		rendered[i] = tm.theme.RenderToast(t.text, t.level, width)
	}

	return strings.Join(rendered, "\n")
}
