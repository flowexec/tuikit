# Terminal UI Kit 

[![Go Reference](https://pkg.go.dev/badge/github.com/flowexec/tuikit.svg)](https://pkg.go.dev/github.com/flowexec/tuikit)
[![GitHub release](https://img.shields.io/github/v/release/flowexec/tuikit)](https://github.com/flowexec/tuikit/releases)

This repo contains types, interfaces, and utilities for building terminal user interfaces in Go.
It's an opinionated framework that uses [charm](https://charm.sh) TUI components and packages for rendering
and handling terminal events.

## Usage

First, install the package:

```bash
go get -u github.com/flowexec/tuikit@latest
```

You can then use the package in your Go code:

```go
package main

import (
    "context"

    "github.com/flowexec/tuikit"
    "github.com/flowexec/tuikit/views"
)

func main() {
    ctx := context.Background()
    // Define your application metadata
    app := &tuikit.Application{Name: "MyApp"}

    // Create and start the container
    container, err := tuikit.NewContainer(ctx, app)
    if err != nil {
        panic(err)
    }
    if err := container.Start(); err != nil {
        panic(err)
    }
    
    // Create and set your view - the example below used the Markdown view type.
    // There are other view types available in the views package.
    view := views.NewMarkdownView(container.RenderState(), "# Hello, world!")
    if err := container.SetView(view); err != nil {
        panic(err)
    }
    
    // Wait for the container to exit. Before getting here, you can handle events, update the view, etc.
    container.WaitForExit()
}
```

### Mouse and clipboard

Pass `tuikit.WithMouse()` to `NewContainer` to enable mouse support:

- The scroll wheel scrolls content and moves the selection in lists and tables.
- Clicking a list item or table row selects it; clicking it again opens it, just like pressing enter.
- Clicking an earlier breadcrumb in a `Library` view navigates back to that page.

Mouse support is off by default because, while it's enabled, most terminals need you to hold Shift
(Option on iTerm2) to select text. To make copying easy either way, views copy the selected item or
displayed content to the clipboard when `c` is pressed (unless a view's key callback uses `c`). Copying
writes both the native system clipboard and OSC 52, so it works in terminals that ignore OSC 52
(such as macOS Terminal) and over SSH; tmux needs `set -g set-clipboard on` for the OSC 52 path. Key
callbacks can copy the same way with `Container.CopyToClipboard`.

Also see the [sample app](sample/main.go) for examples of how different views can be used.
