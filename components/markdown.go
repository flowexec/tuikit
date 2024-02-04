package components

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/jahvon/tuikit/styles"
)

type MarkdownView struct {
	content       string
	viewport      viewport.Model
	err           TeaModel
	styles        styles.Theme
	width, height int
}

func NewMarkdownView(state *TerminalState, content string) TeaModel {
	vp := viewport.New(state.Width, state.Height)
	vp.Style = state.Theme.EntityView().Width(state.Width)
	return &MarkdownView{
		content:  content,
		styles:   state.Theme,
		width:    state.Width,
		height:   state.Height,
		viewport: vp,
	}
}

func (v *MarkdownView) Init() tea.Cmd {
	return nil
}

func (v *MarkdownView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *MarkdownView) View() string {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(v.styles.MarkdownStyleJSON)),
		glamour.WithWordWrap(v.width-2),
	)
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}

	viewStr, err := renderer.Render(v.content)
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	v.viewport.SetContent(viewStr)
	return v.viewport.View()
}

func (v *MarkdownView) HelpMsg() string {
	return ""
}

func (v *MarkdownView) Interactive() bool {
	return v.err == nil
}

func (v *MarkdownView) Type() string {
	return "markdown"
}
