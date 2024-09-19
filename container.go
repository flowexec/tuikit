package tuikit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jahvon/tuikit/styles"
	"github.com/jahvon/tuikit/types"
	"github.com/jahvon/tuikit/views"
)

type View interface {
	tea.Model

	ShowFooter() bool
	HelpMsg() string
	Type() string
}

type Container struct {
	ctx     context.Context
	cancel  context.CancelFunc
	app     *Application
	program *Program
	render  *types.RenderState

	previousView, currentView, nextView View
	showHelp                            bool
}

type ContainerOptions func(*Container)

var tickTime = time.Millisecond * 250

func NewContainer(
	ctx context.Context,
	app *Application,
	opts ...ContainerOptions,
) (*Container, error) {
	if app == nil {
		return nil, errors.New("application required")
	}

	ctxx, cancel := context.WithCancel(ctx)
	c := &Container{
		ctx:    ctxx,
		cancel: cancel,
		app:    app,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.program == nil {
		c.program = NewProgram(ctx, c, os.Stdin, os.Stdout)
	}
	if c.program.in == nil {
		c.program.in = os.Stdin
	}
	if c.program.out == nil {
		c.program.out = os.Stdout
	}
	if c.render == nil {
		c.render = &types.RenderState{}
	}
	if c.render.Theme == nil {
		def := styles.EverforestTheme()
		c.render.Theme = &def
	}

	return c, nil
}

func (c *Container) Start() error {
	go func() {
		_, err := c.program.Run()
		if err != nil {
			c.HandleError(err)
		}
		c.cancel()
	}()

	readyTimout := time.Now().Add(10 * time.Second)
	for {
		if c.Ready() {
			break
		} else if time.Now().After(readyTimout) {
			return errors.New("timed out waiting for container to be ready")
		}
		time.Sleep(tickTime)
	}
	return nil
}

func (c *Container) WaitForExit() {
	<-c.ctx.Done()
}

func (c *Container) HandleError(err error) {
	if err == nil {
		return
	}

	cErr := c.SetView(views.NewErrorView(err, *c.render.Theme))
	if cErr != nil {
		panic(err)
	}
}

func (c *Container) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	if c.currentView == nil {
		c.currentView = c.loadingView()
	}
	cmds = append(
		cmds,
		tea.SetWindowTitle(c.app.Name),
		c.doTick(),
		c.CurrentView().Init(),
	)
	return tea.Batch(cmds...)
}

//nolint:gocognit,funlen
func (c *Container) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	fwdMsg := msg
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return c, tea.Quit
	case tea.WindowSizeMsg:
		c.render = &types.RenderState{
			Width:         msg.Width,
			Height:        msg.Height,
			ContentWidth:  msg.Width,
			ContentHeight: msg.Height - (styles.HeaderHeight + styles.FooterHeight),
			Theme:         c.render.Theme,
		}
		if c.CurrentView().Type() == views.FormViewType {
			fwdMsg = tea.WindowSizeMsg{Width: c.render.ContentWidth, Height: c.render.ContentHeight}
		} else {
			fwdMsg = c.render
		}
	case types.ReplaceViewMsg:
		var err error
		switch {
		case c.NextView() != nil:
			err = c.SetView(c.NextView())
		case c.PreviousView() != nil:
			err = c.SetView(c.PreviousView())
		case c.CurrentView().Type() == views.FormViewType:
			return c, tea.Quit
		default:
			err = c.SetView(c.loadingView())
		}
		if err != nil {
			c.HandleError(err)
		}
	case tea.KeyMsg:
		if c.CurrentView().Type() == views.FormViewType {
			fwdMsg = nil
			_, cmd := c.CurrentView().Update(msg)
			cmds = append(cmds, cmd)
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			c.CurrentView().Update(tea.Quit())
			return c, tea.Quit
		case "esc", "backspace":
			if c.PreviousView() == nil || c.CurrentView() == c.PreviousView() {
				c.CurrentView().Update(tea.Quit())
				return c, tea.Quit
			} else {
				if err := c.SetView(c.PreviousView()); err != nil {
					c.HandleError(err)
				}
				return c, nil
			}
		case "h":
			if !c.CurrentView().ShowFooter() {
				break
			}
			fwdMsg = nil
			c.showHelp = !c.showHelp
		}
	case types.TickMsg:
		if c.Ready() && c.CurrentView().Type() == views.LoadingViewType && c.nextView != nil {
			c.currentView = c.nextView
			c.nextView = nil
		}
		cmds = append(cmds, types.Tick)
	case tea.Cmd:
		cmds = append(cmds, msg)
	}
	if fwdMsg != nil {
		_, cmd := c.CurrentView().Update(fwdMsg)
		cmds = append(cmds, cmd)
	}
	return c, tea.Batch(cmds...)
}

