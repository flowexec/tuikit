package styles

import (
	_ "embed"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

//go:embed mdstyles.tmpl.json
var mdstylesTemplateJSON string

// EverforestTheme Uses the colors from the Everforest color scheme
// See https://gogh-co.github.io/Gogh/
func EverforestTheme() Theme {
	return Theme{
		BodyColor:      lipgloss.AdaptiveColor{Dark: "#D3C6AA", Light: "#343F44"},
		EmphasisColor:  lipgloss.AdaptiveColor{Dark: "#E67E80", Light: "#E67E80"},
		BorderColor:    lipgloss.AdaptiveColor{Dark: "#5C6A72", Light: "#D3C6AA"},
		PrimaryColor:   lipgloss.AdaptiveColor{Dark: "#7FBBB3", Light: "#7FBBB3"},
		SecondaryColor: lipgloss.AdaptiveColor{Dark: "#83C092", Light: "#83C092"},
		TertiaryColor:  lipgloss.AdaptiveColor{Dark: "#D699B6", Light: "#D699B6"},
		SuccessColor:   lipgloss.AdaptiveColor{Dark: "#8DA101", Light: "#8DA101"},
		WarningColor:   lipgloss.AdaptiveColor{Dark: "#DFA000", Light: "#DFA000"},
		InfoColor:      lipgloss.AdaptiveColor{Dark: "#3A94C5", Light: "#3A94C5"},
		ErrorColor:     lipgloss.AdaptiveColor{Dark: "#F85552", Light: "#F85552"},
		White:          lipgloss.AdaptiveColor{Dark: "#DFDDC8", Light: "#DFDDC8"},
		Gray:           lipgloss.AdaptiveColor{Dark: "#5C6A72", Light: "#5C6A72"},
		Black:          lipgloss.AdaptiveColor{Dark: "#343F44", Light: "#343F44"},
		SpinnerType:    spinner.Points,
	}
}
