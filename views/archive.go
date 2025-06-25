package views

import (
	"errors"
	"fmt"
	stdIO "io"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"github.com/jahvon/tuikit/io"
	"github.com/jahvon/tuikit/themes"
	"github.com/jahvon/tuikit/types"
)

type LogArchiveView struct {
	archiveDir    string
	cachedEntries []io.ArchiveEntry

	model       *list.Model
	items       []list.Item
	activeEntry *io.ArchiveEntry
	err         *ErrorView

	width, height int
	styles        themes.Theme
}

func NewLogArchiveView(state *types.RenderState, archiveDir string, lastEntry bool) *LogArchiveView {
	items := make([]list.Item, 0)
	entries, err := io.ListArchiveEntries(archiveDir)
	var errView *ErrorView
	if err != nil {
		errView = NewErrorView(err, state.Theme)
		return &LogArchiveView{err: errView}
	}
	slices.Reverse(entries)
	for _, entry := range entries {
		items = append(items, entry)
	}
	delegate := &logArchiveDelegate{theme: state.Theme}
	model := list.New(items, delegate, state.Width, state.Height)
	model.SetShowTitle(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetStatusBarItemName("log entry", "log entries")
	model.Styles = state.Theme.ListStyles()

	var lastEntryFile *io.ArchiveEntry
	if lastEntry {
		lastEntryFile = &entries[0]
	}
	return &LogArchiveView{
		archiveDir:    archiveDir,
		cachedEntries: entries,
		activeEntry:   lastEntryFile,
		model:         &model,
		items:         items,
		width:         state.ContentWidth,
		height:        state.ContentHeight,
		styles:        state.Theme,
	}
}

func (v *LogArchiveView) Init() tea.Cmd {
	return nil
}

//nolint:gocognit,funlen
func (v *LogArchiveView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	switch msg := msg.(type) {
	case types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.model.SetSize(v.width, v.height)
	case types.TickMsg:
		if v.activeEntry != nil {
			time.Sleep(time.Second)
			return v, tea.Quit
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "x":
			if v.activeEntry != nil {
				return v, nil
			}
			for _, entry := range v.cachedEntries {
				if err := io.DeleteArchiveEntry(entry.Path); err != nil {
					v.err = NewErrorView(err, v.styles)
				}
			}
			v.items = nil
			v.model.SetItems(v.items)
		case "d":
			if v.activeEntry != nil {
				return v, nil
			}
			selected := v.model.SelectedItem()
			if selected == nil {
				return v, nil
			}
			var selectedEntry *io.ArchiveEntry
			for i, entry := range v.cachedEntries {
				if entry.FilterValue() == selected.FilterValue() {
					v.cachedEntries = append(v.cachedEntries[:i], v.cachedEntries[i+1:]...)
					selectedEntry = &v.cachedEntries[i]
					break
				}
			}
			if err := io.DeleteArchiveEntry(selectedEntry.Path); err != nil {
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
			if v.activeEntry != nil {
				return v, nil
			}
			selected := v.model.SelectedItem()
			if selected == nil {
				return v, nil
			}
			for i, entry := range v.cachedEntries {
				if entry.FilterValue() == selected.FilterValue() {
					v.activeEntry = &v.cachedEntries[i]
					return v, nil
				}
			}
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
	case v.activeEntry != nil:
		content, err = v.activeEntry.Read()
		if err != nil {
			v.err = NewErrorView(err, v.styles)
			return v.err.View()
		} else if content == "" {
			content = "\nno data found in log entry\n"
		}
		content = wordwrap.String("\n"+content+"\n", v.width)
	case len(v.items) == 0:
		v.err = NewErrorView(errors.New("no log entries found"), v.styles)
		return v.err.View()
	default:
		v.model.SetSize(v.width, v.height)
		style := v.styles.BoxStyle().Width(v.width)
		content = style.Render(v.model.View())
	}
	return content
}

func (v *LogArchiveView) HelpMsg() string {
	return "[ enter: select ] [ /: filter ] ● [ d: delete selected ] [ x: delete all ]"
}

func (v *LogArchiveView) ShowFooter() bool {
	return v.err == nil && v.activeEntry == nil
}

func (v *LogArchiveView) Type() string {
	return "log-archive"
}

type logArchiveDelegate struct {
	theme themes.Theme
}

func (d *logArchiveDelegate) Height() int                             { return 1 }
func (d *logArchiveDelegate) Spacing() int                            { return 0 }
func (d *logArchiveDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d *logArchiveDelegate) Render(w stdIO.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(io.ArchiveEntry)
	if !ok {
		return
	}
	title := fmt.Sprintf("%d. %s", index+1, i.Title())
	description := i.Description()
	titleStyle := lipgloss.NewStyle().Foreground(d.theme.ColorPalette().WhiteColor()).PaddingLeft(2).Render
	descriptionStyle := lipgloss.NewStyle().Foreground(d.theme.ColorPalette().WhiteColor()).Render
	if index == m.Index() {
		titleStyle = func(s ...string) string {
			return lipgloss.NewStyle().
				Foreground(d.theme.ColorPalette().SecondaryColor()).
				BorderForeground(d.theme.ColorPalette().SecondaryColor()).
				BorderLeft(true).
				PaddingLeft(2).
				Render("" + strings.Join(s, " "))
		}
		descriptionStyle = lipgloss.NewStyle().Foreground(d.theme.ColorPalette().SecondaryColor()).Render
	}
	itemStr := titleStyle(title) + descriptionStyle(fmt.Sprintf(" (%s)", description))
	_, _ = fmt.Fprint(w, itemStr)
}
