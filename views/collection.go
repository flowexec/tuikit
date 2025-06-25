package views

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jahvon/glamour"

	"github.com/jahvon/tuikit/themes"
	"github.com/jahvon/tuikit/types"
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
	case tea.KeyMsg:
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
		return content
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
		selectHelp = "[ enter: select ] "
	}
	msg := fmt.Sprintf("%s[ /: filter ] [ d: docs ] [ y: yaml ] [ j: json ]", selectHelp)

	var extendedHelp string
	for i, cb := range v.callbacks {
		switch {
		case cb.Key == "" || cb.Label == "":
			continue
		case i == 0:
			extendedHelp += fmt.Sprintf("[ %s: %s ]", cb.Key, cb.Label)
		default:
			extendedHelp += fmt.Sprintf(" [ %s: %s ]", cb.Key, cb.Label)
		}
	}
	if extendedHelp != "" {
		msg = fmt.Sprintf("%s ● %s", extendedHelp, msg)
	}
	return msg
}

func (v *CollectionView) ShowFooter() bool {
	return v.err == nil
}

func (v *CollectionView) Type() string {
	return "collection"
}
