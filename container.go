package tuikit

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jahvon/tuikit/components"
	"github.com/jahvon/tuikit/styles"
)

type View interface {
	tea.Model

	ShowFooter() bool
	HelpMsg() string
	Type() string
}

type Container struct {
	appName      string
	headerCtxKey string
	headerCtxVal string
	footerNotice string
	loadingMsg   string
	showHelp     bool

	width         int
	height        int
	contentWidth  int
	contentHeight int

	program                             *tea.Program
	in                                  io.Reader
	out                                 io.Writer
	previousView, currentView, nextView View
	styles                              styles.Theme
	lock                                *sync.RWMutex
}

var tickTime = time.Millisecond * 250

func NewContainer(ctx context.Context, in io.Reader, out io.Writer, styles styles.Theme) *Container {
	c := &Container{
		appName:    "app",
		lock:       &sync.RWMutex{},
		loadingMsg: components.DefaultLoading,
		in:         in,
		out:        out,
	}
	prgm := tea.NewProgram(
		c,
		tea.WithContext(ctx),
		tea.WithInput(c.in),
		tea.WithOutput(c.out),
	)
	c = c.WithProgram(prgm)
	return c
}

func (c *Container) Run() error {
	_, err := c.Program().Run()
	return err
}

func (c *Container) RunAsync(wait bool) {
	go func() {
		_, err := c.Program().Run()
		if err != nil {
			c.HandleError(err)
		}
	}()

	if pc := c.Init(); pc != nil {
		c.Program().Send(pc)
	}

	if wait {
		readyTimout := time.Now().Add(10 * time.Second)
		c.setCurrentView(components.NewLoadingView(c.loadingMsg, c.styles))
		for {
			if c.Ready() {
				break
			} else if time.Now().After(readyTimout) {
				panic("timed out waiting for container to be ready")
			}
			time.Sleep(tickTime)
		}
	}
}

func (c *Container) HandleError(err error) {
	if err == nil {
		return
	}

	cErr := c.SetView(components.NewErrorView(err, c.styles))
	if cErr != nil {
		panic(err)
	}
}

func (c *Container) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	if c.CurrentView() != nil {
		cmds = append(cmds, c.CurrentView().Init())
	}
	cmds = append(
		cmds,
		tea.SetWindowTitle(c.appName),
		tea.Tick(tickTime, func(t time.Time) tea.Msg {
			return components.TickMsg(t)
		}),
	)
	return tea.Batch(cmds...)
}

//nolint:gocognit
func (c *Container) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.QuitMsg:
		return c, tea.Quit
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		c.contentWidth = msg.Width
		c.contentHeight = msg.Height - (styles.HeaderHeight + styles.FooterHeight)
	case components.SubmitMsgType:
		if c.CurrentView().Type() != components.FormViewType {
			return c, nil
		}
		if c.PreviousView() == nil || c.CurrentView() == c.PreviousView() {
			return c, tea.Suspend
		} else {
			c.setNextView(c.PreviousView())
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			c.CurrentView().Update(tea.Quit())
			return c, tea.Quit
		case "esc", "backspace":
			if c.PreviousView() == nil || c.CurrentView() == c.previousView {
				c.CurrentView().Update(tea.Quit())
				return c, tea.Quit
			} else {
				c.setNextView(c.PreviousView())
				return c, nil
			}
		case "h":
			if c.CurrentView().HelpMsg() == "" {
				break
			}
			c.showHelp = !c.showHelp
		}
	case components.TickMsg:
		if c.Ready() && c.NextView() != nil {
			if err := c.SetView(c.NextView()); err != nil {
				c.HandleError(err)
			}
		}
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return components.TickMsg(t)
		}))
	case tea.Cmd:
		cmds = append(cmds, msg)
	}

	_, cmd := c.CurrentView().Update(msg)
	cmds = append(cmds, cmd)
	return c, tea.Batch(cmds...)
}

