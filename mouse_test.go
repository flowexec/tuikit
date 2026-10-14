package tuikit_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/views"
)

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

func TestContainerMouseModeOptIn(t *testing.T) {
	app := tuikit.NewApplication("Test")
	for _, tc := range []struct {
		name string
		opts []tuikit.ContainerOptions
		want tea.MouseMode
	}{
		{"default", nil, tea.MouseModeNone},
		{"enabled", []tuikit.ContainerOptions{tuikit.WithMouse()}, tea.MouseModeCellMotion},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := append([]tuikit.ContainerOptions{
				tuikit.WithInitialTermSize(80, 40),
				tuikit.WithOutput(&bytes.Buffer{}),
			}, tc.opts...)
			container, err := tuikit.NewContainer(t.Context(), app, opts...)
			if err != nil {
				t.Fatal(err)
			}
			container.Init()
			if got := container.View().MouseMode; got != tc.want {
				t.Errorf("expected mouse mode %v, got %v", tc.want, got)
			}
		})
	}
}

func TestContainerMouseClickTranslation(t *testing.T) {
	app := tuikit.NewApplication("Test")
	container, err := tuikit.NewContainer(t.Context(), app,
		tuikit.WithInitialTermSize(80, 40),
		tuikit.WithOutput(&bytes.Buffer{}),
		tuikit.WithMouse(),
	)
	if err != nil {
		t.Fatal(err)
	}
	container.SetSendFunc(func(tea.Msg) {})
	container.Init()
	columns := []views.TableColumn{{Title: "Item", Percentage: 100}}
	table := views.NewTable(container.RenderState(), columns, numberedRows(3), views.TableDisplayFull)
	if err := container.SetView(table); err != nil {
		t.Fatal(err)
	}

	// Coordinates are absolute, including the app header.
	container.Update(leftClick(5, lineOf(t, container.View().Content, "row-01")))
	if got := table.GetSelectedRow().Data()[0]; got != "row-01" {
		t.Errorf("expected click to select row-01, got %s", got)
	}

	// Clicks on the header are dropped.
	container.Update(leftClick(5, 0))
	if got := table.GetSelectedRow().Data()[0]; got != "row-01" {
		t.Errorf("expected header click to be ignored, got %s", got)
	}
}
