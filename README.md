# Terminal UI Kit 

[![Go Report Card](https://goreportcard.com/badge/github.com/jahvon/tuikit)](https://goreportcard.com/report/github.com/jahvon/tuikit)
[![Go Reference](https://pkg.go.dev/badge/github.com/jahvon/tuikit.svg)](https://pkg.go.dev/github.com/jahvon/tuikit)
[![GitHub release](https://img.shields.io/github/v/release/jahvon/tuikit)](https://github.com/jahvon/tuikit/releases)

This repo contains types, interfaces, and utilities for building terminal user interfaces in Go.
It's an opinionated framework that uses [charm](https://charm.sh) TUI components and packages for rendering
and handling terminal events.

## Usage

First, install the package:

```bash
go get -u github.com/jahvon/tuikit@latest
```

You can then use the package in your Go code:

```go
package main

import (
    "context"

    "github.com/jahvon/tuikit"
    "github.com/jahvon/tuikit/views"
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

Also see the [sample app](sample/main.go) for examples of how different views can be used.