func (c *Container) View() string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var footer string
	var footerPrefix string

	if !c.Ready() && c.CurrentView().Type() != components.LoadingViewType {
		return ""
	}
	switch {
	case c.CurrentView().Type() == components.FrameViewType, c.CurrentView().Type() == components.FormViewType:
		return c.CurrentView().View()
	case c.CurrentView().Type() == components.LoadingViewType:
		footer = ""
	case !c.CurrentView().ShowFooter():
		footer = c.styles.RenderFooter(c.footerNotice, c.width)
	case c.showHelp:
		footerPrefix = "[ q: quit ] [ h: hide help ] [ ↑/↓: navigate ]"
		if c.PreviousView() != nil {
			footerPrefix += " [ esc: back ]"
		}
		footer = c.styles.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, c.CurrentView().HelpMsg()), c.width)
	case !c.showHelp && c.CurrentView().HelpMsg() != "":
		footerPrefix = "[ q: quit ] [ h: show help ]"
		if c.footerNotice != "" {
			footer = c.styles.RenderFooter(
				fmt.Sprintf("%s ● %s ● %s", footerPrefix, c.CurrentView().HelpMsg(), c.footerNotice), c.width,
			)
		} else {
			footer = c.styles.RenderFooter(footerPrefix, c.width)
		}
	case !c.showHelp && c.CurrentView().HelpMsg() == "":
		footerPrefix = "[ q: quit ]"
		if c.footerNotice != "" {
			footer = c.styles.RenderFooter(fmt.Sprintf("%s ● %s", footerPrefix, c.footerNotice), c.width)
		} else {
			footer = c.styles.RenderFooter(footerPrefix, c.width)
		}
	}

	header := c.styles.RenderHeader(c.appName, c.headerCtxKey, c.headerCtxVal, c.width)
	return lipgloss.JoinVertical(lipgloss.Top, header, c.CurrentView().View(), footer)
}

func (c *Container) Ready() bool {
	loading := c.CurrentView() != nil && c.CurrentView().Type() == components.LoadingViewType
	switch {
	case !c.SizeSet():
		return false
	case loading && c.NextView() == nil:
		return false
	default:
		return true
	}
}

func (c *Container) Suspend() error {
	return c.Program().ReleaseTerminal()
}

func (c *Container) Resume() error {
	return c.Program().RestoreTerminal()
}

func (c *Container) Shutdown() {
	// exit the program
	_ = c.Suspend()
	c.Update(tea.Quit())

	// clear the screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic("unable to clear screen")
	}
}

func (c *Container) WithProgram(p *tea.Program) *Container {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.program = p
	return c
}

func (c *Container) Program() *tea.Program {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.program
}

func (c *Container) Height() int {
	return c.height
}

func (c *Container) ContentHeight() int {
	return c.contentHeight
}

func (c *Container) Width() int {
	return c.width
}

func (c *Container) ContentWidth() int {
	return c.contentWidth
}

func (c *Container) SetView(v View) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.program == nil {
		c.setNextView(v)
		return nil
	}
	switching := c.CurrentView() != nil && c.CurrentView().Type() != v.Type()
	switch {
	case v.Type() == components.LoadingViewType:
		c.setCurrentView(v)
	case !c.Ready():
		c.setNextView(v)
		if c.CurrentView() != nil && c.CurrentView().Type() != components.LoadingViewType {
			c.setCurrentView(components.NewLoadingView(c.loadingMsg, c.styles))
		}
	case switching:
		c.setPreviousView(c.CurrentView())
		c.setCurrentView(v)
	default:
		c.setCurrentView(v)
		c.setNextView(nil)
	}

	cmd := c.CurrentView().Init()
	if cmd != nil {
		c.Program().Send(cmd)
	}
	return nil
}

func (c *Container) setCurrentView(v View) {
	c.currentView = v
}

func (c *Container) CurrentView() View {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.currentView
}

func (c *Container) setPreviousView(v View) {
	c.previousView = v
}

func (c *Container) PreviousView() View {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.previousView
}

func (c *Container) setNextView(v View) {
	c.nextView = v
}

func (c *Container) NextView() View {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.nextView
}

func (c *Container) SizeSet() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.width > 0 && c.height > 0 && c.contentWidth > 0 && c.contentHeight > 0
}

func (c *Container) WithAppName(name string) *Container {
	c.appName = name
	return c
}

func (c *Container) WithLoadingMsg(msg string) *Container {
	c.loadingMsg = msg
	return c
}

func (c *Container) WithNotice(notice string, lvl styles.NoticeLevel) *Container {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.footerNotice = c.styles.RenderNotice(notice, lvl)
	return c
}

func (c *Container) WithHeaderContext(key, val string) *Container {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.headerCtxKey = key
	c.headerCtxVal = val
	return c
}

func (c *Container) HeaderContext() (string, string) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.headerCtxKey, c.headerCtxVal
}
