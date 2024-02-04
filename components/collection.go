package components

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/jahvon/tuikit/styles"
	"github.com/jahvon/tuikit/types"
)

type CollectionView struct {
	collection types.Collection

	model *list.Model
	items []list.Item
	err   TeaModel

	format        Format
	width, height int
	styles        styles.Theme
	callbacks     []KeyCallback
	selectedFunc  func(header string) error
}

func NewCollectionView(
	state *TerminalState,
	collection types.Collection,
	format Format,
	selectedFunc func(header string) error,
	keys ...KeyCallback,
) TeaModel {
	if format == "" {
		format = FormatDocument
	}
	items := make([]list.Item, 0)
	for _, item := range collection.Items() {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})
	delegate := list.NewDefaultDelegate()
	delegate.Styles = state.Theme.ListItemStyles()

	model := list.New(items, delegate, state.Width, state.Height)
	model.SetShowTitle(false)
	model.SetShowHelp(false)
	model.SetShowPagination(false)
	model.SetStatusBarItemName(collection.Singular(), collection.Plural())
	model.Styles = state.Theme.ListStyles()
	return &CollectionView{
		collection:   collection,
		model:        &model,
		items:        items,
		format:       format,
		width:        state.Width,
		height:       state.Height,
		styles:       state.Theme,
		selectedFunc: selectedFunc,
		callbacks:    keys,
	}
}

func (v *CollectionView) Init() tea.Cmd {
	return nil
}

//nolint:gocognit
func (v *CollectionView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.model.SetSize(v.width, v.height)
	case tea.KeyMsg:
		switch msg.String() {
		case "-", "d":
			if v.format == FormatDocument {
				return v, nil
			}
			v.format = FormatDocument
		case "y":
			if v.format == FormatYAML {
				return v, nil
			}
			v.format = FormatYAML
		case "j":
			if v.format == FormatJSON {
				return v, nil
			}
			v.format = FormatJSON
		case tea.KeyEnter.String():
			if v.selectedFunc == nil {
				return v, nil
			}
			selected := v.model.SelectedItem()
			if selected == nil {
				return v, nil
			}

			if err := v.selectedFunc(selected.FilterValue()); err != nil {
				v.err = NewErrorView(err, v.styles)
			}
			return v, nil
		default:
			for _, cb := range v.callbacks {
				if cb.Key == msg.String() {
					if err := cb.Callback(); err != nil {
						v.err = NewErrorView(err, v.styles)
					}
					return v, nil
				}
			}
		}
	}

	model, cmd := v.model.Update(msg)
	v.model = &model
	return v, cmd
}

func (v *CollectionView) UpdateItemsFromCollections() {
	items := make([]list.Item, 0)
	for _, item := range v.collection.Items() {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].FilterValue() < items[j].FilterValue()
	})
	v.items = items
}

func (v *CollectionView) Items() []list.Item {
	return v.model.Items()
}

func (v *CollectionView) renderedContent() string {
	var content string
	var isMkdwn bool
	var err error
	switch v.format {
	case FormatYAML:
		content, err = v.collection.YAML()
		content = fmt.Sprintf("```yaml\n%s\n```", content)
		isMkdwn = true
	case FormatJSON:
		content, err = v.collection.JSON()
		content = fmt.Sprintf("```json\n%s\n```", content)
		isMkdwn = true
	case FormatDocument:
		fallthrough
	default:
		v.model.SetSize(v.width, v.height)
		v.UpdateItemsFromCollections()
		style := v.styles.Collection().Width(v.width)
		content = style.Render(v.model.View())
	}
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	if content == "" {
		content = "no data"
	}

	if !isMkdwn {
		return content
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(v.styles.MarkdownStyleJSON)),
		glamour.WithWordWrap(v.width-2),
	)
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}

	viewStr, err := renderer.Render(content)
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	return viewStr
}

func (v *CollectionView) View() string {
	if v.err != nil {
		return v.err.View()
	}

	return v.renderedContent()
}

func (v *CollectionView) HelpMsg() string {
	var selectHelp string
	if v.selectedFunc != nil {
		selectHelp = "enter: select • "
	}
	msg := fmt.Sprintf("%s/: filter | d: docs • y: yaml • j: json", selectHelp)

	var extendedHelp string
	for _, cb := range v.callbacks {
		if cb.Key == "" || cb.Label == "" {
			continue
		}
		extendedHelp += fmt.Sprintf(" • %s: %s", cb.Key, cb.Label)
	}
	if extendedHelp != "" {
		msg = fmt.Sprintf("%s | %s", extendedHelp, msg)
	}
	return msg
}

func (v *CollectionView) Interactive() bool {
	return v.err == nil
}

func (v *CollectionView) Type() string {
	return "collection"
}
