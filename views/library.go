package views

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

const LibraryViewType = "library"

// breadcrumbHeight is the number of lines reserved for the breadcrumb bar
// (1 line of text + 1 line of bottom margin).
const breadcrumbHeight = 2

// Selectable is an optional interface sub-views can implement to report
// the current selection. The Library uses this when navigating forward
// to record what the user selected on the current page.
type Selectable interface {
	SelectedIndex() int
	SelectedData() []string
}

// PageSelection records what the user selected on a given page.
type PageSelection struct {
	Index int
	Data  []string
}

// PageFactory builds the sub-view for a given page. It receives the
// render state (adjusted for the breadcrumb) and the selections made
// on all prior pages. It returns the sub-view model and an optional
// slice of KeyCallbacks for domain-specific actions on this page.
type PageFactory func(render *types.RenderState, selections []PageSelection) (tea.Model, []types.KeyCallback)

// LibraryPage defines a single page in the Library's drill-down.
type LibraryPage struct {
	Title   string
	Factory PageFactory
}

// Library is a composite view that manages page-based drill-down
// navigation. Each page is built lazily by a factory function that
// receives the selections from all prior pages.
type Library struct {
	render *types.RenderState

	pages      []LibraryPage
	pageIndex  int
	selections []PageSelection

	activeView  tea.Model
	activeKeys  []types.KeyCallback
	breadcrumbs []string
}

// NewLibrary creates a new Library with the given pages. At least one
// page is required. Page 0 is activated immediately.
func NewLibrary(render *types.RenderState, pages ...LibraryPage) *Library {
	if len(pages) == 0 {
		panic("Library requires at least one page")
	}
	lib := &Library{
		render:     render,
		pages:      pages,
		selections: make([]PageSelection, 0, len(pages)),
	}
	lib.activatePage(0)
	return lib
}

func (l *Library) Init() tea.Cmd {
	if l.activeView == nil {
		return nil
	}
	return l.activeView.Init()
}

func (l *Library) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *types.RenderState:
		l.render = msg
		sub := l.subViewRenderState()
		_, cmd := l.activeView.Update(sub)
		return l, cmd
	case tea.KeyPressMsg:
		return l, l.handleKeyMsg(msg)
	}

	_, cmd := l.activeView.Update(msg)
	return l, cmd
}

func (l *Library) handleKeyMsg(msg tea.KeyPressMsg) tea.Cmd {
	// If the sub-view is capturing input (e.g. filter mode),
	// forward everything to it unconditionally.
	if ic, ok := l.activeView.(interface{ CapturingInput() bool }); ok && ic.CapturingInput() {
		_, cmd := l.activeView.Update(msg)
		return cmd
	}

	switch msg.String() {
	case types.KeyEnter, "right":
		return l.handleForwardKey(msg)
	case "esc", "backspace", "left":
		if l.pageIndex > 0 {
			l.navigateBack()
			return l.activeView.Init()
		}
		// Page 0: don't handle, let Container deal with it
		return nil
	default:
		return l.handleCallbackKey(msg)
	}
}

func (l *Library) handleForwardKey(msg tea.KeyPressMsg) tea.Cmd {
	if l.pageIndex < len(l.pages)-1 {
		return l.navigateForward()
	}
	// Last page: forward enter to sub-view
	if msg.String() == types.KeyEnter {
		_, cmd := l.activeView.Update(msg)
		return cmd
	}
	return nil
}

func (l *Library) handleCallbackKey(msg tea.KeyPressMsg) tea.Cmd {
	for _, cb := range l.activeKeys {
		if cb.Key == msg.String() {
			cb := cb
			return func() tea.Msg {
				err := cb.Callback()
				if err != nil {
					return err
				}
				return nil
			}
		}
	}
	_, cmd := l.activeView.Update(msg)
	return cmd
}

func (l *Library) View() tea.View {
	breadcrumb := l.renderBreadcrumb()
	activeContent := l.activeView.View()

	content := lipgloss.JoinVertical(lipgloss.Left, breadcrumb, activeContent.Content)
	return tea.View{Content: content}
}

