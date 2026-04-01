package themes

import (
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
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
	HuhTheme() huh.Theme
}
