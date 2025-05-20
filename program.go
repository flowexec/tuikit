package tuikit

import (
	"context"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea/v2"
)

type Program struct {
	started   bool
	suspended bool

	program *tea.Program
	in      io.Reader
	out     io.Writer
}

func NewProgram(ctx context.Context, model tea.Model, in io.Reader, out io.Writer) *Program {
	p := tea.NewProgram(
		model,
		tea.WithContext(ctx),
		tea.WithInput(in),
		tea.WithOutput(out),
	)
	return &Program{
		program: p,
		in:      in,
		out:     out,
	}
}

func (p *Program) Run() (tea.Model, error) {
	if p.started || p.suspended {
		return nil, fmt.Errorf("program already started")
	}
	p.started = true
	return p.program.Run()
}

func (p *Program) Started() bool {
	return p.started
}

func (p *Program) Suspend() error {
	err := p.program.ReleaseTerminal()
	if err != nil {
		return err
	}
	p.suspended = true
	return nil
}

func (p *Program) Suspended() bool {
	return p.suspended
}

func (p *Program) Resume() error {
	err := p.program.RestoreTerminal()
	if err != nil {
		return err
	}
	p.suspended = false
	return nil
}

func (p *Program) Send(msg tea.Msg) {
	p.program.Send(msg)
}
