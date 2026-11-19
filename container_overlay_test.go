package tuikit_test

import (
	"strings"
	"testing"

	"bloomfx.dev/bloom/anim"
	tea "charm.land/bubbletea/v2"

	"github.com/flowexec/tuikit"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func TestContainerOverlays(t *testing.T) {
	anim.SetMotion(false)
	t.Cleanup(func() { anim.SetMotion(true) })

	container, err := tuikit.NewContainer(
		t.Context(), &tuikit.Application{Name: "tuikit-test"}, tuikit.WithInitialTermSize(80, 24),
	)
	if err != nil {
		t.Fatal(err)
	}
	container.SetSendFunc(func(tea.Msg) {})
	if err := container.SetView(views.NewFrameView(&sampleTypes.Echo{Content: "body text"})); err != nil {
		t.Fatal(err)
	}
	if err := container.SetView(views.NewMarkdownView(container.RenderState(), "body text")); err != nil {
		t.Fatal(err)
	}

	content := func() string { return container.View().Content }
	if c := content(); !strings.Contains(c, "body text") || !strings.Contains(c, "tuikit-test") {
		t.Fatalf("expected header and body, got:\n%s", c)
	}

	container.Update(tea.KeyPressMsg{Text: "?", Code: '?'})
	if c := content(); !strings.Contains(c, "toggle help") || !strings.Contains(c, "tuikit-test") {
		t.Fatalf("expected help over the base view, got:\n%s", c)
	}
	container.Update(tea.KeyPressMsg{Text: "?", Code: '?'})
	if c := content(); strings.Contains(c, "toggle help") {
		t.Fatalf("expected help closed, got:\n%s", c)
	}

	container.Update(types.ToastMsg{Text: "saved it", Level: themes.OutputLevelSuccess})
	toast := content()
	if !strings.Contains(toast, "saved it") {
		t.Fatalf("expected toast, got:\n%s", toast)
	}
	lines := strings.Split(toast, "\n")
	if row := len(lines) - 2; !strings.Contains(lines[row], "└") {
		t.Errorf("expected toast box to end one row above the bottom edge, got:\n%s", toast)
	}
	container.Update(types.ToastDismissMsg{ID: 0})
	if c := content(); strings.Contains(c, "saved it") {
		t.Fatalf("expected toast dismissed, got:\n%s", c)
	}
}
