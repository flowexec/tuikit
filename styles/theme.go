package styles

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

const (
	HeaderHeight = 3
	FooterHeight = 2
)

func (t Theme) Spinner() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.SecondaryColor)
}

func (t Theme) EntityView() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(2)
}

func (t Theme) Collection() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(2).Padding(0, 1)
}

func (t Theme) Box() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.BorderColor).
		Padding(0, 1).
		MarginLeft(2)
}

func (t Theme) RenderBold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}

func (t Theme) RenderInfo(text string) string {
	return lipgloss.NewStyle().Foreground(t.InfoColor).Render(text)
}

func (t Theme) RenderSuccess(text string) string {
	return lipgloss.NewStyle().Foreground(t.SuccessColor).Render(text)
}

func (t Theme) RenderWarning(text string) string {
	return lipgloss.NewStyle().Foreground(t.WarningColor).Render(text)
}

func (t Theme) RenderError(text string) string {
	return lipgloss.NewStyle().Foreground(t.ErrorColor).Render(text)
}

func (t Theme) RenderEmphasis(text string) string {
	return lipgloss.NewStyle().Foreground(t.EmphasisColor).Render(text)
}

func (t Theme) RenderUnknown(text string) string {
	return lipgloss.NewStyle().Foreground(t.Gray).Render(text)
}

func (t Theme) RenderNotice(notice string, lvl NoticeLevel) string {
	if notice == "" {
		return ""
	}

	switch lvl {
	case NoticeLevelSuccess:
		return t.RenderSuccess(notice)
	case NoticeLevelInfo:
		return t.RenderInfo(notice)
	case NoticeLevelWarning:
		return t.RenderWarning(notice)
	case NoticeLevelError:
		return t.RenderError(notice)
	default:
		return t.RenderUnknown(notice)
	}
}

func (t Theme) RenderHeader(appName, ctxKey, ctxVal string, width int) string {
	if width == 0 {
		return t.renderShortHeader(appName, ctxKey, ctxVal)
	}

	appNameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderColor).
		Foreground(t.PrimaryColor).
		AlignVertical(lipgloss.Center).
		Height(HeaderHeight - 2). // top and bottom borders
		Bold(true)
	ctxStyle := lipgloss.NewStyle().
		Foreground(t.SecondaryColor).
		Italic(true).
		AlignVertical(lipgloss.Center).
		Height(HeaderHeight)

	appNameContent := appNameStyle.Render(fmt.Sprintf(" %s ", appName))
	ctxContent := lipgloss.JoinHorizontal(
		0,
		ctxStyle.Render(t.RenderBold(fmt.Sprintf(" %s:", ctxKey))),
		ctxStyle.Render(fmt.Sprintf("%s ", ctxVal)),
	)
	fullContent := lipgloss.JoinHorizontal(0, appNameContent, ctxContent)
	borderWidth := width - lipgloss.Width(fullContent)
	if borderWidth < 0 {
		borderWidth = 0
	}
	borderStyle := lipgloss.NewStyle().
		Foreground(t.BorderColor).
		MaxWidth(borderWidth).
		Height(HeaderHeight).AlignVertical(lipgloss.Center)
	border := borderStyle.Render(strings.Repeat("─", borderWidth))
	return lipgloss.JoinHorizontal(0, fullContent, border, "\n")
}

func (t Theme) renderShortHeader(appName, ctxKey, ctxVal string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(t.PrimaryColor).
		Italic(true).Bold(true)
	return headerStyle.Render(fmt.Sprintf("<!-- %s %s(%s) --!>", appName, ctxKey, ctxVal))
}

func (t Theme) RenderFooter(text string, width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(t.Gray).
		Border(lipgloss.NormalBorder()).
		BorderTop(true).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderForeground(t.BorderColor).
		Height(FooterHeight - 1). // top border
		Width(width)
	return footerStyle.Render(text)
}

func (t Theme) RenderInputForm(text string) string {
	return lipgloss.NewStyle().PaddingLeft(2).Render(text)
}

func (t Theme) RenderInContainer(text string) string {
	return lipgloss.Style{}.MarginLeft(2).Render(text)
}

func (t Theme) ListStyles() list.Styles {
	styles := list.DefaultStyles()
	styles.StatusBar = styles.StatusBar.
		Padding(0, 0, 1, 0).
		Italic(true).
		Foreground(t.TertiaryColor)
	return styles
}

func (t Theme) ListItemStyles() list.DefaultItemStyles {
	styles := list.NewDefaultItemStyles()
	styles.NormalTitle = styles.NormalTitle.Foreground(t.SecondaryColor)
	styles.NormalDesc = styles.NormalDesc.Foreground(t.BodyColor)

	styles.SelectedTitle = styles.SelectedTitle.
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		Foreground(t.PrimaryColor).
		BorderForeground(t.PrimaryColor).
		Bold(true)
	styles.SelectedDesc = styles.SelectedDesc.
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Foreground(t.Gray)
	return styles
}

func (t Theme) LoggerStyles() *log.Styles {
	baseStyles := log.DefaultStyles()
	baseStyles.Timestamp = baseStyles.Timestamp.Foreground(t.Gray)
	baseStyles.Levels = map[log.Level]lipgloss.Style{
		log.InfoLevel:  baseStyles.Levels[log.InfoLevel].Foreground(t.InfoColor).SetString("INF"),
		log.WarnLevel:  baseStyles.Levels[log.WarnLevel].Foreground(t.WarningColor).SetString("WRN"),
		log.ErrorLevel: baseStyles.Levels[log.ErrorLevel].Foreground(t.ErrorColor).SetString("ERR"),
		log.DebugLevel: baseStyles.Levels[log.DebugLevel].Foreground(t.EmphasisColor).SetString("DBG"),
		log.FatalLevel: baseStyles.Levels[log.FatalLevel].Foreground(t.EmphasisColor).SetString("ERR"),
	}
	baseStyles.Message = baseStyles.Message.Foreground(t.BodyColor)
	baseStyles.Key = baseStyles.Key.Foreground(t.SecondaryColor)
	baseStyles.Value = baseStyles.Value.Foreground(t.Gray)

	return baseStyles
}

func (t Theme) MarkdownStyleJSON() (string, error) {
	tmpl, err := template.New("mdstyles").Parse(mdstylesTemplateJSON)
	if err != nil {
		return "", err
	}

	data := t.markdownTemplateData()
	builder := &strings.Builder{}
	err = tmpl.Execute(builder, data)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func (t Theme) FormStyles() *huh.Theme {
	theme := huh.ThemeBase()
	// TODO: add custom styles
	return theme
}
