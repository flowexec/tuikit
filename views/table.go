package views

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/flowexec/tuikit/types"
)

const TableViewType = "table"

type TableDisplayMode int

const (
	TableDisplayFull TableDisplayMode = iota
	TableDisplayMini
)

type TableRow struct {
	Data     []string
	Children []TableRow
	Expanded bool
}

type TableColumn struct {
	Title      string
	Percentage int // width as percentage of total table width
}

type Table struct {
	render      *types.RenderState
	columns     []TableColumn
	rows        []TableRow
	displayMode TableDisplayMode

	selectedIndex int
	scrollOffset  int
	visibleRows   []VisibleRow

	OnSelect func(index int) error
	OnHover  func(index int)

	showBorder bool
}

type VisibleRow struct {
	data      []string
	isChild   bool
	parentIdx int
	childIdx  int
	rowIdx    int // index in original rows slice (-1 for children)
}

func (vr *VisibleRow) Data() []string {
	return vr.data
}

func NewTable(render *types.RenderState, columns []TableColumn, rows []TableRow, mode TableDisplayMode) *Table {
	t := &Table{
		render:      render,
		columns:     columns,
		rows:        rows,
		displayMode: mode,
		showBorder:  mode == TableDisplayMini,
	}
	t.buildVisibleRows()
	return t
}

func (t *Table) Init() tea.Cmd {
	return nil
}

func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *types.RenderState:
		t.render = msg
	case tea.KeyPressMsg:
		return t, t.handleKeyMsg(msg)
	}
	return t, nil
}

func (t *Table) handleKeyMsg(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case types.KeyUp, "k":
		t.moveCursor(-1)
	case types.KeyDown, "j":
		t.moveCursor(1)
	case types.KeyEnter:
		return t.selectRow()
	case "space", "tab":
		t.toggleExpansion()
		t.buildVisibleRows()
		t.ensureSelectedVisible()
	}
	return nil
}

func (t *Table) moveCursor(delta int) {
	next := t.selectedIndex + delta
	if next < 0 || next >= len(t.visibleRows) {
		return
	}
	t.selectedIndex = next
	t.ensureSelectedVisible()
	if t.OnHover != nil {
		t.OnHover(t.selectedIndex)
	}
}

func (t *Table) selectRow() tea.Cmd {
	if t.OnSelect == nil {
		return nil
	}
	return func() tea.Msg {
		err := t.OnSelect(t.selectedIndex)
		if err != nil {
			return err
		}
		return nil
	}
}

func (t *Table) View() tea.View {
	if t.render == nil || len(t.visibleRows) == 0 {
		return tea.View{Content: "No data"}
	}

	tableWidth := t.calculateTableWidth()
	colWidths := t.calculateColumnWidths(tableWidth)

	var content strings.Builder

	header := t.renderHeader(colWidths)
	content.WriteString(header)
	content.WriteString("\n")

	maxRows := t.maxVisibleRows()
	start := t.scrollOffset
	end := min(start+maxRows, len(t.visibleRows))

	if start > 0 {
		scrollHint := lipgloss.NewStyle().
			Foreground(t.render.Theme.ColorPalette().GrayColor()).
			Width(tableWidth).Align(lipgloss.Center).
			Render(fmt.Sprintf("↑ %d more", start))
		content.WriteString(scrollHint)
		content.WriteString("\n")
	}

	for i := start; i < end; i++ {
		rowStr := t.renderRow(t.visibleRows[i], colWidths, i == t.selectedIndex)
		content.WriteString(rowStr)
		content.WriteString("\n")
	}

	remaining := len(t.visibleRows) - end
	if remaining > 0 {
		scrollHint := lipgloss.NewStyle().
			Foreground(t.render.Theme.ColorPalette().GrayColor()).
			Width(tableWidth).Align(lipgloss.Center).
			Render(fmt.Sprintf("↓ %d more", remaining))
		content.WriteString(scrollHint)
		content.WriteString("\n")
	}

	result := content.String()

	if t.displayMode == TableDisplayMini && t.showBorder {
		result = t.renderMiniTable(result, tableWidth)
	}

	// Pad to fill available height so the table occupies the full content area.
	rendered := lipgloss.NewStyle().
		Width(t.render.ContentWidth).
		Height(t.render.ContentHeight).
		Render(result)
	return tea.View{Content: rendered}
}

func (t *Table) HelpMsg() string {
	return "↑/↓: navigate • enter: select • space/tab: expand/collapse"
}

func (t *Table) ShowFooter() bool {
	return true
}

func (t *Table) Type() string {
	return TableViewType
}

func (t *Table) SetOnSelect(callback func(index int) error) {
	t.OnSelect = callback
}

func (t *Table) SetOnHover(callback func(index int)) {
	t.OnHover = callback
}

func (t *Table) SetRows(rows []TableRow) {
	t.rows = rows
	t.selectedIndex = 0
	t.buildVisibleRows()
}

func (t *Table) GetSelectedRow() *VisibleRow {
	if t.selectedIndex >= 0 && t.selectedIndex < len(t.visibleRows) {
		return &t.visibleRows[t.selectedIndex]
	}
	return nil
}

func (t *Table) calculateTableWidth() int {
	if t.displayMode == TableDisplayMini {
		maxWidth := int(float64(t.render.ContentWidth) * 0.66)
		minWidth := 30
		if maxWidth < minWidth {
			return minWidth
		}
		return maxWidth
	}
	return t.render.ContentWidth
}

