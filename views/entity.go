package views

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/v2/viewport"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/jahvon/glamour"

	"github.com/jahvon/tuikit/themes"
	"github.com/jahvon/tuikit/types"
)

type EntityView struct {
	entity types.Entity

	viewport viewport.Model
	err      *ErrorView

	styles        themes.Theme
	width, height int
	format        types.Format
	callbacks     []types.KeyCallback
}

func NewEntityView(
	state *types.RenderState,
	entity types.Entity,
	format types.Format,
	keys ...types.KeyCallback,
) *EntityView {
	switch format {
	case "yaml", "yml", "YAML", "YML":
		format = types.EntityFormatYAML
	case "json", "JSON":
		format = types.EntityFormatJSON
	case "doc", "md", "text", "markdown":
		format = types.EntityFormatDocument
	default:
		format = types.EntityFormatDocument
	}

	vp := viewport.New(viewport.WithWidth(state.ContentWidth), viewport.WithHeight(state.ContentHeight))
	vp.Style = state.Theme.EntityViewStyle().Width(state.ContentWidth)
	return &EntityView{
		entity:    entity,
		styles:    state.Theme,
		width:     state.ContentWidth,
		height:    state.ContentHeight,
		format:    format,
		callbacks: keys,
		viewport:  vp,
	}
}

func (v *EntityView) Init() (tea.Model, tea.Cmd) {
	return v, nil
}

//nolint:gocognit
func (v *EntityView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)

	switch msg := msg.(type) {
	case types.RenderState:
		v.viewport.SetWidth(msg.ContentWidth)
		v.viewport.SetHeight(msg.ContentHeight)
		v.viewport.SetContent(v.renderedContent())
	case tea.KeyMsg:
		switch msg.String() {
		case "-", "d":
			if v.format == types.EntityFormatDocument {
				return v, nil
			}
			v.format = types.EntityFormatDocument
			v.viewport.GotoTop()
		case "y":
			if v.format == types.CollectionFormatYAML {
				return v, nil
			}
			v.format = types.CollectionFormatYAML
			v.viewport.GotoTop()
		case "j":
			if v.format == types.CollectionFormatJSON {
				return v, nil
			}
			v.format = types.CollectionFormatJSON
			v.viewport.GotoTop()
		case "up":
			v.viewport.LineUp(1)
		case "down":
			v.viewport.LineDown(1)
		default:
			for _, cb := range v.callbacks {
				if cb.Key == msg.String() {
					if err := cb.Callback(); err != nil {
						v.err = NewErrorView(err, v.styles)
					}
				}
			}
		}
	}

	return v, cmd
}

func (v *EntityView) renderedContent() string {
	var content string
	var err error
	switch v.format {
	case types.CollectionFormatYAML:
		content, err = v.entity.YAML()
		content = fmt.Sprintf("```yaml\n%s\n```", content)
	case types.CollectionFormatJSON:
		content, err = v.entity.JSON()
		content = fmt.Sprintf("```json\n%s\n```", content)
	case types.EntityFormatDocument:
		content = v.entity.Markdown()
	case types.CollectionFormatList:
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

	mdStyles, err := v.styles.GlamourMarkdownStyleJSON()
	if err != nil {
		v.err = NewErrorView(err, v.styles)
		return v.err.View()
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(mdStyles)),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(int(math.Floor(float64(v.width)*0.95))),
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

func (v *EntityView) View() string {
	if v.err != nil {
		return v.err.View()
	}
	v.viewport.SetContent(v.renderedContent())
	return v.viewport.View()
}

func (v *EntityView) HelpMsg() string {
	msg := "[ d: docs ] [ y: yaml ] [ j: json ]"

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

func (v *EntityView) ShowFooter() bool {
	return v.err == nil
}

func (v *EntityView) Type() string {
	return "entity"
}
