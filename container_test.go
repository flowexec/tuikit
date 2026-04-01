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
		ContentHeight: 35,
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
