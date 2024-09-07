package components

import (
	"math"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jahvon/glamour"
	"golang.org/x/term"

	"github.com/jahvon/tuikit/styles"
)

type MarkdownView struct {
	appName, ctxKey, ctxValue string
	content                   string
	viewport                  viewport.Model
	err                       *ErrorView
	styles                    styles.Theme
	width, height             int
}

func RunMarkdownView(theme styles.Theme, appName, ctxKey, ctxValue, content string) error {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return err
	}
	view := &MarkdownView{
		appName:  appName,
		ctxKey:   ctxKey,
		ctxValue: ctxValue,
		content:  content,
		styles:   theme,
		width:    w,
		height:   h - styles.FooterHeight,
	}
	vp := viewport.New(view.width, view.height)
	vp.Style = theme.EntityView().Width(w).Height(h)
	view.viewport = vp
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height - styles.FooterHeight - styles.HeaderHeight
		v.viewport.Width = v.width
		v.viewport.Height = v.height
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
	mdStyles, err := v.styles.MarkdownStyleJSON()
	if err != nil {
		v.err = NewErrorView(err, v.styles)
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
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}

	viewStr, err := renderer.Render(v.content)
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	v.viewport.SetContent(viewStr)
	header := v.styles.RenderHeader(v.appName, v.ctxKey, v.ctxValue, v.width)
	footer := v.styles.RenderFooter("[ q: quit ] [ ↑/↓: navigate ]", v.width)
	return lipgloss.JoinVertical(lipgloss.Top, header, v.viewport.View(), footer)
}

func (v *MarkdownView) RunProgram() error {
	p := tea.NewProgram(v)
	_, err := p.Run()
	return err
}

func (v *MarkdownView) HelpMsg() string {
	return ""
}

func (v *MarkdownView) ShowFooter() bool {
	return v.err == nil
}

func (v *MarkdownView) Type() string {
	return "markdown"
}
