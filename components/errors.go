package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

const ErrorViewType = "error"

type ErrorView struct {
	err    error
	styles styles.Theme
}

func NewErrorView(err error, styles styles.Theme) TeaModel {
	return &ErrorView{
		err:    err,
		styles: styles,
	}
}

func (v *ErrorView) Init() tea.Cmd {
	return nil
}

func (v *ErrorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TickMsg:
		return v, tea.Quit
	}
	return v, nil
}

func (v *ErrorView) View() string {
	return v.styles.RenderError(
		fmt.Errorf("!! encountered error !!\n%w", v.err).Error(),
	)
}

func (v *ErrorView) HelpMsg() string {
	return ""
}

func (v *ErrorView) Interactive() bool {
	return false
}

func (v *ErrorView) Type() string {
	return ErrorViewType
}
