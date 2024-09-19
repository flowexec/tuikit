package tuikit_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/jahvon/tuikit"
	sampleTypes "github.com/jahvon/tuikit/sample/types"
	"github.com/jahvon/tuikit/types"
	"github.com/jahvon/tuikit/views"
)

// TODO: DRY up the tests by creating a helper function that takes a view and a container and runs the test
//       Also consider using more teatest utils to make the tests more concise

func TestFrameOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	inner := &sampleTypes.Echo{Content: "Hello, world!"}
	view := views.NewFrameView(inner)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestLoadingOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	view := views.NewLoadingView("more loading...", *container.RenderState().Theme)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestErrorOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	view := views.NewErrorView(errors.New("something went wrong - please try again"), *container.RenderState().Theme)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestMarkdownOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	md := "# Hello!\n\n I am a **Markdown** document."
	view := views.NewMarkdownView(container.RenderState(), md)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestEntityOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	e := &sampleTypes.Thing{Name: "Green", Type: "Color"}
	view := views.NewEntityView(container.RenderState(), e, types.EntityFormatDocument)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestCollectionOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	c := sampleTypes.NewThingList("Color",
		&types.EntityInfo{ID: "red", Header: "Red", SubHeader: "Primary Color"},
		&types.EntityInfo{ID: "green", Header: "Green", SubHeader: "Secondary Color"},
		&types.EntityInfo{ID: "blue-violet", Header: "Blue Violet", SubHeader: "Tertiary Color"},
	)
	view := views.NewCollectionView(container.RenderState(), c, types.CollectionFormatList, nil)
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.QuitMsg{}, 100*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}

func TestFormOutput(t *testing.T) {
	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-test"}

	var buf bytes.Buffer
	rw := bufio.NewReadWriter(bufio.NewReader(&buf), bufio.NewWriter(&buf))
	container, err := tuikit.NewContainer(
		ctx, app,
		tuikit.WithOutput(rw),
		tuikit.WithInitialTermSize(80, 40),
	)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	if err := container.Start(); err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

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
		t.Fatalf("Failed to create form view: %v", err)
	}
	view.Callback = func(v map[string]any) error {
		inner := &sampleTypes.Echo{Content: "Thank you for confirming!"}
		next := views.NewFrameView(inner)
		container.SetNextView(next)
		return nil
	}
	if err := container.SetView(view); err != nil {
		t.Fatalf("Failed to set view: %v", err)
	}
	container.Send(tea.KeyMsg{Type: tea.KeyEnter}, 100*time.Millisecond)
	container.Send(tea.QuitMsg{}, 200*time.Millisecond)
	container.WaitForExit()
	if err := rw.Flush(); err != nil {
		t.Fatalf("Failed to flush buffer: %v", err)
	}
	teatest.RequireEqualOutput(t, buf.Bytes())
}
