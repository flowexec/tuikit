package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/flowexec/tuikit"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func main() {
	var viewType string
	flag.StringVar(&viewType, "view", "frame", "view type")
	flag.Parse()

	ctx := context.Background()
	app := &tuikit.Application{Name: "tuikit-sample"}

	container, err := tuikit.NewContainer(ctx, app)
	if err != nil {
		panic(err)
	}
	if err := container.Start(); err != nil {
		panic(err)
	}

	view := buildView(viewType, container)
	if err := container.SetView(view); err != nil {
		panic(err)
	}
	container.WaitForExit()
}

func buildView(viewType string, container *tuikit.Container) tuikit.View {
	var view tuikit.View
	switch viewType {
	case "frame":
		inner := &sampleTypes.Echo{
			Content: "You are currently viewing a rendered frame. " +
				"Use the --view flag to switch to a different view.",
		}
		view = views.NewFrameView(inner)
	case "loading":
		view = views.NewLoadingView(
			"waiting for the paint to dry...",
			container.RenderState().Theme,
		)
	case "error":
		view = views.NewErrorView(
			errors.New("something went wrong - please try again"),
			container.RenderState().Theme,
		)
	case "markdown":
		md := "# Hmmm...\n\n > To be, or not to be, " +
			"**that is the question**.\n> *William Shakespeare*"
		view = views.NewMarkdownView(container.RenderState(), md)
	case "entity":
		e := &sampleTypes.Thing{Name: "William Shakespeare", Type: "Author"}
		view = views.NewEntityView(
			container.RenderState(), e, types.EntityFormatDocument,
		)
	case "collection":
		view = buildCollectionView(container)
	case "detail":
		view = buildDetailView(container)
	case "table":
		view = buildTableFullView(container)
	case "table-mini":
		view = buildTableMiniView(container)
	case "table-mini-multi":
		view = buildTableMiniMultiView(container)
	case "form":
		view = buildFormView(container)
	}
	return view
}

func buildDetailView(container *tuikit.Container) tuikit.View {
	body := `2026-03-30 14:22:01 [INFO]  deploy-pipeline read secret DATABASE_URL
2026-03-30 08:15:33 [INFO]  api-server read secret DATABASE_URL
2026-03-29 22:00:00 [WARN]  rotation check: 75 days remaining
2026-03-28 16:45:12 [INFO]  api-server read secret DATABASE_URL
2026-03-28 09:30:00 [INFO]  deploy-pipeline read secret DATABASE_URL`

	return views.NewDetailView(
		container.RenderState(),
		body,
		views.DetailField{Key: "Name", Value: "DATABASE_URL"},
		views.DetailField{Key: "Environment", Value: "production"},
		views.DetailField{Key: "Created", Value: "2026-01-15 10:30:00"},
		views.DetailField{Key: "Rotation", Value: "90 days"},
	)
}

func buildCollectionView(container *tuikit.Container) tuikit.View {
	c := sampleTypes.NewThingList("Author",
		&types.EntityInfo{
			ID: "william", Header: "William Shakespeare",
			SubHeader: "English Playwright",
		},
		&types.EntityInfo{
			ID: "jane", Header: "Jane Austen",
			SubHeader: "English Novelist",
		},
		&types.EntityInfo{
			ID: "mark", Header: "Mark Twain",
			SubHeader: "American Author",
		},
	)
	return views.NewCollectionView(
		container.RenderState(), c, types.CollectionFormatList, nil,
	)
}

