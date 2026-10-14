package views_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	tuikitIO "github.com/flowexec/tuikit/io"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func testRenderState() *types.RenderState {
	return &types.RenderState{
		Width:         80,
		Height:        40,
		ContentWidth:  80,
		ContentHeight: 38,
		Theme:         themes.EverforestTheme(),
	}
}

func lineOf(t *testing.T, content, text string) int {
	t.Helper()
	for i, line := range strings.Split(ansi.Strip(content), "\n") {
		if strings.Contains(line, text) {
			return i
		}
	}
	t.Fatalf("%q not found in:\n%s", text, ansi.Strip(content))
	return -1
}

func leftClick(x, y int) tea.MouseClickMsg {
	return tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft}
}

func numberedRows(n int) []views.TableRow {
	rows := make([]views.TableRow, n)
	for i := range rows {
		rows[i] = views.TableRow{Data: []string{fmt.Sprintf("row-%02d", i)}}
	}
	return rows
}

// lineOf returns the index of the first rendered line containing text.
// runCmd executes cmd and returns the messages it produces, flattening batches.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var msgs []tea.Msg
		for _, c := range batch {
			msgs = append(msgs, runCmd(c)...)
		}
		return msgs
	}
	return []tea.Msg{msg}
}

func isEnter(msg tea.Msg) bool {
	key, ok := msg.(tea.KeyPressMsg)
	return ok && key.String() == types.KeyEnter
}

func TestTableMouseWheel(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"First"}},
		{Data: []string{"Second"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	table.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if selected := table.GetSelectedRow(); selected == nil || selected.Data()[0] != "Second" {
		t.Errorf("expected wheel down to select second row, got %v", selected)
	}

	table.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if selected := table.GetSelectedRow(); selected == nil || selected.Data()[0] != "First" {
		t.Errorf("expected wheel up to select first row, got %v", selected)
	}
}

func TestCollectionViewMouseWheel(t *testing.T) {
	state := testRenderState()
	collection := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red", SubHeader: "Primary Color"},
		&types.EntityInfo{ID: "green", Header: "Green", SubHeader: "Secondary Color"},
	)
	var selected string
	view := views.NewCollectionView(state, collection, types.CollectionFormatList, func(id string) error {
		selected = id
		return nil
	})
	view.View()

	// Items are sorted, so "green" is first.
	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	view.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if selected != "red" {
		t.Errorf("expected wheel down to select 'red', got %q", selected)
	}

	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	view.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if selected != "green" {
		t.Errorf("expected wheel up to select 'green', got %q", selected)
	}
}

func TestMarkdownViewMouseWheel(t *testing.T) {
	state := testRenderState()
	state.ContentHeight = 5
	var md strings.Builder
	for i := range 50 {
		fmt.Fprintf(&md, "line-%d\n\n", i)
	}
	view := views.NewMarkdownView(state, md.String())
	before := view.View().Content
	if !strings.Contains(before, "line-0") {
		t.Fatalf("expected first line visible initially, got %q", before)
	}

	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if after := view.View().Content; strings.Contains(after, "line-0") {
		t.Errorf("expected wheel down to scroll past first line, got %q", after)
	}
}

func TestTableMouseClick(t *testing.T) {
	for _, mode := range []views.TableDisplayMode{views.TableDisplayFull, views.TableDisplayMini} {
		t.Run(fmt.Sprintf("mode-%d", mode), func(t *testing.T) {
			state := testRenderState()
			columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
			table := views.NewTable(state, columns, numberedRows(3), mode)
			selected := -1
			table.SetOnSelect(func(i int) error { selected = i; return nil })

			y := lineOf(t, table.View().Content, "row-02")
			if _, cmd := table.Update(leftClick(5, y)); cmd != nil {
				t.Error("expected first click to only select the row")
			}
			if got := table.GetSelectedRow().Data()[0]; got != "row-02" {
				t.Fatalf("expected click to select row-02, got %s", got)
			}

			_, cmd := table.Update(leftClick(5, y))
			msgs := runCmd(cmd)
			if len(msgs) != 1 || !isEnter(msgs[0]) {
				t.Fatalf("expected second click to press enter, got %v", msgs)
			}
			_, cmd = table.Update(msgs[0])
			runCmd(cmd)
			if selected != 2 {
				t.Errorf("expected OnSelect(2), got %d", selected)
			}
		})
	}
}

