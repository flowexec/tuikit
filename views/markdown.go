package views

import (
	"math"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jahvon/glamour"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

type MarkdownView struct {
	content       string
	viewport      viewport.Model
	err           *ErrorView
	theme         themes.Theme
	width, height int
}

func NewMarkdownView(state *types.RenderState, content string) *MarkdownView {
	vp := viewport.New(state.ContentWidth, state.ContentHeight)
	vp.Style = state.Theme.EntityViewStyle().Width(state.ContentWidth).Height(state.ContentHeight)
	return &MarkdownView{
		content:  content,
		viewport: vp,
		theme:    state.Theme,
	}
}

func (v *MarkdownView) Init() tea.Cmd {
	return v.viewport.Init()
}

func (v *MarkdownView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	switch msg := msg.(type) {
	case types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.viewport.Width = v.width
		v.viewport.Height = v.height
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyUp.String():
			v.viewport.ScrollUp(1)
		case tea.KeyDown.String():
			v.viewport.ScrollDown(1)
		}
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *MarkdownView) View() string {
	mdStyles, err := v.theme.GlamourMarkdownStyleJSON()
	if err != nil {
		v.err = NewErrorView(err, v.theme)
		return v.err.View()
	}
	if v.err != nil {
		return v.err.View()
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(mdStyles)),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(int(math.Floor(float64(v.width)*0.95))),
	)
	if err != nil {
		v.err = NewErrorView(err, v.theme)
		return v.err.View()
	}

	viewStr, err := renderer.Render(v.content)
	if err != nil {
		v.err = NewErrorView(err, v.theme)
		return v.err.View()
	}
	v.viewport.SetContent(viewStr)
	return v.viewport.View()
}

func (v *MarkdownView) HelpMsg() string {
	return ""
}

func (v *MarkdownView) ShowFooter() bool {
	return true
}

func (v *MarkdownView) Type() string {
	return "markdown"
}
