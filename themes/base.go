package themes

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/charmbracelet/bubbles/v2/list"
	"github.com/charmbracelet/bubbles/v2/spinner"
	"github.com/charmbracelet/huh"
	lipglossv1 "github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/lipgloss/v2/compat"
	"github.com/charmbracelet/log"
)

const (
	HeaderHeight = 3
	FooterHeight = 2
)

//go:embed mdstyles.tmpl.json
var mdstylesTemplateJSON string

type baseTheme struct {
	Name        string          `json:"-" yaml:"-"`
	SpinnerType spinner.Spinner `json:"-" yaml:"-"`
	Colors      ColorPalette
}

func (t baseTheme) String() string {
	return t.Name
}

func (t baseTheme) ColorPalette() ColorPalette {
	return t.Colors
}

func (t baseTheme) Spinner() spinner.Spinner {
	return t.SpinnerType
}

func (t baseTheme) SpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Secondary))
}

func (t baseTheme) EntityViewStyle() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(2)
}

func (t baseTheme) CollectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(2).Padding(0, 1)
}

func (t baseTheme) BoxStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(t.Colors.Border)).
		Padding(0, 1).
		MarginLeft(2)
}

func (t baseTheme) RenderBold(text string) string {
	return lipgloss.NewStyle().Bold(true).Render(text)
}

func (t baseTheme) RenderInfo(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Info)).Render(text)
}

func (t baseTheme) RenderNotice(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Emphasis)).Render(text)
}

func (t baseTheme) RenderSuccess(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Success)).Render(text)
}

func (t baseTheme) RenderWarning(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Warning)).Render(text)
}

func (t baseTheme) RenderError(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Error)).Render(text)
}

func (t baseTheme) RenderEmphasis(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Emphasis)).Render(text)
}

func (t baseTheme) RenderUnknown(text string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Gray)).Render(text)
}

func (t baseTheme) RenderLevel(str string, lvl OutputLevel) string {
	if str == "" {
		return ""
	}

	switch lvl {
	case OutputLevelSuccess:
		return t.RenderSuccess(str)
	case OutputLevelNotice:
		return t.RenderNotice(str)
	case OutputLevelInfo:
		return t.RenderInfo(str)
	case OutputLevelWarning:
		return t.RenderWarning(str)
	case OutputLevelError:
		return t.RenderError(str)
	default:
		return t.RenderUnknown(str)
	}
}

func (t baseTheme) RenderHeader(appName, stateKey, stateVal string, width int) string {
	if width == 0 {
		return t.renderShortHeader(appName, stateKey, stateVal)
	}

	appNameStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Colors.Border)).
		Foreground(lipgloss.Color(t.Colors.Primary)).
		AlignVertical(lipgloss.Center).
		Height(HeaderHeight - 2). // top and bottom borders
		Bold(true)
	ctxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Secondary)).
		Italic(true).
		AlignVertical(lipgloss.Center).
		Height(HeaderHeight)

	appNameContent := appNameStyle.Render(fmt.Sprintf(" %s ", appName))
	var stateContent string
	if stateKey != "" && stateVal != "" {
		stateContent = lipgloss.JoinHorizontal(
			0,
			ctxStyle.Render(t.RenderBold(fmt.Sprintf(" %s:", stateKey))),
			ctxStyle.Render(fmt.Sprintf("%s ", stateVal)),
		)
	}
	fullContent := lipgloss.JoinHorizontal(0, appNameContent, stateContent)
	borderWidth := width - lipgloss.Width(fullContent)
	if borderWidth < 0 {
		borderWidth = 0
	}
	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Border)).
		MaxWidth(borderWidth).
		Height(HeaderHeight).AlignVertical(lipgloss.Center)
	border := borderStyle.Render(strings.Repeat("─", borderWidth))
	return lipgloss.JoinHorizontal(0, fullContent, border, "\n")
}

func (t baseTheme) renderShortHeader(appName, ctxKey, ctxVal string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Primary)).
		Italic(true).Bold(true)
	return headerStyle.Render(fmt.Sprintf("<!-- %s %s(%s) --!>", appName, ctxKey, ctxVal))
}

