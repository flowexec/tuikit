package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Theme struct {
	SpinnerType       spinner.Spinner
	MarkdownStyleJSON string
	PrimaryColor      lipgloss.AdaptiveColor
	SecondaryColor    lipgloss.AdaptiveColor
	TertiaryColor     lipgloss.AdaptiveColor
	WarningColor      lipgloss.AdaptiveColor
	ErrorColor        lipgloss.AdaptiveColor
	InfoColor         lipgloss.AdaptiveColor
	White             lipgloss.AdaptiveColor
	Gray              lipgloss.AdaptiveColor
	Black             lipgloss.AdaptiveColor
}

func (t *Theme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.SecondaryColor)
}

func (t *Theme) EntityView() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.TertiaryColor).
		MarginLeft(2)
}

func (t *Theme) Collection() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.TertiaryColor).
		MarginLeft(2).
		Padding(0, 1)
}

func (t *Theme) Box() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(t.TertiaryColor).
		Padding(0, 1).
		MarginLeft(2)
}

func (t *Theme) RenderBold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}

func (t *Theme) RenderInfo(text string) string {
	return lipgloss.NewStyle().Foreground(t.InfoColor).Render(text)
}

func (t *Theme) RenderSuccess(text string) string {
	return lipgloss.NewStyle().Foreground(t.SecondaryColor).Render(text)
}

func (t *Theme) RenderWarning(text string) string {
	return lipgloss.NewStyle().Foreground(t.WarningColor).Render(text)
}

func (t *Theme) RenderError(text string) string {
	return lipgloss.NewStyle().Foreground(t.ErrorColor).Render(text)
}

func (t *Theme) RenderUnknown(text string) string {
	return lipgloss.NewStyle().Foreground(t.Gray).Render(text)
}

func (t *Theme) RenderBrand(text string) string {
	return lipgloss.NewStyle().
		PaddingRight(2).
		PaddingLeft(4).
		Italic(true).
		Bold(true).
		Foreground(t.Black).
		Background(t.SecondaryColor).Render(text)
}

func (t *Theme) RenderContext(label, value string) string {
	labelBlock := lipgloss.NewStyle().
		PaddingRight(0).
		PaddingLeft(2).
		Foreground(t.Black).
		Bold(true).
		Background(t.TertiaryColor).Render(label)
	valueBlock := lipgloss.NewStyle().
		PaddingRight(10).
		PaddingLeft(1).
		Foreground(t.Black).
		Background(t.TertiaryColor).Render(value)
	return labelBlock + valueBlock
}

func (t *Theme) RenderInputForm(text string) string {
	return lipgloss.NewStyle().
		PaddingLeft(2).
		Render(text)
}

func (t *Theme) RenderHelp(text string) string {
	return lipgloss.NewStyle().
		MarginLeft(2).
		Foreground(t.Gray).Render(text)
}

func (t *Theme) RenderInContainer(text string) string {
	return lipgloss.Style{}.MarginLeft(2).Render(text)
}

func (t *Theme) ListStyles() list.Styles {
	styles := list.DefaultStyles()
	styles.StatusBar = styles.StatusBar.
		Padding(0, 0, 1, 0).
		Italic(true).
		Foreground(t.TertiaryColor)
	return styles
}

func (t *Theme) ListItemStyles() list.DefaultItemStyles {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.
		Foreground(t.TertiaryColor).
		Bold(true)
	styles.NormalDesc = styles.NormalDesc.Foreground(t.White)

	styles.SelectedTitle = styles.SelectedTitle.
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		Foreground(t.SecondaryColor).
		BorderForeground(t.SecondaryColor).
		Bold(true)
	styles.SelectedDesc = styles.SelectedDesc.
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Foreground(t.White)
	return styles
}

func (t *Theme) LoggerStyles() *log.Styles {
	baseStyles := log.DefaultStyles()
	baseStyles.Timestamp = baseStyles.Timestamp.Foreground(lipgloss.AdaptiveColor{Dark: "#505050", Light: "#505050"})
	baseStyles.Key = baseStyles.Key.Foreground(t.SecondaryColor)
	baseStyles.Value = baseStyles.Value.Foreground(t.Gray)
	baseStyles.Levels = map[log.Level]lipgloss.Style{
		log.InfoLevel:  baseStyles.Levels[log.InfoLevel].Foreground(t.InfoColor),
		log.WarnLevel:  baseStyles.Levels[log.WarnLevel].Foreground(t.WarningColor),
		log.ErrorLevel: baseStyles.Levels[log.ErrorLevel].Foreground(t.ErrorColor).SetString("ERR"),
		log.DebugLevel: baseStyles.Levels[log.DebugLevel].Foreground(t.TertiaryColor),
		log.FatalLevel: baseStyles.Levels[log.FatalLevel].Foreground(t.SecondaryColor).SetString("ERR"),
	}
	return baseStyles
}
