package tuikit_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/flowexec/tuikit"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func TestMain(m *testing.M) {
	os.Setenv("NO_COLOR", "1")
	os.Setenv("TERM", "dumb")
	lipgloss.Writer.Profile = colorprofile.Ascii
	os.Exit(m.Run())
}

// testRenderState returns a RenderState suitable for unit-testing views
// outside of a full bubbletea program.
func testRenderState() *types.RenderState {
	return &types.RenderState{
		Width:         80,
		Height:        40,
		ContentWidth:  80,
		ContentHeight: 38,
		Theme:         themes.EverforestTheme(),
	}
}

// --- View unit tests ---
// These test View().Content directly, avoiding bubbletea's renderer
// and the non-deterministic terminal escape sequences it produces.

func TestFrameView(t *testing.T) {
	inner := &sampleTypes.Echo{Content: "Hello, world!"}
	view := views.NewFrameView(inner)
	content := view.View().Content
	if content != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", content)
	}
}

func TestLoadingView(t *testing.T) {
	state := testRenderState()
	view := views.NewLoadingView("thinking...", state.Theme)
	content := view.View().Content
	if !strings.Contains(content, "thinking...") {
		t.Errorf("expected loading message, got %q", content)
	}
}

func TestErrorView(t *testing.T) {
	state := testRenderState()
	view := views.NewErrorView(errors.New("something went wrong"), state.Theme)
	content := view.View().Content
	if !strings.Contains(content, "something went wrong") {
		t.Errorf("expected error message, got %q", content)
	}
	if !strings.Contains(content, "encountered error") {
		t.Errorf("expected error header, got %q", content)
	}
}

func TestMarkdownView(t *testing.T) {
	state := testRenderState()
	md := "# Hello!\n\nI am a **Markdown** document."
	view := views.NewMarkdownView(state, md)
	content := view.View().Content
	if !strings.Contains(content, "Hello") {
		t.Errorf("expected heading, got %q", content)
	}
	if !strings.Contains(content, "Markdown") {
		t.Errorf("expected bold text, got %q", content)
	}
}

func TestEntityView(t *testing.T) {
	state := testRenderState()
	e := &sampleTypes.Thing{Name: "Green", Type: "Color"}
	view := views.NewEntityView(state, e, types.EntityFormatDocument)
	content := view.View().Content
	if !strings.Contains(content, "Green") {
		t.Errorf("expected entity name, got %q", content)
	}
	if !strings.Contains(content, "Color") {
		t.Errorf("expected entity type, got %q", content)
	}
}

func TestCollectionView(t *testing.T) {
	state := testRenderState()
	collection := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red", SubHeader: "Primary Color"},
		&types.EntityInfo{ID: "green", Header: "Green", SubHeader: "Secondary Color"},
	)
	view := views.NewCollectionView(state, collection, types.CollectionFormatList, nil)
	content := view.View().Content
	if !strings.Contains(content, "Red") {
		t.Errorf("expected 'Red' in collection, got %q", content)
	}
	if !strings.Contains(content, "Green") {
		t.Errorf("expected 'Green' in collection, got %q", content)
	}
}

func TestDetailView(t *testing.T) {
	state := testRenderState()
	view := views.NewDetailView(
		state,
		"Body content here.",
		views.DetailField{Key: "Name", Value: "Test Item"},
		views.DetailField{Key: "Status", Value: "Active"},
	)
	content := view.View().Content
	if !strings.Contains(content, "Name") {
		t.Errorf("expected metadata key, got %q", content)
	}
	if !strings.Contains(content, "Test Item") {
		t.Errorf("expected metadata value, got %q", content)
	}
	if !strings.Contains(content, "Body content here.") {
		t.Errorf("expected body content, got %q", content)
	}
}

func TestTableView(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{
		{Title: "Name", Percentage: 50},
		{Title: "Status", Percentage: 50},
	}
	rows := []views.TableRow{
		{Data: []string{"Alpha", "Active"}},
		{Data: []string{"Beta", "Inactive"}, Children: []views.TableRow{
			{Data: []string{"Beta-1", "Running"}},
		}},
	}
	view := views.NewTable(state, columns, rows, views.TableDisplayFull)
	content := view.View().Content
	if !strings.Contains(content, "Name") {
		t.Errorf("expected column header, got %q", content)
	}
	if !strings.Contains(content, "Alpha") {
		t.Errorf("expected row data, got %q", content)
	}
	if !strings.Contains(content, "Beta") {
		t.Errorf("expected row data, got %q", content)
	}
}