func (t baseTheme) RenderFooter(text string, width int) string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Gray)).
		Border(lipgloss.NormalBorder()).
		BorderTop(true).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderForeground(lipgloss.Color(t.Colors.Border)).
		Height(FooterHeight - 1). // top border
		Width(width)
	return footerStyle.Render(text)
}

func (t baseTheme) RenderKeyAndValue(key, value string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Secondary)).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Gray))
	return keyStyle.Render(key) + ": " + valueStyle.Render(value)
}

func (t baseTheme) RenderKeyAndValueWithBreak(key, value string) string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Secondary)).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Gray))
	return keyStyle.Render(key) + "\n" + valueStyle.Render(value)
}

func (t baseTheme) RenderInputForm(text string) string {
	return lipgloss.NewStyle().PaddingLeft(2).Render(text)
}

func (t baseTheme) RenderInContainer(text string) string {
	return lipgloss.Style{}.MarginLeft(2).Render(text)
}

func (t baseTheme) ListStyles() list.Styles {
	s := list.DefaultStyles(compat.HasDarkBackground)
	s.StatusBar = s.StatusBar.
		Padding(0, 0, 1, 0).
		Italic(true).
		Foreground(lipglossv1.Color(t.Colors.Tertiary))
	return s
}

func (t baseTheme) ListItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles(compat.HasDarkBackground)
	s.NormalTitle = s.NormalTitle.Foreground(lipglossv1.Color(t.Colors.Secondary))
	s.NormalDesc = s.NormalDesc.Foreground(lipglossv1.Color(t.Colors.Body))

	s.SelectedTitle = s.SelectedTitle.
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		Foreground(lipglossv1.Color(t.Colors.Primary)).
		BorderForeground(lipglossv1.Color(t.Colors.Primary)).
		Bold(true)
	s.SelectedDesc = s.SelectedDesc.
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Foreground(lipgloss.Color(t.Colors.Gray))
	return s
}

func (t baseTheme) LoggerStyles() *log.Styles {
	baseStyles := log.DefaultStyles()
	baseStyles.Timestamp = baseStyles.Timestamp.Foreground(lipglossv1.Color(t.Colors.Gray))
	baseStyles.Levels = map[log.Level]lipglossv1.Style{
		log.InfoLevel:  baseStyles.Levels[log.InfoLevel].Foreground(lipglossv1.Color(t.Colors.Info)).SetString("INF"),
		LogNoticeLevel: baseStyles.Levels[LogNoticeLevel].Foreground(lipglossv1.Color(t.Colors.Warning)).SetString("NTC"),
		log.WarnLevel:  baseStyles.Levels[log.WarnLevel].Foreground(lipglossv1.Color(t.Colors.Warning)).SetString("WRN"),
		log.ErrorLevel: baseStyles.Levels[log.ErrorLevel].Foreground(lipglossv1.Color(t.Colors.Error)).SetString("ERR"),
		log.DebugLevel: baseStyles.Levels[log.DebugLevel].Foreground(lipglossv1.Color(t.Colors.Emphasis)).SetString("DBG"),
		log.FatalLevel: baseStyles.Levels[log.FatalLevel].Foreground(lipglossv1.Color(t.Colors.Error)).SetString("ERR"),
	}
	baseStyles.Message = baseStyles.Message.Foreground(lipglossv1.Color(t.Colors.Body))
	baseStyles.Key = baseStyles.Key.Foreground(lipglossv1.Color(t.Colors.Secondary))
	baseStyles.Value = baseStyles.Value.Foreground(lipglossv1.Color(t.Colors.Gray))

	return baseStyles
}

