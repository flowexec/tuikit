package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

const (
	LoadingViewType = "loading"
	DefaultLoading  = "loading..."
)

type LoadingView struct {
	styles  styles.Theme
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
		txt = fmt.Sprintf("\n\n %s %s\n\n", v.spinner.View(), v.styles.RenderInfo(DefaultLoading))
	} else {
		txt = fmt.Sprintf("\n\n %s %s\n\n", v.spinner.View(), v.styles.RenderInfo(v.msg))
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

func NewLoadingView(msg string, styles styles.Theme) *LoadingView {
	spin := spinner.New()
	spin.Style = styles.Spinner()
	spin.Spinner = styles.SpinnerType
	return &LoadingView{
		styles:  styles,
		msg:     msg,
		spinner: spin,
	}
}
