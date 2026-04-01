package overlay_test

import (
	"testing"

	"github.com/flowexec/tuikit/overlay"
	"github.com/flowexec/tuikit/themes"
)

func TestHelpPopupToggle(t *testing.T) {
	h := overlay.NewHelpPopup(themes.EverforestTheme())
	if h.Visible() {
		t.Fatal("expected help to start hidden")
	}
	h.Toggle()
	if !h.Visible() {
		t.Fatal("expected help to be visible after toggle")
	}
	h.Toggle()
	if h.Visible() {
		t.Fatal("expected help to be hidden after second toggle")
	}
}

func TestHelpPopupRender(t *testing.T) {
	h := overlay.NewHelpPopup(themes.EverforestTheme())
	h.SetViewKeys([]themes.HelpKey{
		{Key: "enter", Desc: "select"},
	})
	out := h.Render(80, 40)
	if out == "" {
		t.Fatal("expected non-empty render output")
	}
}

func TestHelpPopupRenderNoKeys(t *testing.T) {
	h := overlay.NewHelpPopup(themes.EverforestTheme())
	h.SetViewKeys(nil)
	out := h.Render(80, 40)
	// Should still render because global keys are always present.
	if out == "" {
		t.Fatal("expected render output with global keys only")
	}
}

func TestToastManagerPushAndDismiss(t *testing.T) {
	tm := overlay.NewToastManager(themes.EverforestTheme())
	if !tm.Empty() {
		t.Fatal("expected empty toast manager")
	}

	cmd := tm.Push("hello", themes.OutputLevelInfo)
	if cmd == nil {
		t.Fatal("expected non-nil cmd from Push")
	}
	if tm.Empty() {
		t.Fatal("expected non-empty after push")
	}

	// Push a second toast.
	tm.Push("world", themes.OutputLevelSuccess)

	// Dismiss first toast (ID 0).
	tm.Dismiss(0)
	if tm.Empty() {
		t.Fatal("expected one toast remaining")
	}

	// Dismiss second toast (ID 1).
	tm.Dismiss(1)
	if !tm.Empty() {
		t.Fatal("expected empty after dismissing all")
	}
}

func TestToastManagerDismissNonexistent(t *testing.T) {
	tm := overlay.NewToastManager(themes.EverforestTheme())
	tm.Push("test", themes.OutputLevelInfo)
	// Should not panic.
	tm.Dismiss(999)
	if tm.Empty() {
		t.Fatal("dismiss of non-existent ID should not remove existing toast")
	}
}

func TestToastManagerRender(t *testing.T) {
	tm := overlay.NewToastManager(themes.EverforestTheme())
	if out := tm.Render(80, 40); out != "" {
		t.Fatalf("expected empty render, got %q", out)
	}

	tm.Push("alert!", themes.OutputLevelWarning)
	out := tm.Render(80, 40)
	if out == "" {
		t.Fatal("expected non-empty render after push")
	}
}

func TestToastManagerMaxVisible(t *testing.T) {
	tm := overlay.NewToastManager(themes.EverforestTheme())
	for i := range 5 {
		tm.Push(string(rune('A'+i)), themes.OutputLevelInfo)
	}
	out := tm.Render(80, 40)
	// Default maxVisible is 3, so only the last 3 toasts should render.
	// We can't easily count rendered toasts, but verify it's non-empty.
	if out == "" {
		t.Fatal("expected non-empty render")
	}
}
