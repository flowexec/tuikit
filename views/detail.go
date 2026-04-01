package views

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

const DetailViewType = "detail"

// DetailField is a key-value pair displayed in the fixed metadata header.
type DetailField struct {
	Key   string
	Value string
}

// DetailView displays a single item with a fixed metadata header
// (key-value pairs) above a scrollable body box.
type DetailView struct {
	metadata       []DetailField
	body           string
	metadataHeight int

	viewport viewport.Model
	theme    themes.Theme
	width    int
	height   int
}

func NewDetailView(
	state *types.RenderState,
	body string,
	metadata ...DetailField,
) *DetailView {
	v := &DetailView{
		metadata: metadata,
		body:     body,
		theme:    state.Theme,
		width:    state.ContentWidth,
		height:   state.ContentHeight,
	}
	v.syncViewport()
	return v
}

func (v *DetailView) Init() tea.Cmd {
	return nil
}

func (v *DetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *types.RenderState:
		v.width = msg.ContentWidth
		v.height = msg.ContentHeight
		v.theme = msg.Theme
		v.syncViewport()
	case tea.KeyPressMsg:
		halfPage := max(v.viewport.Height()/2, 1)
		switch msg.String() {
		case "k":
			v.viewport.ScrollUp(1)
		case "j":
			v.viewport.ScrollDown(1)
		case "u":
			v.viewport.ScrollUp(halfPage)
		case "d":
			v.viewport.ScrollDown(halfPage)
		case "g":
			v.viewport.GotoTop()
		case "G":
			v.viewport.GotoBottom()
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v *DetailView) View() tea.View {
	metaStr := v.renderMetadata()
	v.viewport.SetContent(v.body)

	var sections []string
	if metaStr != "" {
		sections = append(sections, metaStr)
	}
	sections = append(sections, v.renderBodyBox())

	content := lipgloss.NewStyle().MarginLeft(2).Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
	return tea.View{Content: content}
}

func (v *DetailView) HelpBindings() []themes.HelpKey {
	return []themes.HelpKey{
		{Key: "j/k", Desc: "scroll"},
		{Key: "u/d", Desc: "half-page"},
		{Key: "g/G", Desc: "top/bottom"},
	}
}

func (v *DetailView) Type() string {
	return DetailViewType
}

func (v *DetailView) SetBody(body string) {
	v.body = body
}

func (v *DetailView) SetMetadata(metadata ...DetailField) {
	v.metadata = metadata
	v.syncViewport()
}

func (v *DetailView) syncViewport() {
	v.metadataHeight = v.calcMetadataHeight()
	// Body box border (2) + padding (2) are chrome around the viewport
	bodyChrome := 4
	vpHeight := max(v.height-v.metadataHeight-bodyChrome, 1)

	v.viewport.SetHeight(vpHeight)
	v.viewport.SetWidth(v.width - 10) // account for margin (2) + border (2) + padding (4) + buffer (2)
}

func (v *DetailView) calcMetadataHeight() int {
	if len(v.metadata) == 0 {
		return 0
	}
	// rows + border top/bottom (2) + padding top/bottom (2) + margin bottom (1)
	return len(v.metadata) + 2 + 2 + 1
}

func (v *DetailView) renderMetadata() string {
	if len(v.metadata) == 0 {
		return ""
	}

	cp := v.theme.ColorPalette()
	maxKeyLen := 0
	for _, f := range v.metadata {
		if len(f.Key) > maxKeyLen {
			maxKeyLen = len(f.Key)
		}
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(cp.SecondaryColor()).
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
	for _, f := range v.metadata {
		row := keyStyle.Render(f.Key) + " " + sep + valStyle.Render(f.Value)
		rows = append(rows, row)
	}

	tableWidth := min(v.width-6, 60)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cp.BorderColor()).
		Padding(1, 1).
		Width(tableWidth).
		MarginBottom(1).
		Render(strings.Join(rows, "\n"))
}

func (v *DetailView) renderBodyBox() string {
	cp := v.theme.ColorPalette()
	bodyWidth := v.width - 4
	// viewport.View() returns the scrolled content; wrap it in the box
	vpContent := v.viewport.View()

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cp.BorderColor()).
		Foreground(cp.BodyColor()).
		Padding(1, 2).
		Width(bodyWidth).
		Render(vpContent)
}
