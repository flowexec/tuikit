package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/themes"
	"github.com/jahvon/tuikit/types"
)

const ErrorViewType = "error"

type ErrorView struct {
	err   error
	theme themes.Theme
}

func NewErrorView(err error, theme themes.Theme) *ErrorView {
	return &ErrorView{
		err:   err,
		theme: theme,
	}
}

func (v *ErrorView) Init() tea.Cmd {
	return nil
}

func (v *ErrorView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//nolint:gocritic
	switch msg.(type) {
	case types.TickMsg:
		return v, tea.Quit
	}
	return v, nil
}

func (v *ErrorView) View() string {
	return v.theme.RenderError(errorString(v.err))
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
	// split on `:` or ` - ` to print wrapped errors on new lines
	// TODO: this is a hacky way to handle wrapped errors. Instead a defined error pattern should be enforced
	parts := strings.Split(err.Error(), ":")
	if len(parts) == 1 {
		parts = strings.Split(err.Error(), " - ")
	}
	for _, part := range parts {
		errStr += fmt.Sprintf("%s\n", part)
	}
	return errStr
}
