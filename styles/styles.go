package styles

import (
	_ "embed"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

//go:embed mdstyles.json
var BaseMarkdownStylesJSON string

// BaseTheme Inspired by Wryan and Jellybeans https://gogh-co.github.io/Gogh/
func BaseTheme() Theme {
	return Theme{
		PrimaryColor:      lipgloss.AdaptiveColor{Dark: "#477AB3", Light: "#31658C"},
		SecondaryColor:    lipgloss.AdaptiveColor{Dark: "#7E62B3", Light: "#5E468C"},
		TertiaryColor:     lipgloss.AdaptiveColor{Dark: "#9E9ECB", Light: "#7C7C99"},
		WarningColor:      lipgloss.AdaptiveColor{Dark: "#FFDCA0", Light: "#FFBA7B"},
		InfoColor:         lipgloss.AdaptiveColor{Dark: "#53A6A6", Light: "#287373"},
		ErrorColor:        lipgloss.AdaptiveColor{Dark: "#BF4D80", Light: "#8C4665"},
		White:             lipgloss.AdaptiveColor{Dark: "#C0C0C0", Light: "#899CA1"},
		Gray:              lipgloss.AdaptiveColor{Dark: "#3D3D3D", Light: "#333333"},
		Black:             lipgloss.AdaptiveColor{Dark: "#2b2a2a", Light: "#000000"},
		SpinnerType:       spinner.Points,
		MarkdownStyleJSON: BaseMarkdownStylesJSON,
	}
}
