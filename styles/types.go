package styles

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type OutputLevel string

const (
	OutputLevelSuccess OutputLevel = "success"
	OutputLevelNotice  OutputLevel = "notice"
	OutputLevelInfo    OutputLevel = "info"
	OutputLevelWarning OutputLevel = "warning"
	OutputLevelError   OutputLevel = "error"

	// LogNoticeLevel is a custom log level for notices.
	LogNoticeLevel = log.InfoLevel + 1
)

type Theme struct {
	Name string
	// see https://github.com/alecthomas/chroma
	ChromaCodeStyle string

	SpinnerType spinner.Spinner

	BodyColor      lipgloss.Color
	EmphasisColor  lipgloss.Color
	BorderColor    lipgloss.Color
	PrimaryColor   lipgloss.Color
	SecondaryColor lipgloss.Color
	TertiaryColor  lipgloss.Color
	SuccessColor   lipgloss.Color
	WarningColor   lipgloss.Color
	ErrorColor     lipgloss.Color
	InfoColor      lipgloss.Color
	White          lipgloss.Color
	Gray           lipgloss.Color
	Black          lipgloss.Color
}

type templateData struct {
	BodyColor         string
	TitleColor        string
	HeadingColor      string
	SmallHeadingColor string
	DividerColor      string
	LinkColor         string
	QuoteColor        string
	ItemColor         string
	EmphasisColor     string
	CodeTextColor     string
	CodeBgColor       string
	DarkFgColor       string

	ChromaTheme string
}

func (t Theme) markdownTemplateData() templateData {
	return templateData{
		BodyColor:     string(t.BodyColor),
		TitleColor:    string(t.EmphasisColor),
		HeadingColor:  string(t.PrimaryColor),
		LinkColor:     string(t.TertiaryColor),
		QuoteColor:    string(t.Gray),
		ItemColor:     string(t.BodyColor),
		EmphasisColor: string(t.EmphasisColor),
		DividerColor:  string(t.BodyColor),
		CodeTextColor: string(t.White),
		CodeBgColor:   string(t.Gray),
		DarkFgColor:   string(t.Black),
		ChromaTheme:   t.ChromaCodeStyle,
	}
}
