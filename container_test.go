package tuikit_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"

	"github.com/jahvon/tuikit"
	sampleTypes "github.com/jahvon/tuikit/sample/types"
	"github.com/jahvon/tuikit/types"
	"github.com/jahvon/tuikit/views"
)

func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
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

	container.Send(tea.Quit(), 500*time.Millisecond)
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
	container.Send(tea.KeyMsg{Type: tea.KeyEnter}, 100*time.Millisecond)
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Thank you for confirming!"))
	}, teatest.WithCheckInterval(100*time.Millisecond), teatest.WithDuration(5*time.Second))
	container.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}, 100*time.Millisecond)
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}
