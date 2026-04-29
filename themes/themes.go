package themes

import (
	_ "embed"
	"os"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
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
	return baseTheme{
		Name:        name,
		SpinnerType: spinner.Points,
		Colors:      &colors,
		isDark:      detectDarkBackground(),
	}
}

// detectDarkBackground checks if the terminal has a dark background.
// It only queries the terminal when both stdin and stdout are real TTYs,
// avoiding hangs in CI environments and headless contexts where the
// terminal won't respond to OSC escape sequences.
func detectDarkBackground() bool {
	if !term.IsTerminal(os.Stdin.Fd()) || !term.IsTerminal(os.Stdout.Fd()) {
		return true // default to dark when not a real terminal
	}
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
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
	return NewTheme(everforest, ColorPalette{
		Secondary: "#A7C080", // yellow-green — more distinct from teal primary than default sage
	})
}

// DarkTheme Use colors from GitHub Dark
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dark.json
func DarkTheme() Theme {
	return NewTheme(
		dark,
		ColorPalette{
			ChromaCodeStyle: "github-dark",
			Primary:         "#539BF5",
			Secondary:       "#57AB5A",
			Tertiary:        "#986EE2",
			Emphasis:        "#E88B2E",
			Info:            "#539BF5",
			Body:            "#E6EDF3",
			Border:          "#30363D",
			Warning:         "#D29922",
			Gray:            "#768390",
			AppName:         "#D897FF",
			Black:           "#0D1117",
			White:           "#F0F6FF",
		})
}

// DraculaTheme Use colors from Glamour Dracula theme
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/dracula.json
func DraculaTheme() Theme {
	return NewTheme(
		dracula,
		ColorPalette{
			ChromaCodeStyle: "dracula",
			Primary:         "#bd93f9",
			Secondary:       "#8be9fd",
			Tertiary:        "#50fa7b",
			Emphasis:        "#f1fa8c",
			Info:            "#bd93f9",
			Body:            "#f8f8f2",
			Border:          "#6272A4",
			Warning:         "#ffb86c",
			Error:           "#ff5555",
			Gray:            "#8694AA",
			AppName:         "#ff79c6",
			Black:           "#21222C",
			White:           "#F8F8F2",
		},
	)
}

// LightTheme Use colors from GitHub Light
// See https://raw.githubusercontent.com/charmbracelet/glamour/refs/heads/master/styles/light.json
func LightTheme() Theme {
	return NewTheme(
		light,
		ColorPalette{
			ChromaCodeStyle: "github",
			Primary:         "#0969DA",
			Secondary:       "#1F883D",
			Tertiary:        "#8250DF",
			Emphasis:        "#9A3FB4",
			Info:            "#0969DA",
			Body:            "#24292F",
			Border:          "#D0D7DE",
			Warning:         "#BF8700",
			Gray:            "#959da5",
			AppName:         "#6f42c1",
			Black:           "#1F2328",
			White:           "#F6F8FA",
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
			Primary:         "#bb9af7",
			Secondary:       "#7aa2f7",
			Tertiary:        "#2ac3de",
			Emphasis:        "#ff9e64",
			Info:            "#bb9af7",
			Body:            "#a9b1d6",
			Border:          "#565f89",
			Warning:         "#e0af68",
			Gray:            "#737aa2",
			AppName:         "#f7768e",
			Black:           "#1A1B26",
			White:           "#C0CAF5",
		},
	)
}