func TestTableMiniView(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{
		{Title: "Command", Percentage: 100},
	}
	rows := []views.TableRow{
		{Data: []string{"build"}},
		{Data: []string{"test"}},
		{Data: []string{"deploy"}},
	}
	view := views.NewTable(state, columns, rows, views.TableDisplayMini)
	content := view.View().Content
	if !strings.Contains(content, "build") {
		t.Errorf("expected 'build' row, got %q", content)
	}
	if !strings.Contains(content, "deploy") {
		t.Errorf("expected 'deploy' row, got %q", content)
	}
}

func TestTableKeyNavigation(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"First"}},
		{Data: []string{"Second"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	selected := table.GetSelectedRow()
	if selected == nil || selected.Data()[0] != "First" {
		t.Fatal("expected initial selection on first row")
	}

	table.Update(tea.KeyPressMsg{Text: "j"})
	selected = table.GetSelectedRow()
	if selected == nil || selected.Data()[0] != "Second" {
		t.Errorf("expected selection to move down, got %v", selected)
	}

	table.Update(tea.KeyPressMsg{Text: "k"})
	selected = table.GetSelectedRow()
	if selected == nil || selected.Data()[0] != "First" {
		t.Errorf("expected selection to move up, got %v", selected)
	}
}

func TestTableExpansion(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"Parent"}, Children: []views.TableRow{
			{Data: []string{"Child"}},
		}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	// Children should not be visible initially
	content := table.View().Content
	if strings.Contains(content, "Child") {
		t.Error("children should be collapsed initially")
	}

	// Toggle expansion with space
	table.Update(tea.KeyPressMsg{Code: tea.KeySpace})
	content = table.View().Content
	if !strings.Contains(content, "Child") {
		t.Error("children should be visible after expansion")
	}
}

// --- Table filter tests ---

