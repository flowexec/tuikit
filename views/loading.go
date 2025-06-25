package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/themes"
)

const (
	LoadingViewType = "loading"
	DefaultLoading  = "loading..."
)

type LoadingView struct {
	theme   themes.Theme
	msg     string
	spinner spinner.Model
}

func (v *LoadingView) Init() tea.Cmd {
	return v.spinner.Tick
}

func (v *LoadingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case error:
		v.msg = msg.Error()
	case string:
		v.msg = msg
	}
	v.spinner, cmd = v.spinner.Update(msg)
	return v, cmd
}

func (v *LoadingView) View() string {
	var txt string
	if v.msg == "" {
		txt = fmt.Sprintf("\n\n %s %s\n\n", v.spinner.View(), v.theme.RenderInfo(DefaultLoading))
	} else {
		txt = fmt.Sprintf("\n\n %s %s\n\n", v.spinner.View(), v.theme.RenderInfo(v.msg))
	}
	return txt
}

func (v *LoadingView) HelpMsg() string {
	return ""
}

func (v *LoadingView) ShowFooter() bool {
	return false
}

func (v *LoadingView) Type() string {
	return LoadingViewType
}

func NewLoadingView(msg string, theme themes.Theme) *LoadingView {
	spin := spinner.New()
	spin.Style = theme.SpinnerStyle()
	spin.Spinner = theme.Spinner()
	return &LoadingView{
		theme:   theme,
		msg:     msg,
		spinner: spin,
	}
}
