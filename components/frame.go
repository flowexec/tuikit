package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

const FrameViewType = "frame"

type FrameView struct {
	model tea.Model
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

func (v *FrameView) Interactive() bool {
	return false
}

func (v *FrameView) Type() string {
	return FrameViewType
}

func NewFrameView(model tea.Model) TeaModel {
	return &FrameView{
		model: model,
	}
}
