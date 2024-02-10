package components

import (
	"math"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"golang.org/x/term"

	"github.com/jahvon/tuikit/styles"
)

type MarkdownView struct {
	content       string
	viewport      viewport.Model
	err           TeaModel
	styles        styles.Theme
	width, height int
}

func RunMarkdownView(theme styles.Theme, content string) error {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	w = int(math.Floor(float64(w) * 0.90))
	vp := viewport.New(w, h)
	vp.Style = theme.EntityView().Width(w).Height(h)
	view := &MarkdownView{
		content:  content,
		styles:   theme,
		width:    w,
		height:   h,
		viewport: vp,
	}
	p := tea.NewProgram(view, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (v *MarkdownView) Init() tea.Cmd {
	return v.viewport.Init()
}

func (v *MarkdownView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return v, tea.Quit
		case "up":
			v.viewport.LineUp(1)
		case "down":
			v.viewport.LineDown(1)
		}
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *MarkdownView) View() string {
	if v.err != nil {
		return v.err.View()
	}
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
	help := v.styles.RenderHelp("\nq: quit • ↑/↓: navigate")
	return v.viewport.View() + help
}

func (v *MarkdownView) RunProgram() error {
	p := tea.NewProgram(v)
	_, err := p.Run()
	return err
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
