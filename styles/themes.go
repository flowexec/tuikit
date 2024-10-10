package styles

import (
	_ "embed"

	"github.com/charmbracelet/bubbles/spinner"
)

//go:embed mdstyles.tmpl.json
var mdstylesTemplateJSON string

const (
	everforest = "everforest"
	dark       = "dark"
	dracula    = "dracula"
	light      = "light"
	tokyoNight = "tokyo-night"
)

func BaseTheme() Theme {
	theme := Theme{
		ChromaCodeStyle: "friendly",
		EmphasisColor:   "#E67E80",
		SuccessColor:    "#8DA101",
		WarningColor:    "#5C6A72",
		InfoColor:       "#3A94C5",
		ErrorColor:      "#F85552",
		White:           "#DFDDC8",
		Gray:            "#5C6A72",
		Black:           "#343F44",
		SpinnerType:     spinner.Points,
	}
	return theme
}

type ThemeFunc func() Theme

func AllThemes() map[string]ThemeFunc {
	return map[string]ThemeFunc{
		everforest: EverforestTheme,
		dark:       DarkTheme,
		dracula:    DraculaTheme,
		light:      LightTheme,
		tokyoNight: TokyoNightTheme,
	}
}

// EverforestTheme Uses the colors from the Everforest color scheme
// See https://gogh-co.github.io/Gogh/
func EverforestTheme() Theme {
	theme := BaseTheme()
	theme.Name = everforest
	theme.BodyColor = "#D3C6AA"
	theme.EmphasisColor = "#E67E80"
	theme.BorderColor = "#5C6A72"
	theme.PrimaryColor = "#7FBBB3"
	theme.SecondaryColor = "#83C092"
	theme.TertiaryColor = "#D699B6"
	return theme
}

// DarkTheme Use colors from Glamour Dracula theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dark.json
func DarkTheme() Theme {
	theme := BaseTheme()
	theme.Name = dark
	theme.ChromaCodeStyle = "github-dark"
	theme.BodyColor = "252"
	theme.BorderColor = "240"
	theme.EmphasisColor = "30"
	theme.PrimaryColor = "39"
	theme.SecondaryColor = "228"
	theme.TertiaryColor = "63"
	theme.InfoColor = "39"
	return theme
}

// DraculaTheme Use colors from Glamour Dracula theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dracula.json
func DraculaTheme() Theme {
	theme := BaseTheme()
	theme.Name = dracula
	theme.ChromaCodeStyle = "dracula"
	theme.BodyColor = "#f8f8f2"
	theme.BorderColor = "#6272A4"
	theme.EmphasisColor = "#f1fa8c"
	theme.PrimaryColor = "#bd93f9"
	theme.SecondaryColor = "#8be9fd"
	theme.TertiaryColor = "#ffb86c"
	theme.InfoColor = "#bd93f9"
	return theme
}

// LightTheme Use colors from Glamour Light theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/light.json
func LightTheme() Theme {
	theme := BaseTheme()
	theme.ChromaCodeStyle = "github"
	theme.Name = light
	theme.BodyColor = "234"
	theme.BorderColor = "249"
	theme.EmphasisColor = "36"
	theme.PrimaryColor = "228"
	theme.SecondaryColor = "27"
	theme.TertiaryColor = "205"
	theme.InfoColor = "228"
	return theme
}

// TokyoNightTheme Use colors from Glamour Tokyo Night theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/tokyo-night.json
func TokyoNightTheme() Theme {
	theme := BaseTheme()
	theme.ChromaCodeStyle = "tokyonight-night"
	theme.Name = tokyoNight
	theme.BodyColor = "#a9b1d6"
	theme.BorderColor = "#565f89"
	theme.EmphasisColor = "#7aa2f7"
	theme.PrimaryColor = "#bb9af7"
	theme.SecondaryColor = "#7aa2f7"
	theme.TertiaryColor = "#2ac3de"
	theme.InfoColor = "#bb9af7"
	return theme
}
