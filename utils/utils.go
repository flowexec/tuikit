package utils

import (
	tea "github.com/charmbracelet/bubbletea"
	teaV2 "github.com/charmbracelet/bubbletea/v2"
)

func CmdToV2Cmd(oldCmd tea.Cmd) (newCmd teaV2.Cmd) {
	newCmd = func() teaV2.Msg {
		if oldCmd != nil {
			return oldCmd()
		}
		return nil
	}
	return newCmd
}

func V2CmdToCmd(newCmd teaV2.Cmd) (oldCmd tea.Cmd) {
	oldCmd = func() tea.Msg {
		if newCmd != nil {
			return newCmd()
		}
		return nil
	}
	return oldCmd
}