func (c *Container) View() string {
	var footer string
	var footerPrefix string

	if !c.Ready() && c.CurrentView().Type() != views.LoadingViewType {
		return ""
	}
	switch {
	case c.CurrentView().Type() == views.FrameViewType:
		return c.CurrentView().View()
	case c.CurrentView().Type() == views.LoadingViewType, c.CurrentView().Type() == views.FormViewType:
		footer = ""
	case !c.CurrentView().ShowFooter():
		footer = c.render.Theme.RenderFooter(c.app.notice, c.render.Width)
	case c.CurrentView().ShowFooter() && c.showHelp:
		footerPrefix = "[ q: quit ] [ h: hide help ] [ ↑/↓: navigate ]"
		if c.PreviousView() != nil {
			footerPrefix += " [ esc: back ]"
		}
		footer = c.render.Theme.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, c.CurrentView().HelpMsg()), c.render.Width)
	case c.CurrentView().ShowFooter() && !c.showHelp && c.CurrentView().HelpMsg() != "":
		footerPrefix = "[ q: quit ] [ h: show help ]"
		if c.app.notice != "" {
			footer = c.render.Theme.RenderFooter(
				fmt.Sprintf("%s ● %s ● %s", footerPrefix, c.CurrentView().HelpMsg(), c.app.notice), c.render.Width,
			)
		} else {
			footer = c.render.Theme.RenderFooter(footerPrefix, c.render.Width)
		}
	case c.CurrentView().ShowFooter() && !c.showHelp:
		footerPrefix = "[ q: quit ] [ ↑/↓: navigate ]"
		if c.app.notice != "" {
			footer = c.render.Theme.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, c.app.notice), c.render.Width)
		} else {
			footer = c.render.Theme.RenderFooter(footerPrefix, c.render.Width)
		}
	}

	header := c.render.Theme.RenderHeader(c.app.Name, c.app.stateKey, c.app.stateVal, c.render.Width)
	return lipgloss.JoinVertical(lipgloss.Top, header, c.CurrentView().View(), footer)
}

func (c *Container) Ready() bool {
	switch {
	case !c.program.Started():
		return false
	case !c.SizeSet():
		return false
	default:
		return true
	}
}

func (c *Container) Shutdown() {
	// exit the program
	_ = c.program.Suspend()
	c.Update(tea.Quit())

	// clear the screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic("unable to clear screen")
	}
}

func (c *Container) Height() int {
	return c.render.Height
}

func (c *Container) ContentHeight() int {
	return c.render.ContentHeight
}

func (c *Container) Width() int {
	return c.render.Width
}

func (c *Container) ContentWidth() int {
	return c.render.ContentWidth
}

func (c *Container) SetView(v View) error {
	switch {
	case v == nil:
		return errors.New("view not provided")
	case !c.program.Started():
		c.SetNextView(v)
		return nil
	case c.program.Suspended():
		if err := c.program.Resume(); err != nil {
			return fmt.Errorf("unable to resume program - %w", err)
		}
	}

	switching := c.CurrentView() != nil && c.CurrentView().Type() != v.Type() &&
		c.CurrentView().Type() != views.LoadingViewType && c.CurrentView().Type() != views.FormViewType
	switch {
	case !c.Ready():
		c.SetNextView(v)
	case switching:
		c.previousView = c.CurrentView()
		fallthrough
	default:
		c.currentView = v
		if c.currentView == c.nextView {
			c.nextView = nil
		}
		cmd := c.CurrentView().Init()
		if cmd != nil {
			c.Send(cmd, 0)
		}
	}

	return nil
}

func (c *Container) Send(msg tea.Msg, delay time.Duration) {
	if delay > 0 {
		go func() {
			time.Sleep(delay)
			c.program.Send(msg)
		}()
	} else {
		c.program.Send(msg)
	}
}

func (c *Container) SetNextView(v View) {
	c.nextView = v
}

func (c *Container) CurrentView() View {
	return c.currentView
}

func (c *Container) PreviousView() View {
	return c.previousView
}

func (c *Container) NextView() View {
	return c.nextView
}

func (c *Container) SizeSet() bool {
	return c.render.Width > 0 && c.render.Height > 0 && c.render.ContentWidth > 0 && c.render.ContentHeight > 0
}

func (c *Container) RenderState() *types.RenderState {
	return c.render
}

func (c *Container) SetNotice(notice string, lvl styles.NoticeLevel) {
	c.app.notice = c.render.Theme.RenderNotice(notice, lvl)
}

func (c *Container) SetState(key, val string) {
	c.app.stateKey = key
	c.app.stateVal = val
}

func (c *Container) SetStateValue(val string) {
	c.app.stateVal = val
}

func (c *Container) State() (string, string) {
	return c.app.stateKey, c.app.stateVal
}

func (c *Container) doTick() tea.Cmd {
	return tea.Tick(tickTime, func(t time.Time) tea.Msg {
		return types.TickMsg(t)
	})
}
func (c *Container) loadingView() View {
	return views.NewLoadingView(c.app.loadingMsg, *c.render.Theme)
}

func WithInitialTermSize(width, height int) ContainerOptions {
	return func(c *Container) {
		if c.render == nil {
			c.render = &types.RenderState{}
		}
		c.render.Width = width
		c.render.Height = height
		c.render.ContentWidth = width
		c.render.ContentHeight = height - (styles.HeaderHeight + styles.FooterHeight)
	}
}

func WithInput(in io.Reader) ContainerOptions {
	return func(c *Container) {
		if c.program == nil {
			c.program = NewProgram(c.ctx, c, in, os.Stdout)
		}
		c.program.in = in
	}
}

func WithOutput(out io.Writer) ContainerOptions {
	return func(c *Container) {
		if c.program == nil {
			c.program = NewProgram(c.ctx, c, os.Stdin, out)
		}
		c.program.out = out
	}
}

func WithTheme(styles styles.Theme) ContainerOptions {
	return func(c *Container) {
		c.render.Theme = &styles
	}
}
