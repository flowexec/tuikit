package views

import (
	"fmt"
	"sort"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

type CollectionView struct {
	collection types.Collection

	model *list.Model
	items []list.Item
	err   *ErrorView

	format        types.Format
	width, height int
	styles        themes.Theme
	callbacks     []types.KeyCallback
	selectedFunc  func(header string) error
}

func NewCollectionView(
	state *types.RenderState,
	collection types.Collection,
	format types.Format,
	selectedFunc func(header string) error,
	keys ...types.KeyCallback,
) *CollectionView {
	//nolint:exhaustive
	switch format {
	case "yaml", "yml", "YAML", "YML":
		format = types.CollectionFormatYAML
	case "json", "JSON":
		format = types.CollectionFormatJSON
	case "list", "ls", "browse":
		format = types.CollectionFormatList
	default:
		format = types.CollectionFormatList
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
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

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
		width:        state.ContentWidth,
		height:       state.ContentHeight,
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
	case types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.model.SetSize(v.width, v.height)
	case tea.KeyPressMsg:
		// When the filter input is active, pass all keys through to the list.
		if v.model.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "-", "l":
			if v.format == types.CollectionFormatList {
				return v, nil
			}
			v.format = types.CollectionFormatList
		case "y":
			if v.format == types.CollectionFormatYAML {
				return v, nil
			}
			v.format = types.CollectionFormatYAML
		case "j":
			if v.format == types.CollectionFormatJSON {
				return v, nil
			}
			v.format = types.CollectionFormatJSON
		case types.KeyEnter:
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

func (v *CollectionView) renderedView() tea.View {
	var content string
	var isMkdwn bool
	var err error
	switch v.format {
	case types.CollectionFormatYAML:
		content, err = v.collection.YAML()
		content = fmt.Sprintf("```yaml\n%s\n```", content)
		isMkdwn = true
	case types.CollectionFormatJSON:
		content, err = v.collection.JSON()
		content = fmt.Sprintf("```json\n%s\n```", content)
		isMkdwn = true
	case types.CollectionFormatList:
		v.model.SetSize(v.width, v.height)
		v.UpdateItemsFromCollections()
		style := v.styles.CollectionStyle().Width(v.width)
		content = style.Render(v.model.View())
	case types.EntityFormatDocument:
		fallthrough
	default:
		content = "unsupported format"
	}
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	if content == "" {
		content = "no data"
	}

	if !isMkdwn {
		return tea.View{Content: content}
	}

	mdStyles, err := v.styles.GlamourMarkdownStyleJSON()
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(mdStyles)),
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
	return tea.View{Content: viewStr}
}

func (v *CollectionView) View() tea.View {
	if v.err != nil {
		return v.err.View()
	}

	return v.renderedView()
}

func (v *CollectionView) HelpBindings() []themes.HelpKey {
	if v.err != nil {
		return nil
	}
	keys := make([]themes.HelpKey, 0)
	for _, cb := range v.callbacks {
		if cb.Key != "" && cb.Label != "" {
			keys = append(keys, themes.HelpKey{Key: cb.Key, Desc: cb.Label})
		}
	}
	if v.selectedFunc != nil {
		keys = append(keys, themes.HelpKey{Key: "enter", Desc: "select"})
	}
	keys = append(keys,
		themes.HelpKey{Key: "/", Desc: "filter"},
		themes.HelpKey{Key: "l", Desc: "list"},
		themes.HelpKey{Key: "y", Desc: "yaml"},
		themes.HelpKey{Key: "j", Desc: "json"},
	)
	return keys
}

func (v *CollectionView) CapturingInput() bool {
	return v.model.FilterState() == list.Filtering
}

func (v *CollectionView) Type() string {
	return "collection"
}