func TestTableFilterBasic(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{
		{Title: "Name", Percentage: 50},
		{Title: "Status", Percentage: 50},
	}
	rows := []views.TableRow{
		{Data: []string{"Alpha", "Active"}},
		{Data: []string{"Beta", "Inactive"}},
		{Data: []string{"Gamma", "Active"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	// Activate filter with "/"
	table.Update(tea.KeyPressMsg{Text: "/"})

	// Type "alpha"
	for _, ch := range "alpha" {
		table.Update(tea.KeyPressMsg{Text: string(ch)})
	}

	content := table.View().Content
	if !strings.Contains(content, "Alpha") {
		t.Error("expected Alpha to be visible")
	}
	if strings.Contains(content, "Beta") {
		t.Error("expected Beta to be filtered out")
	}
	if strings.Contains(content, "Gamma") {
		t.Error("expected Gamma to be filtered out")
	}
}

func TestTableFilterChildren(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"Parent"}, Children: []views.TableRow{
			{Data: []string{"MatchChild"}},
			{Data: []string{"Other"}},
		}},
		{Data: []string{"Unrelated"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	// Activate and type filter
	table.Update(tea.KeyPressMsg{Text: "/"})
	for _, ch := range "match" {
		table.Update(tea.KeyPressMsg{Text: string(ch)})
	}

	content := table.View().Content
	if !strings.Contains(content, "Parent") {
		t.Error("expected Parent to be visible (has matching child)")
	}
	if !strings.Contains(content, "MatchChild") {
		t.Error("expected MatchChild to be visible")
	}
	if strings.Contains(content, "Other") {
		t.Error("expected Other child to be filtered out")
	}
	if strings.Contains(content, "Unrelated") {
		t.Error("expected Unrelated to be filtered out")
	}
}

func TestTableFilterNoMatch(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"Alpha"}},
		{Data: []string{"Beta"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	table.Update(tea.KeyPressMsg{Text: "/"})
	for _, ch := range "zzz" {
		table.Update(tea.KeyPressMsg{Text: string(ch)})
	}

	content := table.View().Content
	if !strings.Contains(content, "No matches") {
		t.Error("expected 'No matches' when filter has no results")
	}
	if !strings.Contains(content, "Filter:") {
		t.Error("expected filter bar to be visible even with no matches")
	}
}

func TestTableFilterEscCancel(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"Alpha"}},
		{Data: []string{"Beta"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	// Activate filter and type
	table.Update(tea.KeyPressMsg{Text: "/"})
	for _, ch := range "zzz" {
		table.Update(tea.KeyPressMsg{Text: string(ch)})
	}

	// Esc should cancel and restore all rows
	table.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	content := table.View().Content
	if !strings.Contains(content, "Alpha") {
		t.Error("expected Alpha to be visible after cancel")
	}
	if !strings.Contains(content, "Beta") {
		t.Error("expected Beta to be visible after cancel")
	}
}

func TestTableCapturingInput(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	rows := []views.TableRow{{Data: []string{"Alpha"}}}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)

	ic, ok := any(table).(tuikit.InputCapturer)
	if !ok {
		t.Fatal("Table should implement InputCapturer")
	}
	if ic.CapturingInput() {
		t.Error("should not be capturing input initially")
	}
	table.Update(tea.KeyPressMsg{Text: "/"})
	if !ic.CapturingInput() {
		t.Error("should be capturing input after activating filter")
	}
}

func TestTableFilterCustomFunc(t *testing.T) {
	state := testRenderState()
	columns := []views.TableColumn{{Title: "Name", Percentage: 100}}
	rows := []views.TableRow{
		{Data: []string{"Alice"}},
		{Data: []string{"Bob"}},
	}
	table := views.NewTable(state, columns, rows, views.TableDisplayFull)
	// Custom filter that only matches exact first character.
	table.FilterFunc = func(query string, row []string) bool {
		return len(row) > 0 && len(row[0]) > 0 && strings.EqualFold(string(row[0][0]), query)
	}

	table.Update(tea.KeyPressMsg{Text: "/"})
	table.Update(tea.KeyPressMsg{Text: "b"})

	content := table.View().Content
	if !strings.Contains(content, "Bob") {
		t.Error("expected Bob to match custom filter")
	}
	if strings.Contains(content, "Alice") {
		t.Error("expected Alice to be filtered out by custom filter")
	}
}

// --- HelpBindings tests ---

func TestHelpBindingsNilViews(t *testing.T) {
	state := testRenderState()
	nilViews := map[string]tuikit.View{
		"loading": views.NewLoadingView("test", state.Theme),
		"error":   views.NewErrorView(errors.New("err"), state.Theme),
		"frame":   views.NewFrameView(&sampleTypes.Echo{Content: "x"}),
	}
	for name, v := range nilViews {
		t.Run(name, func(t *testing.T) {
			if bindings := v.HelpBindings(); bindings != nil {
				t.Errorf("expected nil HelpBindings, got %v", bindings)
			}
		})
	}
}

func helpKeys(v tuikit.View) map[string]bool {
	m := make(map[string]bool)
	for _, b := range v.HelpBindings() {
		m[b.Key] = true
	}
	return m
}

func TestHelpBindingsTable(t *testing.T) {
	state := testRenderState()
	cols := []views.TableColumn{{Title: "X", Percentage: 100}}
	rows := []views.TableRow{{Data: []string{"a"}}}
	table := views.NewTable(state, cols, rows, views.TableDisplayFull)
	for _, key := range []string{"↑/↓/j/k", "enter", "space/tab", "/"} {
		if !helpKeys(table)[key] {
			t.Errorf("missing key %q in table help bindings", key)
		}
	}
}

func TestHelpBindingsDetail(t *testing.T) {
	state := testRenderState()
	view := views.NewDetailView(state, "body")
	for _, key := range []string{"j/k", "u/d", "g/G"} {
		if !helpKeys(view)[key] {
			t.Errorf("missing key %q in detail help bindings", key)
		}
	}
}

func TestHelpBindingsMarkdown(t *testing.T) {
	state := testRenderState()
	view := views.NewMarkdownView(state, "# hi")
	if !helpKeys(view)["↑/↓"] {
		t.Error("missing key ↑/↓ in markdown help bindings")
	}
}

// --- Theme render tests ---

func TestRenderHeader(t *testing.T) {
	theme := themes.EverforestTheme()
	header := theme.RenderHeader("MyApp", "v1.0", "Env", "prod", 80)
	if header == "" {
		t.Fatal("expected non-empty header")
	}
	if !strings.Contains(header, "MyApp") {
		t.Error("expected app name in header")
	}
	if !strings.Contains(header, "help") {
		t.Error("expected help hint in header")
	}
	if !strings.Contains(header, "v1.0") {
		t.Error("expected version in header")
	}
}

func TestRenderHeaderNoVersion(t *testing.T) {
	theme := themes.EverforestTheme()
	header := theme.RenderHeader("MyApp", "", "", "", 80)
	if !strings.Contains(header, "MyApp") {
		t.Error("expected app name in header")
	}
	if strings.Contains(header, "v1") {
		t.Error("expected no version in header")
	}
}

func TestRenderHelpPopup(t *testing.T) {
	theme := themes.EverforestTheme()
	keys := []themes.HelpKey{
		{Key: "q", Desc: "quit"},
		{Key: "enter", Desc: "select"},
	}
	out := theme.RenderHelpPopup(keys, 80, 40)
	if out == "" {
		t.Fatal("expected non-empty popup")
	}
}

func TestRenderHelpPopupEmpty(t *testing.T) {
	theme := themes.EverforestTheme()
	out := theme.RenderHelpPopup(nil, 80, 40)
	if out != "" {
		t.Fatalf("expected empty string for nil keys, got %q", out)
	}
}

func TestRenderToast(t *testing.T) {
	theme := themes.EverforestTheme()
	levels := []themes.OutputLevel{
		themes.OutputLevelSuccess,
		themes.OutputLevelWarning,
		themes.OutputLevelError,
		themes.OutputLevelInfo,
		themes.OutputLevelNotice,
	}
	for _, lvl := range levels {
		out := theme.RenderToast("test message", lvl, 80)
		if out == "" {
			t.Errorf("expected non-empty toast for level %s", lvl)
		}
	}
}

// --- Library view tests ---

func testLibrary() *views.Library {
	state := testRenderState()
	page0 := views.LibraryPage{
		Title: "Categories",
		Factory: func(render *types.RenderState, _ []views.PageSelection) (tea.Model, []types.KeyCallback) {
			cols := []views.TableColumn{{Title: "Name", Percentage: 50}, {Title: "Count", Percentage: 50}}
			rows := []views.TableRow{
				{Data: []string{"Alpha", "3"}},
				{Data: []string{"Beta", "5"}},
			}
			return views.NewTable(render, cols, rows, views.TableDisplayFull), nil
		},
	}
	page1 := views.LibraryPage{
		Title: "Items",
		Factory: func(render *types.RenderState, selections []views.PageSelection) (tea.Model, []types.KeyCallback) {
			cat := selections[0].Data[0]
			cols := []views.TableColumn{{Title: "Item", Percentage: 100}}
			rows := []views.TableRow{
				{Data: []string{cat + "-item-1"}},
				{Data: []string{cat + "-item-2"}},
			}
			keys := []types.KeyCallback{
				{Key: "x", Label: "action", Callback: func() error { return nil }},
			}
			return views.NewTable(render, cols, rows, views.TableDisplayFull), keys
		},
	}
	page2 := views.LibraryPage{
		Title: "Details",
		Factory: func(render *types.RenderState, selections []views.PageSelection) (tea.Model, []types.KeyCallback) {
			item := selections[1].Data[0]
			return views.NewDetailView(render, "Detail for "+item), nil
		},
	}
	return views.NewLibrary(state, page0, page1, page2)
}

func TestLibraryViewType(t *testing.T) {
	lib := testLibrary()
	if lib.Type() != views.LibraryViewType {
		t.Errorf("expected type %q, got %q", views.LibraryViewType, lib.Type())
	}
}

func TestLibraryInitialRender(t *testing.T) {
	lib := testLibrary()
	content := lib.View().Content
	// Breadcrumb should show the first page title
	if !strings.Contains(content, "Categories") {
		t.Error("expected breadcrumb with 'Categories'")
	}
	// Table content from page 0
	if !strings.Contains(content, "Alpha") {
		t.Error("expected page 0 table row 'Alpha'")
	}
	if !strings.Contains(content, "Beta") {
		t.Error("expected page 0 table row 'Beta'")
	}
}

func TestLibraryNavigateForward(t *testing.T) {
	lib := testLibrary()

	// Navigate forward from page 0 -> page 1
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	content := lib.View().Content

	// Breadcrumb should show both pages
	if !strings.Contains(content, "Categories") {
		t.Error("expected breadcrumb with 'Categories'")
	}
	if !strings.Contains(content, "Items") {
		t.Error("expected breadcrumb with 'Items'")
	}
	// Page 1 content derived from page 0 selection ("Alpha")
	if !strings.Contains(content, "Alpha-item-1") {
		t.Error("expected page 1 row 'Alpha-item-1'")
	}
}

func TestLibraryNavigateBackward(t *testing.T) {
	lib := testLibrary()

	// Go to page 1
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	// Go back to page 0
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	content := lib.View().Content
	if !strings.Contains(content, "Alpha") {
		t.Error("expected page 0 content after navigating back")
	}
	if strings.Contains(content, "Alpha-item-1") {
		t.Error("should not show page 1 content after navigating back")
	}
}

func TestLibraryCapturingInput(t *testing.T) {
	lib := testLibrary()
	ic, ok := any(lib).(tuikit.InputCapturer)
	if !ok {
		t.Fatal("Library should implement InputCapturer")
	}

	// On page 0, not capturing
	if ic.CapturingInput() {
		t.Error("should not be capturing input on page 0")
	}

	// Navigate to page 1
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !ic.CapturingInput() {
		t.Error("should be capturing input on page > 0")
	}

	// Navigate back to page 0
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if ic.CapturingInput() {
		t.Error("should not be capturing input after returning to page 0")
	}
}

func TestLibraryHelpBindings(t *testing.T) {
	lib := testLibrary()

	// Page 0: should have drill-down key but not go-back
	keys := helpKeys(lib)
	if !keys["enter/→"] {
		t.Error("expected 'enter/→' help binding on page 0")
	}
	if keys["esc/<-"] {
		t.Error("should not have 'esc/<-' on page 0")
	}

	// Navigate to page 1 (middle page): should have both
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	keys = helpKeys(lib)
	if !keys["enter/→"] {
		t.Error("expected 'enter/→' on middle page")
	}
	if !keys["esc/←"] {
		t.Error("expected 'esc/←' on middle page")
	}
	// Domain callback key from page 1
	if !keys["x"] {
		t.Error("expected domain key 'x' on page 1")
	}
}

func TestLibraryLastPageForwardsEnter(t *testing.T) {
	lib := testLibrary()

	// Navigate to page 1, then page 2 (last page)
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	content := lib.View().Content
	if !strings.Contains(content, "Detail for Alpha-item-1") {
		t.Errorf("expected detail page content, got %q", content)
	}

	// Help should not have drill-down on last page
	keys := helpKeys(lib)
	if keys["enter/→"] {
		t.Error("should not have 'enter/→' on last page")
	}
	if !keys["esc/←"] {
		t.Error("expected 'esc/←' on last page")
	}
}

func TestLibrarySubviewFilterCapture(t *testing.T) {
	lib := testLibrary()
	ic, ok := any(lib).(tuikit.InputCapturer)
	if !ok {
		t.Fatal("Library should implement InputCapturer")
	}

	// Activate table filter on page 0
	lib.Update(tea.KeyPressMsg{Text: "/"})
	if !ic.CapturingInput() {
		t.Error("should be capturing input when sub-view filter is active")
	}

	// Esc should close filter, not propagate
	lib.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if ic.CapturingInput() {
		t.Error("should stop capturing after filter closed")
	}
}

// --- Integration test ---
// The form test needs the full bubbletea lifecycle to verify
// interactive input handling and view transitions.

func TestFormInteraction(t *testing.T) {
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(t.Context(), app)
	if err != nil {
		t.Fatal(err)
	}

	tm := teatest.NewTestModel(t, container,
		teatest.WithInitialTermSize(80, 80),
		teatest.WithProgramOptions(
			tea.WithColorProfile(colorprofile.Ascii),
			tea.WithEnvironment([]string{"NO_COLOR=1", "TERM=dumb"}),
		),
	)
	container.SetSendFunc(tm.Send)
	view, err := views.NewFormView(
		container.RenderState(),
		&views.FormField{
			Key:      "confirm",
			Title:    "Are you sure?",
			Required: true,
			Type:     views.PromptTypeConfirm,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	view.Callback = func(v map[string]any) error {
		inner := &sampleTypes.Echo{Content: "Thank you for confirming!"}
		next := views.NewFrameView(inner)
		container.SetNextView(next)
		return nil
	}
	if err := container.SetView(view); err != nil {
		t.Fatal(err)
	}

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Are you sure?"))
	}, teatest.WithCheckInterval(100*time.Millisecond), teatest.WithDuration(3*time.Second))
	container.Send(tea.KeyPressMsg{Code: tea.KeyEnter}, 100*time.Millisecond)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Thank you for confirming!"))
	}, teatest.WithCheckInterval(100*time.Millisecond), teatest.WithDuration(5*time.Second))
	container.Send(tea.KeyPressMsg{Text: "q"}, 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