func (l *Library) HelpBindings() []themes.HelpKey {
	keys := make([]themes.HelpKey, 0)

	// Domain-specific keys from the active page
	for _, cb := range l.activeKeys {
		if cb.Key != "" && cb.Label != "" {
			keys = append(keys, themes.HelpKey{Key: cb.Key, Desc: cb.Label})
		}
	}

	// Build set of keys the Library intercepts so we can skip
	// conflicting sub-view bindings.
	intercepted := make(map[string]bool)
	if l.pageIndex < len(l.pages)-1 {
		intercepted["enter"] = true
	}
	if l.pageIndex > 0 {
		intercepted["esc"] = true
	}

	// Sub-view keys (e.g. Table's filter, scroll), minus intercepted ones
	if hb, ok := l.activeView.(interface{ HelpBindings() []themes.HelpKey }); ok {
		for _, k := range hb.HelpBindings() {
			if !intercepted[k.Key] {
				keys = append(keys, k)
			}
		}
	}

	// Library navigation keys
	if l.pageIndex < len(l.pages)-1 {
		keys = append(keys, themes.HelpKey{Key: "enter/→", Desc: "drill down"})
	}
	if l.pageIndex > 0 {
		keys = append(keys, themes.HelpKey{Key: "esc/←", Desc: "go back"})
	}

	return keys
}

func (l *Library) Type() string {
	return LibraryViewType
}

// CapturingInput implements InputCapturer. When on a deeper page
// (pageIndex > 0), the Library captures esc/backspace for internal
// back-navigation. It also delegates to the sub-view if it is
// capturing input (e.g. Table filter mode).
func (l *Library) CapturingInput() bool {
	if l.pageIndex > 0 {
		return true
	}
	if ic, ok := l.activeView.(interface{ CapturingInput() bool }); ok {
		return ic.CapturingInput()
	}
	return false
}

func (l *Library) activatePage(index int) {
	l.pageIndex = index
	if len(l.selections) > index {
		l.selections = l.selections[:index]
	}

	sub := l.subViewRenderState()
	view, keys := l.pages[index].Factory(sub, l.selections)
	l.activeView = view
	l.activeKeys = keys
	l.buildBreadcrumbs()
}

func (l *Library) navigateForward() tea.Cmd {
	sel, ok := l.extractSelection()
	if !ok {
		return nil
	}

	l.selections = append(l.selections[:l.pageIndex], sel)
	l.activatePage(l.pageIndex + 1)
	return l.activeView.Init()
}

func (l *Library) navigateBack() {
	l.activatePage(l.pageIndex - 1)
}

func (l *Library) extractSelection() (PageSelection, bool) {
	if s, ok := l.activeView.(Selectable); ok {
		data := s.SelectedData()
		if data == nil {
			return PageSelection{}, false
		}
		return PageSelection{
			Index: s.SelectedIndex(),
			Data:  data,
		}, true
	}
	return PageSelection{}, false
}

func (l *Library) subViewRenderState() *types.RenderState {
	return &types.RenderState{
		Width:         l.render.Width,
		Height:        l.render.Height,
		ContentWidth:  l.render.ContentWidth,
		ContentHeight: l.render.ContentHeight - breadcrumbHeight,
		Theme:         l.render.Theme,
	}
}

func (l *Library) buildBreadcrumbs() {
	l.breadcrumbs = make([]string, 0, l.pageIndex+1)
	for i := 0; i <= l.pageIndex; i++ {
		label := l.pages[i].Title
		if i < len(l.selections) && len(l.selections[i].Data) > 0 {
			label += ": " + l.selections[i].Data[0]
		}
		l.breadcrumbs = append(l.breadcrumbs, label)
	}
}

func (l *Library) renderBreadcrumb() string {
	if l.render == nil {
		return ""
	}
	cp := l.render.Theme.ColorPalette()

	parts := make([]string, len(l.breadcrumbs))
	for i, b := range l.breadcrumbs {
		if i == len(l.breadcrumbs)-1 {
			parts[i] = lipgloss.NewStyle().
				Foreground(cp.PrimaryColor()).Bold(true).
				Render(b)
		} else {
			parts[i] = lipgloss.NewStyle().
				Foreground(cp.GrayColor()).
				Render(b)
		}
	}

	sep := lipgloss.NewStyle().Foreground(cp.GrayColor()).Render(" > ")
	trail := strings.Join(parts, sep)
	return lipgloss.NewStyle().MarginLeft(2).MarginBottom(1).Render(trail)
}
