package overlay

import (
	"github.com/flowexec/tuikit/themes"
)

// HelpPopup manages a centered help overlay that displays keybindings.
// It is not a tea.Model — the Container owns it and handles key interception.
type HelpPopup struct {
	visible    bool
	globalKeys []themes.HelpKey
	viewKeys   []themes.HelpKey
	theme      themes.Theme
}

// NewHelpPopup creates a new HelpPopup with the given theme and global keybindings.
func NewHelpPopup(theme themes.Theme) *HelpPopup {
	return &HelpPopup{
		theme: theme,
		globalKeys: []themes.HelpKey{
			{Key: "q", Desc: "quit"},
			{Key: "esc/bksp", Desc: "back"},
			{Key: "?/h", Desc: "toggle help"},
		},
	}
}

func (h *HelpPopup) Toggle() {
	h.visible = !h.visible
}

func (h *HelpPopup) Visible() bool {
	return h.visible
}

func (h *HelpPopup) SetViewKeys(keys []themes.HelpKey) {
	h.viewKeys = keys
}

// Render produces the styled help popup string for overlay composition.
func (h *HelpPopup) Render(width, height int) string {
	keys := make([]themes.HelpKey, 0, len(h.viewKeys)+len(h.globalKeys))
	keys = append(keys, h.viewKeys...)
	keys = append(keys, h.globalKeys...)
	return h.theme.RenderHelpPopup(keys, width, height)
}