func TestTableMouseClickScrolled(t *testing.T) {
	state := testRenderState()
	state.ContentHeight = 10
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	table := views.NewTable(state, columns, numberedRows(40), views.TableDisplayFull)
	for range 20 {
		table.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	content := table.View().Content
	if !strings.Contains(content, "more") {
		t.Fatalf("expected table to be scrolled, got:\n%s", content)
	}

	table.Update(leftClick(5, lineOf(t, content, "row-17")))
	if got := table.GetSelectedRow().Data()[0]; got != "row-17" {
		t.Errorf("expected click to select row-17, got %s", got)
	}

	// Clicking the header or a scroll hint does nothing.
	table.Update(leftClick(5, 0))
	table.Update(leftClick(5, lineOf(t, content, "more")))
	if got := table.GetSelectedRow().Data()[0]; got != "row-17" {
		t.Errorf("expected selection to stay on row-17, got %s", got)
	}
}

func TestCollectionViewMouseClick(t *testing.T) {
	state := testRenderState()
	collection := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red"},
		&types.EntityInfo{ID: "green", Header: "Green"},
		&types.EntityInfo{ID: "blue", Header: "Blue"},
	)
	var selected string
	view := views.NewCollectionView(state, collection, types.CollectionFormatList, func(id string) error {
		selected = id
		return nil
	})

	// Items are sorted: blue, green, red.
	y := lineOf(t, view.View().Content, "Red")
	if _, cmd := view.Update(leftClick(5, y)); cmd != nil {
		t.Error("expected first click to only select the item")
	}
	_, cmd := view.Update(leftClick(5, y))
	for _, msg := range runCmd(cmd) {
		view.Update(msg)
	}
	if selected != "red" {
		t.Errorf("expected clicks to select and open 'red', got %q", selected)
	}
}

func TestLogArchiveViewMouseClick(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"first", "second"} {
		path := filepath.Join(dir, tuikitIO.NewArchiveFileName(id))
		if err := os.WriteFile(path, []byte("log for "+id), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	view := views.NewLogArchiveView(testRenderState(), dir, false)
	y := lineOf(t, view.View().Content, "second")
	view.Update(leftClick(5, y))
	_, cmd := view.Update(leftClick(5, y))
	for _, msg := range runCmd(cmd) {
		view.Update(msg)
	}
	if content := view.View().Content; !strings.Contains(content, "log for second") {
		t.Errorf("expected clicks to open 'second' entry, got:\n%s", content)
	}
}

func TestLibraryMouse(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Name", Percentage: 100}}
	lib := views.NewLibrary(state,
		views.LibraryPage{
			Title: "Workspaces",
			Factory: func(r *types.RenderState, _ []views.PageSelection) (tea.Model, []types.KeyCallback) {
				return views.NewTable(r, columns, []views.TableRow{
					{Data: []string{"alpha"}}, {Data: []string{"beta"}},
				}, views.TableDisplayFull), nil
			},
		},
		views.LibraryPage{
			Title: "Executables",
			Factory: func(r *types.RenderState, _ []views.PageSelection) (tea.Model, []types.KeyCallback) {
				return views.NewTable(r, columns, []views.TableRow{{Data: []string{"build"}}}, views.TableDisplayFull), nil
			},
		},
	)

	// Clicking a row twice drills down into it.
	y := lineOf(t, lib.View().Content, "beta")
	lib.Update(leftClick(5, y))
	_, cmd := lib.Update(leftClick(5, y))
	for _, msg := range runCmd(cmd) {
		lib.Update(msg)
	}
	content := lib.View().Content
	if !strings.Contains(content, "Workspaces: beta") || !strings.Contains(content, "build") {
		t.Fatalf("expected to drill into 'beta', got:\n%s", ansi.Strip(content))
	}

	// Clicking the first breadcrumb navigates back.
	lib.Update(leftClick(len("  Work"), 0))
	if content = lib.View().Content; !strings.Contains(content, "alpha") {
		t.Errorf("expected breadcrumb click to return to first page, got:\n%s", ansi.Strip(content))
	}
}

func TestCopyKey(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	table := views.NewTable(state, columns, numberedRows(3), views.TableDisplayFull)
	table.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	_, cmd := table.Update(tea.KeyPressMsg{Text: "c"})
	var toast *types.ToastMsg
	for _, msg := range runCmd(cmd) {
		if m, ok := msg.(types.ToastMsg); ok {
			toast = &m
		}
	}
	if toast == nil || toast.Text != "Copied row-01" {
		t.Errorf("expected copy toast for row-01, got %v", toast)
	}

	table.CopyFunc = func(row []string) string { return "custom:" + row[0] }
	_, cmd = table.Update(tea.KeyPressMsg{Text: "c"})
	toast = nil
	for _, msg := range runCmd(cmd) {
		if m, ok := msg.(types.ToastMsg); ok {
			toast = &m
		}
	}
	if toast == nil || toast.Text != "Copied custom:row-01" {
		t.Errorf("expected CopyFunc to be used, got %v", toast)
	}
}

func TestCopyKeyCallbackOverride(t *testing.T) {
	state := testRenderState()
	called := false
	view := views.NewDetailView(state, "body")
	view.SetKeyCallbacks([]types.KeyCallback{{Key: "c", Label: "custom", Callback: func() error {
		called = true
		return nil
	}}})
	view.Update(tea.KeyPressMsg{Text: "c"})
	if !called {
		t.Error("expected a 'c' key callback to take precedence over copy")
	}
	for _, k := range view.HelpBindings() {
		if k.Desc == "copy" {
			t.Error("expected copy help binding to be hidden when overridden")
		}
	}
}

// copyToast returns the text of the toast produced by a copy command.
func copyToast(t *testing.T, cmd tea.Cmd) string {
	t.Helper()
	for _, msg := range runCmd(cmd) {
		if m, ok := msg.(types.ToastMsg); ok {
			return m.Text
		}
	}
	t.Fatal("expected a copy toast")
	return ""
}

// clipboardText returns the text a copy command writes to the clipboard.
func clipboardText(cmd tea.Cmd) string {
	for _, msg := range runCmd(cmd) {
		if _, ok := msg.(types.ToastMsg); !ok {
			return fmt.Sprint(msg)
		}
	}
	return ""
}

func pressKey(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: text, Code: []rune(text)[0]}
}

