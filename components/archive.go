package components

import (
	"errors"
	"fmt"
	stdIO "io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jahvon/tuikit/io"
	"github.com/jahvon/tuikit/styles"
)

type LogArchiveView struct {
	archiveDir    string
	cachedEntries []string

	model       *list.Model
	items       []list.Item
	activeEntry string
	err         TeaModel

	width, height int
	styles        styles.Theme
}

func NewLogArchiveView(state *TerminalState, archiveDir string, lastEntry bool) TeaModel {
	items := make([]list.Item, 0)
	entries, err := io.ListArchiveEntries(archiveDir)
	if err != nil {
		return NewErrorView(err, state.Theme)
	}
	slices.Reverse(entries)
	for _, entry := range entries {
		items = append(items, &logArchiveItem{archivePath: entry})
	}
	delegate := &logArchiveDelegate{styles: state.Theme}
	model := list.New(items, delegate, state.Width, state.Height)
	model.SetShowTitle(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetStatusBarItemName("log entry", "log entries")
	model.Styles = state.Theme.ListStyles()

	var lastEntryFile string
	if lastEntry {
		lastEntryFile = entries[0]
	}
	return &LogArchiveView{
		archiveDir:    archiveDir,
		cachedEntries: entries,
		activeEntry:   lastEntryFile,
		model:         &model,
		items:         items,
		width:         state.Width,
		height:        state.Height,
		styles:        state.Theme,
	}
}

func (v *LogArchiveView) Init() tea.Cmd {
	return nil
}

//nolint:gocognit
func (v *LogArchiveView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.model.SetSize(v.width, v.height)
	case TickMsg:
		if v.activeEntry != "" {
			time.Sleep(time.Second)
			return v, tea.Quit
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "x":
			if v.activeEntry != "" {
				return v, nil
			}
			for _, entry := range v.cachedEntries {
				if err := io.DeleteArchiveEntry(entry); err != nil {
					v.err = NewErrorView(err, v.styles)
				}
			}
			v.items = nil
			v.model.SetItems(v.items)
		case "d":
			if v.activeEntry != "" {
				return v, nil
			}
			selected := v.model.SelectedItem()
			if selected == nil {
				return v, nil
			}
			if err := io.DeleteArchiveEntry(selected.FilterValue()); err != nil {
				v.err = NewErrorView(err, v.styles)
			}
			for i, entry := range v.items {
				if entry.FilterValue() == selected.FilterValue() {
					v.items = append(v.items[:i], v.items[i+1:]...)
					break
				}
			}
			v.model.SetItems(v.items)
		case tea.KeyEnter.String():
			if v.activeEntry != "" {
				return v, nil
			}
			selected := v.model.SelectedItem()
			if selected == nil {
				return v, nil
			}
			v.activeEntry = selected.FilterValue()
			return v, nil
		}
	}

	model, cmd := v.model.Update(msg)
	v.model = &model
	return v, cmd
}

func (v *LogArchiveView) View() string {
	if v.err != nil {
		return v.err.View()
	}

	var content string
	var err error
	switch {
	case v.activeEntry != "":
		content, err = io.ReadArchiveEntry(v.activeEntry)
		if err != nil {
			v.err = NewErrorView(err, v.styles)
			return v.err.View()
		} else if content == "" {
			content = "\nno data found in log entry\n"
		}
		content = "\n" + content + "\n"
	case len(v.cachedEntries) == 0:
		v.err = NewErrorView(errors.New("no log entries found"), v.styles)
		return v.err.View()
	default:
		v.model.SetSize(v.width, v.height)
		style := v.styles.Box().Width(v.width)
		content = style.Render(v.model.View())
	}
	return content
}

func (v *LogArchiveView) HelpMsg() string {
	return "enter: select • /: filter | d: delete selected • x: delete all"
}

func (v *LogArchiveView) Interactive() bool {
	return v.err == nil && v.activeEntry == ""
}

func (v *LogArchiveView) Type() string {
	return "log-archive"
}

type logArchiveItem struct{ archivePath string }

type logArchiveDelegate struct {
	styles styles.Theme
}

func (d *logArchiveDelegate) Height() int                             { return 1 }
func (d *logArchiveDelegate) Spacing() int                            { return 0 }
func (d *logArchiveDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d *logArchiveDelegate) Render(w stdIO.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(*logArchiveItem)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i.Title())
	fn := lipgloss.NewStyle().Foreground(d.styles.White).PaddingLeft(2).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().
				Foreground(d.styles.SecondaryColor).
				BorderForeground(d.styles.SecondaryColor).
				BorderLeft(true).
				PaddingLeft(2).
				Render("" + strings.Join(s, " "))
		}
	}
	_, _ = fmt.Fprint(w, fn(str))
}

func (l *logArchiveItem) Title() string {
	name := strings.TrimSuffix(filepath.Base(l.archivePath), ".log")
	nameTime, err := time.Parse(io.LogEntryTimeFormat, name)
	if err != nil {
		return name
	}
	return nameTime.Format("03:04PM 01/02/2006")
}
func (l *logArchiveItem) Description() string { return "" }
func (l *logArchiveItem) FilterValue() string { return l.archivePath }
