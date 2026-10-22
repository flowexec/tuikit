package views

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/x/ansi"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

// Mouse coordinates received by views are relative to the view's own top-left corner.
// The Container and composite views (e.g. Library) translate them before forwarding.

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

// pressEnter re-dispatches an enter key press so that clicking an already selected item
// behaves exactly like pressing enter, including any interception by parent views.
func pressEnter() tea.Msg {
	return tea.KeyPressMsg{Code: tea.KeyEnter}
}

// writeNativeClipboard writes to the local system clipboard. It is a variable so tests do not
// touch the real clipboard.
var writeNativeClipboard = clipboard.WriteAll

// CopyToClipboard returns a command that copies text to the clipboard and confirms with a toast.
// It writes both the native system clipboard and an OSC 52 escape: some local terminals, such as
// macOS Terminal, ignore OSC 52, while over SSH only OSC 52 reaches the user's machine.
func CopyToClipboard(text string) tea.Cmd {
	if text == "" {
		return nil
	}
	preview := strings.Join(strings.Fields(text), " ")
	preview = ansi.Truncate(preview, 40, "…")
	return tea.Batch(
		tea.SetClipboard(text),
		func() tea.Msg {
			// A failure is expected where there is no native clipboard (e.g. a headless
			// SSH host); the OSC 52 copy still applies there.
			_ = writeNativeClipboard(text)
			return types.ToastMsg{Text: "Copied " + preview, Level: themes.OutputLevelSuccess}
		},
	)
}

// copyHelpKey returns the help binding for the copy key unless a callback overrides it.
func copyHelpKey(callbacks []types.KeyCallback) []themes.HelpKey {
	if hasCallback(callbacks, types.KeyCopy) {
		return nil
	}
	return []themes.HelpKey{{Key: types.KeyCopy, Desc: "copy"}}
}

func hasCallback(callbacks []types.KeyCallback, key string) bool {
	for _, cb := range callbacks {
		if cb.Key == key {
			return true
		}
	}
	return false
}

// styleTopOffset returns the number of lines a style adds above its content.
func styleTopOffset(s lipgloss.Style) int {
	return s.GetMarginTop() + s.GetBorderTopSize() + s.GetPaddingTop()
}

// clickListItem selects the list item at line y of the list's rendered output. Clicking the
// item that is already selected returns a command that activates it as if enter was pressed.
func clickListItem(m *list.Model, delegate list.ItemDelegate, y int) tea.Cmd {
	idx, ok := listItemAt(*m, delegate, y)
	if !ok {
		return nil
	}
	if idx == m.Index() {
		return pressEnter
	}
	m.Select(idx)
	return nil
}

const listMarker = "<<tuikit-item:"

// listItemAt returns the index of the visible item rendered at line y of m.View().
// The list does not expose its layout, so it renders a copy of the model with a delegate
// that writes a marker for each item and finds the marker lines.
func listItemAt(m list.Model, delegate list.ItemDelegate, y int) (int, bool) {
	if y < 0 || len(m.VisibleItems()) == 0 {
		return 0, false
	}
	m.SetDelegate(markerDelegate{height: delegate.Height(), spacing: delegate.Spacing()})
	for i, line := range strings.Split(m.View(), "\n") {
		_, after, found := strings.Cut(line, listMarker)
		if !found {
			continue
		}
		end := strings.IndexByte(after, ';')
		if end < 0 {
			continue
		}
		idx, err := strconv.Atoi(after[:end])
		if err != nil {
			continue
		}
		if y >= i && y < i+delegate.Height() {
			return idx, true
		}
	}
	return 0, false
}

type markerDelegate struct {
	height, spacing int
}

func (d markerDelegate) Height() int                             { return d.height }
func (d markerDelegate) Spacing() int                            { return d.spacing }
func (d markerDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d markerDelegate) Render(w io.Writer, _ list.Model, index int, _ list.Item) {
	_, _ = fmt.Fprintf(w, "%s%d;%s", listMarker, index, strings.Repeat("\n", max(d.height-1, 0)))
}
