package components

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jahvon/tuikit/styles"
)

type ContainerView struct {
	appName, headerCtxKey, headerCtxVal string
	footerNotice                        string
	ctx                                 context.Context

	program     *tea.Program
	pendingView TeaModel
	activeView  TeaModel
	lastView    TeaModel
	styles      styles.Theme

	width, height, fullHeight int
	ready                     bool
	showHelp                  bool
}

func InitalizeContainer(
	ctx context.Context,
	cancel context.CancelFunc,
	appName, headerCtxKey, headerCtxVal string,
	styles styles.Theme,
) *ContainerView {
	activeView := NewLoadingView("", styles)
	a := &ContainerView{
		appName:      appName,
		headerCtxKey: headerCtxKey,
		headerCtxVal: headerCtxVal,
		ctx:          ctx,
		styles:       styles,
		activeView:   activeView,
	}
	prgm := tea.NewProgram(a, tea.WithContext(ctx))
	go func() {
		_, _ = prgm.Run()
		cancel()
	}()
	readyTimout := time.Now().Add(10 * time.Second)
	for {
		if a.Ready() {
			break
		} else if time.Now().After(readyTimout) {
			panic("timed out waiting for container to be ready")
		}
		time.Sleep(100 * time.Millisecond)
	}
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
		tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
	return tea.Batch(cmds...)
}

func (a *ContainerView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return a, tea.Quit
	case tea.WindowSizeMsg:
		a.ready = true
		a.width = msg.Width
		a.height = msg.Height - (styles.HeaderHeight + styles.FooterHeight)
		a.fullHeight = msg.Height
		if a.pendingView != nil {
			a.activeView = a.pendingView
			a.pendingView = nil
		}
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
		case "h":
			if a.activeView.HelpMsg() == "" {
				break
			}
			a.showHelp = !a.showHelp
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
	return a.ready
}

func (a *ContainerView) Height() int {
	return a.height
}

func (a *ContainerView) Width() int {
	return a.width
}

func (a *ContainerView) View() string {
	var footer string
	var footerPrefix string

	if !a.Ready() && a.activeView.Type() != LoadingViewType {
		a.activeView = NewLoadingView("", a.styles)
	}
	switch {
	case a.activeView.Type() == FrameViewType:
		return a.activeView.View()
	case a.activeView.Type() == LoadingViewType:
		footer = ""
	case !a.activeView.Interactive():
		footer = a.styles.RenderFooter(a.footerNotice, a.width)
	case a.showHelp:
		footerPrefix = "[ q: quit] [ h: hide help ] [ ↑/↓: navigate ]"
		if a.lastView != nil {
			footerPrefix += " [ esc: back ]"
		}
		footer = a.styles.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, a.activeView.HelpMsg()), a.width)
	case !a.showHelp && a.activeView.HelpMsg() != "":
		footerPrefix = "[ q: quit] [ h: show help ]"
		if a.footerNotice != "" {
			footer = a.styles.RenderFooter(
				fmt.Sprintf("%s ● %s ● %s", footerPrefix, a.activeView.HelpMsg(), a.footerNotice), a.width,
			)
		} else {
			footer = a.styles.RenderFooter(footerPrefix, a.width)
		}
	case !a.showHelp && a.activeView.HelpMsg() == "":
		footerPrefix = "[ q: quit]"
		if a.footerNotice != "" {
			footer = a.styles.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, a.footerNotice), a.width)
		} else {
			footer = a.styles.RenderFooter(footerPrefix, a.width)
		}
	}

	header := a.styles.RenderHeader(a.appName, a.headerCtxKey, a.headerCtxVal, a.width)
	return lipgloss.JoinVertical(lipgloss.Top, header, a.activeView.View(), footer)
}

func (a *ContainerView) SetContext(ctx string) {
	if ctx != "" && a.headerCtxKey != "" {
		a.headerCtxVal = ctx
	}
}

func (a *ContainerView) SetNotice(notice string, lvl styles.NoticeLevel) {
	a.footerNotice = a.styles.RenderNotice(notice, lvl)
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
	if cmd != nil {
		a.program.Send(cmd)
	}

	if a.width != 0 && a.height != 0 {
		a.program.Send(tea.WindowSizeMsg{Width: a.width, Height: a.fullHeight})
	}
}

func (a *ContainerView) HandleError(err error) {
	if err == nil {
		return
	}

	a.SetView(NewErrorView(err, a.styles))
}

func (a *ContainerView) Shutdown() {
	// exit the program
	_ = a.program.ReleaseTerminal()
	a.Update(tea.Quit())

	// clear the screen
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	if err := c.Run(); err != nil {
		panic("unable to clear screen")
	}
}
