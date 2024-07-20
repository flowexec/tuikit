package components

import (
	"fmt"
	"io"

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
	if t.input == nil {
		return ""
	}
	return t.input.Value()
}

func (t *TextInput) RunProgram(styles styles.Theme, inReader io.Reader) (string, error) {
	in := textinput.New()
	echoMode := textinput.EchoNormal
	if t.Hidden {
		echoMode = textinput.EchoPassword
	}
	in.EchoMode = echoMode
	if t.Placeholder != "" {
		in.Placeholder = t.Placeholder
	}
	in.Prompt = t.Prompt
	in.PromptStyle = in.PromptStyle.Foreground(styles.InfoColor).PaddingRight(2)
	in.Focus()
	t.input = &in

	var p *tea.Program
	if inReader == nil {
		p = tea.NewProgram(t)
	} else {
		p = tea.NewProgram(t, tea.WithInput(inReader))
	}
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

type TextInputList []*TextInput

func (t TextInputList) FindByKey(key string) *TextInput {
	for _, in := range t {
		if in.Key == key {
			return in
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

func ProcessInputs(styles styles.Theme, inReader io.Reader, inputs ...*TextInput) (TextInputList, error) {
	for _, in := range inputs {
		_, err := in.RunProgram(styles, inReader)
		if err != nil {
			return nil, err
		}
	}
	return inputs, nil
}
