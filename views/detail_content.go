package views

import (
	"fmt"
	"math"
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

const DetailContentViewType = "detail-content"

// DetailContentOpts configures the sections of a DetailContentView.
type DetailContentOpts struct {
	Title    string        // primary heading (entity name/ref)
	Subtitle string        // secondary label (type name)
	Tags     string        // pre-rendered lipgloss tag string
	Metadata []DetailField // key-value pairs for the metadata section
	Body     string        // markdown rendered via glamour
	Footer   string        // faint footer line (e.g. file path)
	Entity   types.Entity  // optional: enables y/j/d format switching
}

// DetailContentView combines a lipgloss-styled header and metadata section
// with a glamour-rendered scrollable body. It provides format switching
// between the structured view and raw YAML/JSON when an Entity is set.
type DetailContentView struct {
	opts DetailContentOpts

	viewport  viewport.Model
	theme     themes.Theme
	width     int
	height    int
	format    types.Format
	callbacks []types.KeyCallback
	err       *ErrorView
}

func NewDetailContentView(
	state *types.RenderState,
	opts DetailContentOpts,
	keys ...types.KeyCallback,
) *DetailContentView {
	vp := viewport.New(viewport.WithWidth(state.ContentWidth), viewport.WithHeight(state.ContentHeight))
	v := &DetailContentView{
		opts:      opts,
		theme:     state.Theme,
		width:     state.ContentWidth,
		height:    state.ContentHeight,
		format:    types.EntityFormatDocument,
		callbacks: keys,
		viewport:  vp,
	}
	v.syncViewport()
	return v
}

func (v *DetailContentView) Init() tea.Cmd {
	return nil
}

func (v *DetailContentView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.err != nil {
		return v.err.Update(msg)
	}

	switch msg := msg.(type) {
	case *types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.theme = msg.Theme
		v.syncViewport()
	case types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.theme = msg.Theme
		v.syncViewport()
	case tea.KeyPressMsg:
		if v.handleKeyPress(msg) {
			return v, nil
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *DetailContentView) handleKeyPress(msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "-", "d":
		if v.format != types.EntityFormatDocument {
			v.format = types.EntityFormatDocument
			v.viewport.GotoTop()
		}
		return true
	case "y":
		if v.opts.Entity != nil && v.format != types.EntityFormatYAML {
			v.format = types.EntityFormatYAML
			v.viewport.GotoTop()
		}
		return true
	case "j":
		if v.opts.Entity != nil && v.format != types.EntityFormatJSON {
			v.format = types.EntityFormatJSON
			v.viewport.GotoTop()
		}
		return true
	case types.KeyUp:
		v.viewport.ScrollUp(1)
	case types.KeyDown:
		v.viewport.ScrollDown(1)
	default:
		for _, cb := range v.callbacks {
			if cb.Key == msg.String() {
				if err := cb.Callback(); err != nil {
					v.err = NewErrorView(err, v.theme)
				}
				return true
			}
		}
	}
	return false
}

func (v *DetailContentView) View() tea.View {
	if v.err != nil {
		return v.err.View()
	}

	headerBox := v.renderCombinedHeader()

	var body string
	if v.format != types.EntityFormatDocument && v.opts.Entity != nil {
		body = v.renderEntityFormat()
	} else {
		body = v.renderBody()
	}

	fixedHeight := lipgloss.Height(headerBox)

	var footerBox string
	if v.opts.Footer != "" {
		footerBox = strings.TrimRight(v.renderMarkdown(v.opts.Footer), "\n")
		fixedHeight += lipgloss.Height(footerBox)
	}

	bodyChrome := 4 // border (2) + padding (2)
	vpHeight := max(v.height-fixedHeight-bodyChrome, 3)
	v.viewport.SetHeight(vpHeight)

	bodyWidth := v.width - 4
	v.viewport.SetWidth(bodyWidth - 6) // border (2) + padding (4)
	v.viewport.SetContent(body)

	bodyBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(v.theme.ColorPalette().BorderColor()).
		Foreground(v.theme.ColorPalette().BodyColor()).
		Padding(1, 2).
		Width(bodyWidth).
		Render(v.viewport.View())

	sections := []string{}
	if headerBox != "" {
		sections = append(sections, headerBox)
	}
	sections = append(sections, bodyBox)
	if footerBox != "" {
		sections = append(sections, footerBox)
	}

	content := lipgloss.NewStyle().MarginLeft(2).Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
	return tea.View{Content: content}
}

func (v *DetailContentView) HelpBindings() []themes.HelpKey {
	if v.err != nil {
		return nil
	}
	keys := make([]themes.HelpKey, 0, len(v.callbacks)+3)
	for _, cb := range v.callbacks {
		if cb.Key != "" && cb.Label != "" {
			keys = append(keys, themes.HelpKey{Key: cb.Key, Desc: cb.Label})
		}
	}
	keys = append(keys, themes.HelpKey{Key: "↑/↓", Desc: "scroll"})
	if v.opts.Entity != nil {
		keys = append(keys,
			themes.HelpKey{Key: "d", Desc: "detail"},
			themes.HelpKey{Key: "y", Desc: "yaml"},
			themes.HelpKey{Key: "j", Desc: "json"},
		)
	}
	return keys
}

func (v *DetailContentView) Type() string {
	return DetailContentViewType
}

// renderCombinedHeader produces two boxes (title and metadata) side-by-side.
func (v *DetailContentView) renderCombinedHeader() string {
	cp := v.theme.ColorPalette()
	headerWidth := v.width - 4

	var leftLines []string
	if v.opts.Title != "" {
		leftLines = append(leftLines, lipgloss.NewStyle().Foreground(cp.PrimaryColor()).Bold(true).Render(v.opts.Title))
	}
	if v.opts.Subtitle != "" {
		leftLines = append(leftLines, lipgloss.NewStyle().Foreground(cp.GrayColor()).Italic(true).Render(v.opts.Subtitle))
	}
	if v.opts.Tags != "" {
		leftLines = append(leftLines, v.opts.Tags)
	}

	var hasLeft = len(leftLines) > 0
	var hasRight = len(v.opts.Metadata) > 0

	if !hasLeft && !hasRight {
		return ""
	}

	leftStr := strings.Join(leftLines, "\n")
	var rightStr string

	if hasRight {
		maxKeyLen := 0
		for _, f := range v.opts.Metadata {
			if len(f.Key) > maxKeyLen {
				maxKeyLen = len(f.Key)
			}
		}

		keyStyle := lipgloss.NewStyle().
			Foreground(cp.GrayColor()).
			Bold(true).
			Width(maxKeyLen + 1).
			Align(lipgloss.Right)
		valStyle := lipgloss.NewStyle().
			Foreground(cp.BodyColor()).
			PaddingLeft(1)
		sep := lipgloss.NewStyle().
			Foreground(cp.GrayColor()).
			Render("│")

		var rows []string
		for _, f := range v.opts.Metadata {
			rows = append(rows, keyStyle.Render(f.Key)+" "+sep+valStyle.Render(f.Value))
		}
		rightStr = strings.Join(rows, "\n")
	}

	leftW := headerWidth
	rightW := 0
	if hasLeft && hasRight {
		leftW = (headerWidth / 2) - 1 // Leave a 1 cell gap
		rightW = headerWidth - leftW - 1
	} else if hasRight {
		rightW = headerWidth
	}

	maxInnerH := max(lipgloss.Height(leftStr), lipgloss.Height(rightStr))

	var boxes []string

	if hasLeft {
		style := lipgloss.NewStyle().
			Padding(0, 1). // Minimal padding for compactness
			Width(max(0, leftW)).
			Height(maxInnerH)
		if hasRight {
			style = style.MarginRight(1)
		}
		boxes = append(boxes, style.Render(leftStr))
	}

	if hasRight {
		style := lipgloss.NewStyle().
			Padding(0, 1). // Align seamlessly with left side since neither has borders
			Width(max(0, rightW)).
			Height(maxInnerH)
		boxes = append(boxes, style.Render(rightStr))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, boxes...)
	return lipgloss.NewStyle().MarginBottom(1).Render(content)
}

// renderBody renders the markdown body through glamour.
func (v *DetailContentView) renderBody() string {
	return v.renderMarkdown(v.opts.Body)
}

// renderEntityFormat renders the entity in YAML or JSON format through glamour.
func (v *DetailContentView) renderEntityFormat() string {
	if v.opts.Entity == nil {
		return ""
	}
	var content string
	var err error
	switch v.format {
	case types.EntityFormatYAML:
		content, err = v.opts.Entity.YAML()
		content = fmt.Sprintf("```yaml\n%s\n```", content)
	case types.EntityFormatJSON:
		content, err = v.opts.Entity.JSON()
		content = fmt.Sprintf("```json\n%s\n```", content)
	case types.CollectionFormatList, types.EntityFormatDocument:
		return ""
	}
	if err != nil {
		v.err = NewErrorView(err, v.theme)
		return v.err.View().Content
	}
	return v.renderMarkdown(content)
}

func (v *DetailContentView) renderMarkdown(content string) string {
	if content == "" {
		return "no content"
	}
	mdStyles, err := v.theme.GlamourMarkdownStyleJSON()
	if err != nil {
		return content
	}
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(mdStyles)),
		glamour.WithPreservedNewLines(),
		glamour.WithWordWrap(int(math.Floor(float64(v.width)*0.85))),
	)
	if err != nil {
		return content
	}
	rendered, err := renderer.Render(content)
	if err != nil {
		return content
	}
	return rendered
}

func (v *DetailContentView) syncViewport() {
	headerHeight := v.calcCombinedHeaderHeight()
	footerHeight := 0
	if v.opts.Footer != "" {
		footerHeight = 2 // Text (1) + margin top (1)
	}
	bodyChrome := 4
	vpHeight := max(v.height-headerHeight-footerHeight-bodyChrome, 3)
	bodyWidth := v.width - 4

	v.viewport.SetHeight(vpHeight)
	v.viewport.SetWidth(bodyWidth - 6)
}

func (v *DetailContentView) calcCombinedHeaderHeight() int {
	leftLines := 0
	if v.opts.Title != "" {
		leftLines++
	}
	if v.opts.Subtitle != "" {
		leftLines++
	}
	if v.opts.Tags != "" {
		leftLines++
	}

	rightLines := len(v.opts.Metadata)

	lines := max(leftLines, rightLines)
	if lines == 0 {
		return 0
	}
	// margin bottom (1) -> 1
	return lines + 1
}
