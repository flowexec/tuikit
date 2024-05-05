package styles

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

type NoticeLevel string

const (
	NoticeLevelSuccess NoticeLevel = "success"
	NoticeLevelInfo    NoticeLevel = "info"
	NoticeLevelWarning NoticeLevel = "warning"
	NoticeLevelError   NoticeLevel = "error"
)

type Theme struct {
	SpinnerType spinner.Spinner

	BodyColor      lipgloss.AdaptiveColor
	EmphasisColor  lipgloss.AdaptiveColor
	BorderColor    lipgloss.AdaptiveColor
	PrimaryColor   lipgloss.AdaptiveColor
	SecondaryColor lipgloss.AdaptiveColor
	TertiaryColor  lipgloss.AdaptiveColor
	SuccessColor   lipgloss.AdaptiveColor
	WarningColor   lipgloss.AdaptiveColor
	ErrorColor     lipgloss.AdaptiveColor
	InfoColor      lipgloss.AdaptiveColor
	White          lipgloss.AdaptiveColor
	Gray           lipgloss.AdaptiveColor
	Black          lipgloss.AdaptiveColor
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
		BodyColor:     t.BodyColor.Dark,
		TitleColor:    t.EmphasisColor.Dark,
		HeadingColor:  t.PrimaryColor.Dark,
		LinkColor:     t.TertiaryColor.Dark,
		QuoteColor:    t.Gray.Dark,
		ItemColor:     t.BodyColor.Dark,
		EmphasisColor: t.EmphasisColor.Dark,
		DividerColor:  t.BodyColor.Dark,
		CodeTextColor: t.White.Dark,
		CodeBgColor:   t.Gray.Light,
		DarkFgColor:   t.Black.Dark,
		ChromaTheme:   "friendly",
	}
}
