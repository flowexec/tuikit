package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

type TextInput struct {
	input *textinput.Model
	err   TeaModel

	Key         string
	Prompt      string
	Placeholder string
	Hidden      bool
}

func (t *TextInput) Value() string {
	return t.input.Value()
}

func (t *TextInput) RunProgram(styles styles.Theme) (string, error) {
	in := textinput.New()
	in.PromptStyle = in.PromptStyle.Foreground(styles.InfoColor)
	echoMode := textinput.EchoNormal
	if t.Hidden {
		echoMode = textinput.EchoPassword
	}
	in.EchoMode = echoMode
	if t.Placeholder != "" {
		in.Placeholder = t.Placeholder
	}
	in.Prompt = t.Prompt
	t.input = &in

	p := tea.NewProgram(t)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
	return t.Value(), nil
}

func (t *TextInput) Init() tea.Cmd {
	return textinput.Blink
}

func (t *TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if t.err != nil {
		return t.err.Update(msg)
	}
	var cmd tea.Cmd

	//nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		//nolint:exhaustive
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return t, tea.Quit
		}
	}

	*t.input, cmd = t.input.Update(msg)
	return t, cmd
}

func (t *TextInput) View() string {
	if t.err != nil {
		return t.err.View()
	}

	t.input.Focus()
	return t.input.View()
}

func (t *TextInput) HelpMsg() string {
	return ""
}

func (t *TextInput) Interactive() bool {
	return true
}

func (t *TextInput) Type() string {
	return fmt.Sprintf("textinput-%s", t.Key)
}

type TextInputList []TextInput

func (t TextInputList) FindByKey(key string) *TextInput {
	for _, in := range t {
		if in.Key == key {
			return &in
		}
	}
	return nil
}

func (t TextInputList) ValueMap() map[string]string {
	m := make(map[string]string)
	for _, in := range t {
		m[in.Key] = in.Value()
	}
	return m
}

func ProcessInputs(styles styles.Theme, inputs ...TextInput) (TextInputList, error) {
	for _, in := range inputs {
		_, err := in.RunProgram(styles)
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}
