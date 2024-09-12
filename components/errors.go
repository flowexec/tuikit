package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

const ErrorViewType = "error"

type ErrorView struct {
	err    error
	styles styles.Theme
}

func NewErrorView(err error, styles styles.Theme) *ErrorView {
	return &ErrorView{
		err:    err,
		styles: styles,
	}
}

func (v *ErrorView) Init() tea.Cmd {
	return nil
}

func (v *ErrorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//nolint:gocritic
	switch msg.(type) {
	case TickMsg:
		return v, tea.Quit
	}
	return v, nil
}

func (v *ErrorView) View() string {
	return v.styles.RenderError(errorString(v.err))
}

func (v *ErrorView) HelpMsg() string {
	return ""
}

func (v *ErrorView) ShowFooter() bool {
	return false
}

func (v *ErrorView) Type() string {
	return ErrorViewType
}

func errorString(err error) string {
	errStr := "!! encountered error !!\n\n"
	// split on `:` to print wrapped errors on new lines
	// TODO: this is a hacky way to handle wrapped errors. Instead a defined error pattern should be enforced
	parts := strings.Split(err.Error(), ":")
	for _, part := range parts {
		errStr += fmt.Sprintf("%s\n", part)
	}
	return errStr
}
