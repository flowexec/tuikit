package views

import (
	tea "github.com/charmbracelet/bubbletea"
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

func (v *FrameView) View() string {
	return v.model.View()
}

func (v *FrameView) HelpMsg() string {
	return ""
}

func (v *FrameView) ShowFooter() bool {
	return false
}

func (v *FrameView) Type() string {
	return FrameViewType
}

func NewFrameView(model FramedModel) *FrameView {
	return &FrameView{
		model: model,
	}
}
