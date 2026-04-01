package views

import (
	"fmt"
	"sync"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit/themes"
)

const (
	LoadingViewType = "loading"
	DefaultLoading  = "loading..."
)

type LoadingView struct {
	theme   themes.Theme
	msg     string
	spinner spinner.Model
	mu      sync.RWMutex
}

func (v *LoadingView) Init() tea.Cmd {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.spinner.Tick
}

func (v *LoadingView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	v.mu.Lock()
	defer v.mu.Unlock()
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

func (v *LoadingView) View() tea.View {
	v.mu.RLock()
	defer v.mu.RUnlock()
	msg := v.msg
	if msg == "" {
		msg = DefaultLoading
	}
	txt := fmt.Sprintf("\n\n  %s %s\n\n", v.spinner.View(), v.theme.RenderInfo(msg))
	return tea.View{Content: txt}
}

func (v *LoadingView) HelpBindings() []themes.HelpKey {
	return nil
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