func (t *Table) calculateColumnWidths(totalWidth int) []int {
	widths := make([]int, len(t.columns))
	usedWidth := 0

	for i, col := range t.columns {
		if i == len(t.columns)-1 {
			// last column gets remaining width
			widths[i] = totalWidth - usedWidth
		} else {
			width := (totalWidth * col.Percentage) / 100
			widths[i] = width
			usedWidth += width
		}
	}

	return widths
}

func (t *Table) renderHeader(colWidths []int) string {
	var header string

	style := lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.NormalBorder(), false).
		BorderBottom(true).
		BorderBottomForeground(t.render.Theme.ColorPalette().BorderColor()).
		Foreground(t.render.Theme.ColorPalette().PrimaryColor())

	for i, col := range t.columns {
		title := col.Title
		if len(title) > colWidths[i]-1 {
			title = title[:colWidths[i]-4] + "..."
		}

		cellContent := style.Width(colWidths[i] - 1).Render(title)
		header = lipgloss.JoinHorizontal(lipgloss.Right, header, cellContent)
	}

	return header
}

func (t *Table) rowStyle(row VisibleRow, selected bool) lipgloss.Style {
	cp := t.render.Theme.ColorPalette()
	switch {
	case selected:
		return lipgloss.NewStyle().
			Background(cp.PrimaryColor()).
			Foreground(cp.GrayColor()).Bold(true)
	case row.isChild:
		return lipgloss.NewStyle().Foreground(cp.TertiaryColor())
	default:
		return lipgloss.NewStyle().Foreground(cp.BodyColor())
	}
}

func (t *Table) cellPrefix(row VisibleRow, colIdx int, selected bool) string {
	if colIdx != 0 {
		return ""
	}
	if row.isChild {
		if selected {
			return "  > "
		}
		return "    "
	}
	if row.rowIdx < 0 {
		return ""
	}
	children := t.rows[row.rowIdx].Children
	switch {
	case len(children) > 0 && t.rows[row.rowIdx].Expanded:
		return "◉ "
	case len(children) > 0:
		return "● "
	default:
		return "◌ "
	}
}

func (t *Table) renderRow(row VisibleRow, colWidths []int, selected bool) string {
	var rowStr strings.Builder
	style := t.rowStyle(row, selected)

	for i, cellData := range row.data {
		if i >= len(colWidths) {
			break
		}

		content := t.cellPrefix(row, i, selected) + cellData
		maxLen := colWidths[i] - 1
		if len(content) > maxLen && maxLen > 3 {
			content = content[:maxLen-3] + "..."
		} else if len(content) > maxLen {
			content = content[:maxLen]
		}

		cellContent := style.Width(colWidths[i] - 1).Render(content)
		rowStr.WriteString(cellContent)
	}

	return rowStr.String()
}

func (t *Table) renderMiniTable(content string, tableWidth int) string {
	leftPadding := max((t.render.ContentWidth-tableWidth)/2, 0)
	topMargin := 1
	// Border takes 2 lines (top+bottom), padding takes 2 lines (top+bottom)
	boxHeight := max(t.render.ContentHeight-topMargin-2-2, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.render.Theme.ColorPalette().BorderColor()).
		Padding(1).
		MarginLeft(leftPadding).
		MarginTop(topMargin).
		Height(boxHeight)

	return borderStyle.Render(content)
}

func (t *Table) maxVisibleRows() int {
	if t.render == nil || t.render.ContentHeight <= 0 {
		return len(t.visibleRows)
	}
	// Reserve lines for: header (2 lines: title + border), scroll hints (up to 2 lines)
	available := t.render.ContentHeight - 2
	if t.displayMode == TableDisplayMini {
		// Mini mode border + padding takes extra space
		available -= 4
	}
	if available < 1 {
		available = 1
	}
	if available >= len(t.visibleRows) {
		return len(t.visibleRows)
	}
	return available
}

func (t *Table) ensureSelectedVisible() {
	maxRows := t.maxVisibleRows()
	if t.selectedIndex < t.scrollOffset {
		t.scrollOffset = t.selectedIndex
	} else if t.selectedIndex >= t.scrollOffset+maxRows {
		t.scrollOffset = t.selectedIndex - maxRows + 1
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
}

func (t *Table) buildVisibleRows() {
	t.visibleRows = make([]VisibleRow, 0)

	for i, row := range t.rows {
		t.visibleRows = append(t.visibleRows, VisibleRow{
			data:      row.Data,
			isChild:   false,
			parentIdx: -1,
			childIdx:  -1,
			rowIdx:    i,
		})

		if row.Expanded {
			for j, child := range row.Children {
				t.visibleRows = append(t.visibleRows, VisibleRow{
					data:      child.Data,
					isChild:   true,
					parentIdx: i,
					childIdx:  j,
					rowIdx:    -1,
				})
			}
		}
	}

	if t.selectedIndex >= len(t.visibleRows) {
		t.selectedIndex = len(t.visibleRows) - 1
	}
	if t.selectedIndex < 0 {
		t.selectedIndex = 0
	}
}

func (t *Table) toggleExpansion() {
	if t.selectedIndex < 0 || t.selectedIndex >= len(t.visibleRows) {
		return
	}

	selectedRow := t.visibleRows[t.selectedIndex]
	if selectedRow.isChild || selectedRow.rowIdx < 0 {
		return
	}

	rowIdx := selectedRow.rowIdx
	if rowIdx >= len(t.rows) {
		return
	}
	if len(t.rows[rowIdx].Children) == 0 {
		return
	}

	for i := range t.rows {
		if i != rowIdx {
			t.rows[i].Expanded = false
		}
	}
	t.rows[rowIdx].Expanded = !t.rows[rowIdx].Expanded
}
