//nolint:cyclop
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/flowexec/tuikit"
	sampleTypes "github.com/flowexec/tuikit/sample/types"
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

	var view tuikit.View
	switch viewType {
	case "frame":
		inner := &sampleTypes.Echo{
			Content: "You are currently viewing a rendered frame. Use the --view flag to switch to a different view.",
		}
		view = views.NewFrameView(inner)
	case "loading":
		view = views.NewLoadingView("waiting for the paint to dry...", container.RenderState().Theme)
	case "error":
		view = views.NewErrorView(errors.New("something went wrong - please try again"), container.RenderState().Theme)
	case "markdown":
		md := "# Hmmm...\n\n > To be, or not to be, **that is the question**.\n> *William Shakespeare*"
		view = views.NewMarkdownView(container.RenderState(), md)
	case "entity":
		e := &sampleTypes.Thing{Name: "William Shakespeare", Type: "Author"}
		view = views.NewEntityView(container.RenderState(), e, types.EntityFormatDocument)
	case "collection":
		c := sampleTypes.NewThingList("Author",
			&types.EntityInfo{ID: "william", Header: "William Shakespeare", SubHeader: "English Playwright"},
			&types.EntityInfo{ID: "jane", Header: "Jane Austen", SubHeader: "English Novelist"},
			&types.EntityInfo{ID: "mark", Header: "Mark Twain", SubHeader: "American Author"},
		)
		view = views.NewCollectionView(container.RenderState(), c, types.CollectionFormatList, nil)
	case "form":
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
					"Thank you for tell me that your favorite author is %s and your favorite color is %s!",
					f.FindByKey("author").Value(),
					f.FindByKey("color").Value(),
				),
			}
			view = views.NewFrameView(inner)
			container.SetNextView(view)
			return nil
		}
		view = f
	}

	if err := container.SetView(view); err != nil {
		panic(err)
	}
	container.WaitForExit()
}
