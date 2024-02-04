package components

import (
	"context"
	"fmt"
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jahvon/tuikit/styles"
)

const (
	heightPadding = 5
)

type ContainerView struct {
	appName string
	ctx     context.Context

	program     *tea.Program
	header      Header
	pendingView TeaModel
	activeView  TeaModel
	lastView    TeaModel
	styles      styles.Theme

	width, height int
	ready         bool
}

func InitalizeContainer(
	ctx context.Context,
	cancel context.CancelFunc,
	header Header,
	styles styles.Theme,
) *ContainerView {
	activeView := NewLoadingView("", styles)
	a := &ContainerView{
		appName:    header.Name,
		ctx:        ctx,
		header:     header,
		styles:     styles,
		activeView: activeView,
	}
	prgm := tea.NewProgram(a, tea.WithContext(ctx))
	go func() {
		var err error
		if _, err = prgm.Run(); err != nil {
			panic(fmt.Errorf("error running application: %w", err))
		}
		cancel()
	}()
	a.program = prgm
	return a
}

func (a *ContainerView) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	if a.activeView != nil {
		cmds = append(cmds, a.activeView.Init())
	}
	cmds = append(
		cmds,
		tea.SetWindowTitle(a.appName),
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
	return tea.Batch(cmds...)
}

func (a *ContainerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			a.activeView.Update(tea.Quit())
			return a, tea.Quit
		case "esc", "backspace":
			if a.lastView == nil || a.activeView == a.lastView {
				a.activeView.Update(tea.Quit())
				return a, tea.Quit
			} else {
				a.activeView = a.lastView
				a.lastView = nil
				return a, nil
			}
		}
	case tea.WindowSizeMsg:
		msg.Width = int(math.Floor(float64(msg.Width) * 0.90))
		a.width = msg.Width
		msg.Height = msg.Height - heightPadding
		a.height = msg.Height

		if !a.Ready() {
			if a.pendingView != nil {
				a.activeView = a.pendingView
				a.pendingView = nil
			}
			a.ready = true
		}
	case TickMsg:
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}))
	case tea.Cmd:
		cmds = append(cmds, msg)
	}

	_, cmd := a.activeView.Update(msg)
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)
}

func (a *ContainerView) Ready() bool {
	return a.ready && a.header.CtxVal != "unk"
}

func (a *ContainerView) Finalize() {
	a.header.Notice = ""
}

func (a *ContainerView) Height() int {
	return a.height
}

func (a *ContainerView) Width() int {
	return a.width
}

func (a *ContainerView) View() string {
	var help string
	var lastViewHelp string

	if !a.Ready() && a.activeView.Type() != LoadingViewType {
		a.activeView = NewLoadingView("", a.styles)
	}
	if a.activeView.Interactive() && a.lastView != nil {
		lastViewHelp = "esc: back • "
	}
	if a.activeView.Interactive() && a.activeView.HelpMsg() != "" {
		help = a.styles.RenderHelp(fmt.Sprintf("\n %s | %sq: quit • ↑/↓: navigate", a.activeView.HelpMsg(), lastViewHelp))
	} else if a.activeView.Interactive() {
		help = a.styles.RenderHelp(fmt.Sprintf("\n %sq: quit • ↑/↓: navigate", lastViewHelp))
	}

	header := a.header.View()
	return header + "\n" + a.activeView.View() + help
}

func (a *ContainerView) SetContext(ctx string) {
	if ctx != "" {
		a.header.CtxVal = ctx
	}
}

func (a *ContainerView) SetNotice(notice string, lvl NoticeLevel) {
	a.header.Notice = notice
	a.header.NoticeLevel = lvl
}

func (a *ContainerView) SetView(model TeaModel) {
	if !a.Ready() {
		a.pendingView = model
		return
	}
	if a.activeView != nil && a.activeView.Type() != LoadingViewType && a.activeView.Type() != model.Type() {
		a.lastView = a.activeView
	}
	a.activeView = model
	cmd := a.activeView.Init()
	a.Update(cmd)
}

func (a *ContainerView) HandleError(err error) {
	if err == nil {
		return
	}

	a.SetView(NewErrorView(err, a.styles))
}