func buildTableFullView(container *tuikit.Container) tuikit.View {
	columns := []views.TableColumn{
		{Title: "Workspace", Percentage: 40},
		{Title: "Description", Percentage: 35},
		{Title: "Status", Percentage: 25},
	}
	rows := []views.TableRow{
		{
			Data: []string{"flow-workspace", "Main development workspace", "Active"},
			Children: []views.TableRow{
				{Data: []string{"docs", "", "5 exec"}},
				{Data: []string{"api", "", "12 exec"}},
				{Data: []string{"frontend", "", "8 exec"}},
			},
		},
		{
			Data: []string{"home-lab", "Infrastructure automation", "Inactive"},
			Children: []views.TableRow{
				{Data: []string{"k8s", "", "15 exec"}},
				{Data: []string{"monitoring", "", "6 exec"}},
			},
		},
		{
			Data:     []string{"personal-tools", "Personal utility scripts", "Active"},
			Children: []views.TableRow{},
		},
	}

	table := views.NewTable(
		container.RenderState(), columns, rows, views.TableDisplayFull,
	)
	table.SetOnSelect(func(index int) error {
		selectedRow := table.GetSelectedRow()
		if selectedRow != nil {
			container.SetNotice(
				fmt.Sprintf("Selected: %s", selectedRow.Data()[0]),
				themes.OutputLevelInfo,
			)
		}
		return nil
	})
	table.SetOnHover(func(index int) {
		selectedRow := table.GetSelectedRow()
		if selectedRow != nil {
			container.SetState("Current", selectedRow.Data()[0])
		}
	})
	return table
}

func buildTableMiniView(container *tuikit.Container) tuikit.View {
	columns := []views.TableColumn{
		{Title: "Available Executables", Percentage: 100},
	}
	rows := []views.TableRow{
		{Data: []string{"build app"}},
		{Data: []string{"test unit"}},
		{Data: []string{"deploy staging"}},
		{Data: []string{"deploy production"}},
		{Data: []string{"clean artifacts"}},
	}

	table := views.NewTable(
		container.RenderState(), columns, rows, views.TableDisplayMini,
	)
	table.SetOnSelect(func(index int) error {
		selectedRow := table.GetSelectedRow()
		if selectedRow != nil {
			container.SetNotice(
				fmt.Sprintf("Executing: %s", selectedRow.Data()[0]),
				themes.OutputLevelInfo,
			)
		}
		return nil
	})
	return table
}

func buildTableMiniMultiView(container *tuikit.Container) tuikit.View {
	columns := []views.TableColumn{
		{Title: "Template", Percentage: 60},
		{Title: "Type", Percentage: 40},
	}
	rows := []views.TableRow{
		{Data: []string{"k8s-deployment", "Kubernetes"}},
		{Data: []string{"react-app", "Frontend"}},
		{Data: []string{"go-service", "Backend"}},
		{Data: []string{"terraform-module", "Infrastructure"}},
	}

	table := views.NewTable(
		container.RenderState(), columns, rows, views.TableDisplayMini,
	)
	table.SetOnSelect(func(index int) error {
		selectedRow := table.GetSelectedRow()
		if selectedRow != nil {
			msg := fmt.Sprintf(
				"Selected template: %s (%s)",
				selectedRow.Data()[0], selectedRow.Data()[1],
			)
			container.SetNotice(msg, themes.OutputLevelInfo)
		}
		return nil
	})
	return table
}

func buildFormView(container *tuikit.Container) tuikit.View {
	f, err := views.NewFormView(
		container.RenderState(),
		&views.FormField{
			Key:      "author",
			Title:    "Favorite Author",
			Required: true,
		},
		&views.FormField{
			Key:         "color",
			Title:       "Favorite Color",
			Default:     "pink",
			Required:    false,
			Description: "hint: it's pink",
		},
		&views.FormField{
			Key:   "confirm",
			Title: "Ready to submit?",
			Type:  views.PromptTypeConfirm,
		},
	)
	if err != nil {
		panic(err)
	}
	f.Callback = func(v map[string]any) error {
		inner := &sampleTypes.Echo{
			Content: fmt.Sprintf(
				"Thank you for tell me that your favorite author "+
					"is %s and your favorite color is %s!",
				f.FindByKey("author").Value(),
				f.FindByKey("color").Value(),
			),
		}
		container.SetNextView(views.NewFrameView(inner))
		return nil
	}
	return f
}