func (t baseTheme) GlamourMarkdownStyleJSON() (string, error) {
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

func (t baseTheme) HuhTheme() *huh.Theme {
	baseTheme := huh.ThemeBase()
	baseTheme.FieldSeparator = lipglossv1.NewStyle().SetString("\n\n")

	baseTheme.Focused.Base = baseTheme.Focused.Base.BorderStyle(lipglossv1.HiddenBorder())
	baseTheme.Focused.Title = baseTheme.Focused.Title.Foreground(lipglossv1.Color(t.Colors.Primary)).Bold(true)
	baseTheme.Focused.Description = baseTheme.Focused.Description.Foreground(lipglossv1.Color(t.Colors.Body))
	baseTheme.Focused.ErrorMessage = baseTheme.Focused.ErrorMessage.Foreground(lipglossv1.Color(t.Colors.Error))
	baseTheme.Focused.FocusedButton = baseTheme.Focused.FocusedButton.Foreground(lipglossv1.Color(t.Colors.Primary)).
		Background(lipglossv1.Color(t.Colors.Tertiary))
	baseTheme.Focused.BlurredButton = baseTheme.Focused.BlurredButton.Foreground(lipglossv1.Color(t.Colors.Secondary)).
		Background(lipglossv1.Color(t.Colors.Gray))

	baseTheme.Focused.TextInput.Placeholder =
		baseTheme.Focused.TextInput.Placeholder.Foreground(lipglossv1.Color(t.Colors.Body))
	baseTheme.Focused.TextInput.Cursor =
		baseTheme.Focused.TextInput.Cursor.Foreground(lipglossv1.Color(t.Colors.Secondary))
	baseTheme.Focused.TextInput.CursorText =
		baseTheme.Focused.TextInput.Cursor.Foreground(lipglossv1.Color(t.Colors.Secondary))
	baseTheme.Focused.TextInput.Placeholder =
		baseTheme.Focused.TextInput.Placeholder.Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Focused.TextInput.Text =
		baseTheme.Focused.TextInput.Text.Foreground(lipglossv1.Color(t.Colors.Body))
	baseTheme.Focused.TextInput.Prompt =
		baseTheme.Focused.TextInput.Prompt.Foreground(lipglossv1.Color(t.Colors.Tertiary))

	baseTheme.Blurred.Title = baseTheme.Blurred.Title.Foreground(lipglossv1.Color(t.Colors.White))
	baseTheme.Blurred.Description = baseTheme.Blurred.Description.Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.ErrorMessage = baseTheme.Blurred.ErrorMessage.Foreground(lipglossv1.Color(t.Colors.Emphasis))
	baseTheme.Blurred.FocusedButton = baseTheme.Blurred.FocusedButton.Foreground(lipglossv1.Color(t.Colors.Secondary)).
		Background(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.BlurredButton = baseTheme.Blurred.BlurredButton.Foreground(lipglossv1.Color(t.Colors.Gray)).
		Background(lipglossv1.Color(t.Colors.Gray))

	baseTheme.Blurred.TextInput.Placeholder = baseTheme.Blurred.TextInput.Placeholder.
		Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Cursor = baseTheme.Blurred.TextInput.Cursor.
		Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.CursorText = baseTheme.Blurred.TextInput.CursorText.
		Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Placeholder = baseTheme.Blurred.TextInput.Placeholder.
		Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Text = baseTheme.Blurred.TextInput.Text.
		Foreground(lipglossv1.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Prompt = baseTheme.Blurred.TextInput.Prompt.
		Foreground(lipglossv1.Color(t.Colors.Gray))

	return baseTheme
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

func (t baseTheme) markdownTemplateData() templateData {
	return templateData{
		BodyColor:     t.Colors.Body,
		TitleColor:    t.Colors.Emphasis,
		HeadingColor:  t.Colors.Primary,
		LinkColor:     t.Colors.Tertiary,
		QuoteColor:    t.Colors.Gray,
		ItemColor:     t.Colors.Body,
		EmphasisColor: t.Colors.Emphasis,
		DividerColor:  t.Colors.Body,
		CodeTextColor: t.Colors.White,
		CodeBgColor:   t.Colors.Gray,
		DarkFgColor:   t.Colors.Black,
		ChromaTheme:   t.Colors.ChromaCodeStyle,
	}
}
