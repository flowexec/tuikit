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
	Name        string          `json:"-" yaml:"-"`
	SpinnerType spinner.Spinner `json:"-" yaml:"-"`
	// see https://github.com/alecthomas/chroma
	ChromaCodeStyle string `json:"chromaCodeStyle" yaml:"chromaCodeStyle"`

	BodyColor      lipgloss.Color `json:"bodyColor"      yaml:"bodyColor"`
	EmphasisColor  lipgloss.Color `json:"emphasisColor"  yaml:"emphasisColor"`
	BorderColor    lipgloss.Color `json:"borderColor"    yaml:"borderColor"`
	PrimaryColor   lipgloss.Color `json:"primaryColor"   yaml:"primaryColor"`
	SecondaryColor lipgloss.Color `json:"secondaryColor" yaml:"secondaryColor"`
	TertiaryColor  lipgloss.Color `json:"tertiaryColor"  yaml:"tertiaryColor"`
	SuccessColor   lipgloss.Color `json:"successColor"   yaml:"successColor"`
	WarningColor   lipgloss.Color `json:"warningColor"   yaml:"warningColor"`
	ErrorColor     lipgloss.Color `json:"errorColor"     yaml:"errorColor"`
	InfoColor      lipgloss.Color `json:"infoColor"      yaml:"infoColor"`
	White          lipgloss.Color `json:"white"          yaml:"white"`
	Gray           lipgloss.Color `json:"gray"           yaml:"gray"`
	Black          lipgloss.Color `json:"black"          yaml:"black"`
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
