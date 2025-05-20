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

func V2MsgToMsg(msg teaV2.Msg) (oldMsg tea.Msg) {
	switch m := msg.(type) {
	case teaV2.KeyMsg:
		switch {
		case m.Key().Text != "":
			oldMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(m.Key().Text)}
		case m.Key().Code != 0:
			switch m.Key().Code {
			case teaV2.KeySpace:
				oldMsg = tea.KeyMsg{Type: tea.KeySpace}
			case teaV2.KeyBackspace:
				oldMsg = tea.KeyMsg{Type: tea.KeyBackspace}
			case teaV2.KeyEnter:
				oldMsg = tea.KeyMsg{Type: tea.KeyEnter}
			case teaV2.KeyTab:
				oldMsg = tea.KeyMsg{Type: tea.KeyTab}
			case teaV2.KeyDown:
				oldMsg = tea.KeyMsg{Type: tea.KeyDown}
			case teaV2.KeyUp:
				oldMsg = tea.KeyMsg{Type: tea.KeyUp}
			case teaV2.KeyLeft:
				oldMsg = tea.KeyMsg{Type: tea.KeyLeft}
			case teaV2.KeyRight:
				oldMsg = tea.KeyMsg{Type: tea.KeyRight}
			case teaV2.KeyPgUp:
				oldMsg = tea.KeyMsg{Type: tea.KeyPgUp}
			case teaV2.KeyPgDown:
				oldMsg = tea.KeyMsg{Type: tea.KeyPgDown}
			case teaV2.KeyHome:
				oldMsg = tea.KeyMsg{Type: tea.KeyHome}
			case teaV2.KeyEnd:
				oldMsg = tea.KeyMsg{Type: tea.KeyEnd}
			case teaV2.KeyEsc:
				oldMsg = tea.KeyMsg{Type: tea.KeyEsc}
			case teaV2.KeyDelete:
				oldMsg = tea.KeyMsg{Type: tea.KeyDelete}
			case teaV2.KeyInsert:
				oldMsg = tea.KeyMsg{Type: tea.KeyInsert}
			}
		default:
			// unsupported for now
			panic("unsupported key type")
		}
	default:
		oldMsg = msg
	}
	return oldMsg
}
