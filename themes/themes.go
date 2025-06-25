package themes

import (
	_ "embed"

	"github.com/charmbracelet/bubbles/spinner"
)

const (
	everforest = "everforest"
	dark       = "dark"
	dracula    = "dracula"
	light      = "light"
	tokyoNight = "tokyo-night"
)

func NewTheme(name string, cp ColorPalette) Theme {
	colors := WithDefaultColors(cp)
	return baseTheme{Name: name, SpinnerType: spinner.Points, Colors: &colors}
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
	return NewTheme(everforest, ColorPalette{}) // inherit from defaults
}

// DarkTheme Use colors from Glamour Dracula theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dark.json
func DarkTheme() Theme {
	return NewTheme(
		dark,
		ColorPalette{
			ChromaCodeStyle: "github-dark",
			Body:            "252",
			Border:          "240",
			Emphasis:        "30",
			Primary:         "39",
			Secondary:       "228",
			Tertiary:        "63",
			Info:            "39",
		})
}

// DraculaTheme Use colors from Glamour Dracula theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dracula.json
func DraculaTheme() Theme {
	return NewTheme(
		dracula,
		ColorPalette{
			ChromaCodeStyle: "dracula",
			Body:            "#f8f8f2",
			Border:          "#6272A4",
			Emphasis:        "#f1fa8c",
			Primary:         "#bd93f9",
			Secondary:       "#8be9fd",
			Tertiary:        "#ffb86c",
			Info:            "#bd93f9",
		},
	)
}

// LightTheme Use colors from Glamour Light theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/light.json
func LightTheme() Theme {
	return NewTheme(
		light,
		ColorPalette{
			ChromaCodeStyle: "github",
			Body:            "#ffffff",
			Border:          "#e1e4e8",
			Emphasis:        "#0366d6",
			Primary:         "#24292e",
			Secondary:       "#586069",
			Tertiary:        "#6a737d",
			Info:            "#0366d6",
		},
	)
}

// TokyoNightTheme Use colors from Glamour Tokyo Night theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/tokyo-night.json
func TokyoNightTheme() Theme {
	return NewTheme(
		tokyoNight,
		ColorPalette{
			ChromaCodeStyle: "tokyonight-night",
			Body:            "#a9b1d6",
			Border:          "#565f89",
			Emphasis:        "#7aa2f7",
			Primary:         "#bb9af7",
			Secondary:       "#7aa2f7",
			Tertiary:        "#2ac3de",
			Info:            "#bb9af7",
		},
	)
}
