package tuikit_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/flowexec/tuikit"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func TestMain(m *testing.M) {
	// Force ASCII color profile so lipgloss strips all color/style escape
	// sequences, producing identical output regardless of the terminal
	// capabilities of the host (local dev vs CI).
	os.Setenv("NO_COLOR", "1")
	os.Setenv("TERM", "dumb")
	lipgloss.Writer.Profile = colorprofile.Ascii
	os.Exit(m.Run())
}

func TestFrameOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	view := views.NewFrameView(&sampleTypes.Echo{Content: "Hello, world!"})
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestLoadingOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	view := views.NewLoadingView("thinking...", container.RenderState().Theme)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestErrorOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	view := views.NewErrorView(errors.New("something went wrong - please try again"), container.RenderState().Theme)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestMarkdownOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	md := "# Hello!\n\n I am a **Markdown** document."
	view := views.NewMarkdownView(container.RenderState(), md)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestEntityOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	e := &sampleTypes.Thing{Name: "Green", Type: "Color"}
	view := views.NewEntityView(container.RenderState(), e, types.EntityFormatDocument)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestCollectionOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	collection := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red", SubHeader: "Primary Color"},
		&types.EntityInfo{ID: "green", Header: "Green", SubHeader: "Secondary Color"},
		&types.EntityInfo{ID: "blue-violet", Header: "Blue Violet", SubHeader: "Tertiary Color"},
	)
	view := views.NewCollectionView(container.RenderState(), collection, types.CollectionFormatList, nil)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 10*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestDetailOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	view := views.NewDetailView(
		container.RenderState(),
		"This is the body content of the detail view.",
		views.DetailField{Key: "Name", Value: "Test Item"},
		views.DetailField{Key: "Status", Value: "Active"},
	)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestTableOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
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
	view := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayFull)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestTableMiniOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
	container.SetSendFunc(tm.Send)
	columns := []views.TableColumn{
		{Title: "Command", Percentage: 100},
	}
	rows := []views.TableRow{
		{Data: []string{"build"}},
		{Data: []string{"test"}},
		{Data: []string{"deploy"}},
	}
	view := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
	}

	container.Send(tea.Quit(), 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

func TestFormOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}
	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		t.Errorf("Failed to create container: %v", err)
	}

	tm := teatest.NewTestModel(t, container, teatest.WithInitialTermSize(80, 80))
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
		t.Errorf("Failed to create form view: %v", err)
	}
	view.Callback = func(v map[string]any) error {
		inner := &sampleTypes.Echo{Content: "Thank you for confirming!"}
		next := views.NewFrameView(inner)
		container.SetNextView(next)
		return nil
	}
	if err := container.SetView(view); err != nil {
		t.Errorf("Failed to set view: %v", err)
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
