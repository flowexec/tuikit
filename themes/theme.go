package themes

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

type Theme interface {
	String() string
	ColorPalette() *ColorPalette

	RenderBold(text string) string
	RenderInfo(text string) string
	RenderNotice(text string) string
	RenderSuccess(text string) string
	RenderWarning(text string) string
	RenderError(text string) string
	RenderEmphasis(text string) string
	RenderUnknown(text string) string
	RenderLevel(str string, lvl OutputLevel) string
	RenderHeader(appName, stateKey, stateVal string, width int) string
	RenderFooter(text string, width int) string
	RenderKeyAndValue(key, value string) string
	RenderKeyAndValueWithBreak(key, value string) string
	RenderInputForm(text string) string
	RenderInContainer(text string) string

	Spinner() spinner.Spinner
	SpinnerStyle() lipgloss.Style
	EntityViewStyle() lipgloss.Style
	CollectionStyle() lipgloss.Style
	BoxStyle() lipgloss.Style
	ListStyles() list.Styles
	ListItemStyles() list.DefaultItemStyles

	LoggerStyles() *log.Styles
	GlamourMarkdownStyleJSON() (string, error)
	HuhTheme() *huh.Theme
}