func TestEntityViewCopy(t *testing.T) {
	thing := &sampleTypes.Thing{Name: "widget", Type: "gadget"}
	view := views.NewEntityView(testRenderState(), thing, types.EntityFormatDocument)
	_, cmd := view.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied # widget I am a(n) **gadget.**" {
		t.Errorf("unexpected document copy: %q", got)
	}

	view.Update(pressKey("y"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.Contains(got, "name: widget") {
		t.Errorf("expected YAML copy, got %q", got)
	}

	view.Update(pressKey("j"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.Contains(got, `"data":"widget"`) {
		t.Errorf("expected JSON copy, got %q", got)
	}
	if !hasCopyHelpKey(view.HelpBindings()) {
		t.Error("expected copy help binding")
	}
}

func TestDetailContentViewCopy(t *testing.T) {
	thing := &sampleTypes.Thing{Name: "widget", Type: "gadget"}
	view := views.NewDetailContentView(testRenderState(), views.DetailContentOpts{
		Title:  "ws/ns:widget",
		Entity: thing,
	})
	_, cmd := view.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied ws/ns:widget" {
		t.Errorf("expected title copy, got %q", got)
	}

	view.Update(pressKey("y"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.Contains(got, "name: widget") {
		t.Errorf("expected YAML copy, got %q", got)
	}

	view.Update(pressKey("j"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.Contains(got, `"data":"widget"`) {
		t.Errorf("expected JSON copy, got %q", got)
	}

	custom := views.NewDetailContentView(testRenderState(), views.DetailContentOpts{
		Title:     "Widget",
		CopyValue: "flow exec widget",
	})
	_, cmd = custom.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied flow exec widget" {
		t.Errorf("expected CopyValue copy, got %q", got)
	}
	if !hasCopyHelpKey(custom.HelpBindings()) {
		t.Error("expected copy help binding")
	}
}

func TestDetailAndMarkdownViewCopy(t *testing.T) {
	detail := views.NewDetailView(testRenderState(), "detail body")
	_, cmd := detail.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied detail body" {
		t.Errorf("unexpected detail copy: %q", got)
	}

	md := views.NewMarkdownView(testRenderState(), "# Title")
	_, cmd = md.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied # Title" {
		t.Errorf("unexpected markdown copy: %q", got)
	}
	if !hasCopyHelpKey(md.HelpBindings()) {
		t.Error("expected copy help binding")
	}
}

func TestCollectionViewCopy(t *testing.T) {
	collection := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red"},
		&types.EntityInfo{ID: "green", Header: "Green"},
	)
	view := views.NewCollectionView(testRenderState(), collection, types.CollectionFormatList, nil)
	_, cmd := view.Update(pressKey("c"))
	if got := copyToast(t, cmd); got != "Copied green" {
		t.Errorf("expected selected ID copy, got %q", got)
	}

	view.Update(pressKey("y"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.Contains(got, "header: Red") {
		t.Errorf("expected YAML copy, got %q", got)
	}

	view.Update(pressKey("j"))
	_, cmd = view.Update(pressKey("c"))
	if got := copyToast(t, cmd); !strings.HasPrefix(got, "Copied [") && !strings.HasPrefix(got, "Copied {") {
		t.Errorf("expected JSON copy, got %q", got)
	}
	if !hasCopyHelpKey(view.HelpBindings()) {
		t.Error("expected copy help binding")
	}
}

func TestLogArchiveViewCopyAndWheel(t *testing.T) {
	dir := t.TempDir()
	for _, id := range []string{"first", "second"} {
		path := filepath.Join(dir, tuikitIO.NewArchiveFileName(id))
		if err := os.WriteFile(path, []byte("log for "+id), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	view := views.NewLogArchiveView(testRenderState(), dir, false)
	view.View()
	_, cmd := view.Update(pressKey("c"))
	first := clipboardText(cmd)

	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	_, cmd = view.Update(pressKey("c"))
	second := clipboardText(cmd)
	if first == second || filepath.Dir(first) != dir || filepath.Dir(second) != dir {
		t.Errorf("expected wheel to change the copied path, got %q and %q", first, second)
	}

	// Copying the open entry uses its path; mouse events are ignored while it is open.
	view.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	_, cmd = view.Update(pressKey("c"))
	if got := clipboardText(cmd); got != second {
		t.Errorf("expected open entry path %q, got %q", second, got)
	}
}

func TestLibraryMouseWheelAndIgnoredClicks(t *testing.T) {
	columns := []views.TableColumn{{Title: "Name", Percentage: 100}}
	var table *views.Table
	lib := views.NewLibrary(testRenderState(), views.LibraryPage{
		Title: "Workspaces",
		Factory: func(r *types.RenderState, _ []views.PageSelection) (tea.Model, []types.KeyCallback) {
			table = views.NewTable(r, columns, numberedRows(3), views.TableDisplayFull)
			return table, nil
		},
	})
	lib.Update(tea.MouseWheelMsg{Y: 5, Button: tea.MouseWheelDown})
	if got := table.GetSelectedRow().Data()[0]; got != "row-01" {
		t.Errorf("expected wheel to reach the table, got %s", got)
	}

	// Clicks on the breadcrumb margin, the current page and non-left buttons do nothing.
	for _, msg := range []tea.Msg{
		leftClick(0, 0),
		leftClick(3, 0),
		tea.MouseClickMsg{X: 3, Y: 0, Button: tea.MouseRight},
		leftClick(3, 1),
	} {
		if _, cmd := lib.Update(msg); cmd != nil {
			t.Errorf("expected %v to be ignored", msg)
		}
	}
}

func hasCopyHelpKey(keys []themes.HelpKey) bool {
	for _, k := range keys {
		if k.Key == types.KeyCopy {
			return true
		}
	}
	return false
}

func TestCollectionViewStructuredFormatsScroll(t *testing.T) {
	state := testRenderState()
	state.ContentHeight = 6
	items := make([]*types.EntityInfo, 30)
	for i := range items {
		items[i] = &types.EntityInfo{ID: fmt.Sprintf("item-%02d", i), Header: fmt.Sprintf("Item %02d", i)}
	}
	view := views.NewCollectionView(state, sampleTypes.NewThingList("Thing", items...), types.CollectionFormatYAML, nil)
	if !strings.Contains(view.View().Content, "item-00") {
		t.Fatal("expected first item visible initially")
	}

	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	view.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if strings.Contains(view.View().Content, "item-00") {
		t.Error("expected wheel to scroll YAML past the first item")
	}

	view.Update(pressKey("j"))
	if !strings.Contains(view.View().Content, "item-00") {
		t.Error("expected switching formats to reset the scroll position")
	}
	for range 8 {
		view.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if strings.Contains(view.View().Content, "item-00") {
		t.Error("expected down arrow to scroll JSON past the first item")
	}
	if lines := strings.Count(view.View().Content, "\n") + 1; lines > state.ContentHeight {
		t.Errorf("expected output to fit %d lines, got %d", state.ContentHeight, lines)
	}
}
