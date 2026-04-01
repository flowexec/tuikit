package views

import (
	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit/themes"
)

const FrameViewType = "frame"

type FramedModel interface {
	tea.Model
}

type FrameView struct {
	model FramedModel
}

func (v *FrameView) Init() tea.Cmd {
	return v.model.Init()
}

func (v *FrameView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return v.model.Update(msg)
}

func (v *FrameView) View() tea.View {
	return v.model.View()
}

func (v *FrameView) HelpBindings() []themes.HelpKey {
	return nil
}

func (v *FrameView) Type() string {
	return FrameViewType
}

func NewFrameView(model FramedModel) *FrameView {
	return &FrameView{
		model: model,
	}
}
