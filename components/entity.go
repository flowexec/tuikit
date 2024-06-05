package components

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jahvon/glamour"

	"github.com/jahvon/tuikit/styles"
	"github.com/jahvon/tuikit/types"
)

type EntityView struct {
	entity types.Entity

	viewport viewport.Model
	err      TeaModel

	styles        styles.Theme
	width, height int
	format        Format
	callbacks     []KeyCallback
}

func NewEntityView(
	state *TerminalState,
	entity types.Entity,
	format Format,
	keys ...KeyCallback,
) TeaModel {
	if format == "" {
		format = FormatDocument
	}
	vp := viewport.New(state.Width, state.Height)
	vp.Style = state.Theme.EntityView().Width(state.Width)
	return &EntityView{
		entity:    entity,
		styles:    state.Theme,
		width:     state.Width,
		height:    state.Height,
		format:    format,
		callbacks: keys,
		viewport:  vp,
	}
}

func (v *EntityView) Init() tea.Cmd {
	return nil
}

//nolint:gocognit
func (v *EntityView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}
	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.viewport.Width = msg.Width
		v.viewport.Height = msg.Height - (styles.HeaderHeight + styles.FooterHeight)
		v.viewport.SetContent(v.renderedContent())
	case tea.KeyMsg:
		switch msg.String() {
		case "-", "d":
			if v.format == FormatDocument {
				return v, nil
			}
			v.format = FormatDocument
			v.viewport.GotoTop()
		case "y":
			if v.format == FormatYAML {
				return v, nil
			}
			v.format = FormatYAML
			v.viewport.GotoTop()
		case "j":
			if v.format == FormatJSON {
				return v, nil
			}
			v.format = FormatJSON
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
	case FormatYAML:
		content, err = v.entity.YAML()
		content = fmt.Sprintf("```yaml\n%s\n```", content)
	case FormatJSON:
		content, err = v.entity.JSON()
		content = fmt.Sprintf("```json\n%s\n```", content)
	case FormatDocument:
		content = v.entity.Markdown()
	case FormatList:
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

	mdStyles, err := v.styles.MarkdownStyleJSON()
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
		msg = fmt.Sprintf("%s • %s", extendedHelp, msg)
	}
	return msg
}

func (v *EntityView) Interactive() bool {
	return v.err == nil
}

func (v *EntityView) Type() string {
	return "entity"
}
