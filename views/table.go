package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
		return t, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.selectedIndex > 0 {
				t.selectedIndex--
				if t.OnHover != nil {
					t.OnHover(t.selectedIndex)
				}
			}
		case "down", "j":
			if t.selectedIndex < len(t.visibleRows)-1 {
				t.selectedIndex++
				if t.OnHover != nil {
					t.OnHover(t.selectedIndex)
				}
			}
		case "enter":
			if t.OnSelect != nil {
				return t, func() tea.Msg {
					err := t.OnSelect(t.selectedIndex)
					if err != nil {
						return err
					}
					return nil
				}
			}
		case " ", "tab":
			t.toggleExpansion()
			t.buildVisibleRows()
		}
	}
	return t, nil
}

func (t *Table) View() string {
	if t.render == nil || len(t.visibleRows) == 0 {
		return "No data"
	}

	tableWidth := t.calculateTableWidth()
	colWidths := t.calculateColumnWidths(tableWidth)

	var content strings.Builder

	header := t.renderHeader(colWidths)
	content.WriteString(header)
	content.WriteString("\n")

	for i, row := range t.visibleRows {
		rowStr := t.renderRow(row, colWidths, i == t.selectedIndex)
		content.WriteString(rowStr)
		content.WriteString("\n")
	}

	result := content.String()
	if t.displayMode == TableDisplayMini && t.showBorder {
		return t.renderMiniTable(result, tableWidth)
	}

	return result
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

func (t *Table) renderRow(row VisibleRow, colWidths []int, selected bool) string {
	var rowStr strings.Builder

	var style lipgloss.Style
	if selected {
		style = lipgloss.NewStyle().
			Background(t.render.Theme.ColorPalette().PrimaryColor()).
			Foreground(t.render.Theme.ColorPalette().GrayColor()).Bold(true)
	} else if row.isChild {
		style = lipgloss.NewStyle().
			Foreground(t.render.Theme.ColorPalette().TertiaryColor())
	} else {
		style = lipgloss.NewStyle().
			Foreground(t.render.Theme.ColorPalette().BodyColor())
	}

	for i, cellData := range row.data {
		if i >= len(colWidths) {
			break
		}

		content := cellData
		if i == 0 && !row.isChild && row.rowIdx >= 0 {
			if len(t.rows[row.rowIdx].Children) > 0 {
				if t.rows[row.rowIdx].Expanded {
					content = "◉ " + content
				} else {
					content = "● " + content
				}
			} else {
				content = "◌ " + content
			}
		}

		if i == 0 && row.isChild {
			if selected {
				content = "  > " + content
			} else {
				content = "    " + content
			}
		}

		maxLen := colWidths[i] - 1
		if len(content) > maxLen {
			if maxLen > 3 {
				content = content[:maxLen-3] + "..."
			} else {
				content = content[:maxLen]
			}
		}

		cellContent := style.Width(colWidths[i] - 1).Render(content)
		rowStr.WriteString(cellContent)
	}

	return rowStr.String()
}

func (t *Table) renderMiniTable(content string, tableWidth int) string {
	leftPadding := (t.render.ContentWidth - tableWidth) / 2
	if leftPadding < 0 {
		leftPadding = 0
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.render.Theme.ColorPalette().BorderColor()).
		Padding(1).
		MarginLeft(leftPadding)

	return borderStyle.Render(content)
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
