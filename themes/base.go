package themes

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"charm.land/log/v2"
)

const (
	HeaderHeight = 2
)

//go:embed mdstyles.tmpl.json
var mdstylesTemplateJSON string

type baseTheme struct {
	Name        string          `json:"-" yaml:"-"`
	SpinnerType spinner.Spinner `json:"-" yaml:"-"`
	Colors      *ColorPalette
	isDark      bool
}

func (t baseTheme) String() string {
	return t.Name
}

func (t baseTheme) ColorPalette() *ColorPalette {
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
		BorderStyle(lipgloss.RoundedBorder()).
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

func (t baseTheme) RenderHeader(appName, version, stateKey, stateVal string, width int) string {
	if width == 0 {
		return t.renderShortHeader(appName, version, stateKey, stateVal)
	}

	pad := 1 // left/right padding

	appNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Black)).
		Background(lipgloss.Color(t.Colors.AppName)).
		Bold(true).
		Padding(0, 1)
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Border))

	left := strings.Repeat(" ", pad) + appNameStyle.Render(appName)
	if stateKey != "" && stateVal != "" {
		ctxStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Colors.Secondary)).
			Italic(true)
		left += sepStyle.Render(" · ") +
			ctxStyle.Render(fmt.Sprintf("%s:", stateKey)) +
			ctxStyle.Render(stateVal)
	}

	// Right side: optional version + help hint.
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Gray))
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Tertiary))
	var right string
	if version != "" {
		right = versionStyle.Render(version) + sepStyle.Render(" · ")
	}
	right += hintStyle.Render("? help") + strings.Repeat(" ", pad)

	// Fill the gap with faint dots.
	gapLen := width - lipgloss.Width(left) - lipgloss.Width(right) - 2 // 2 for spaces around dots
	gapLen = max(gapLen, 1)
	dotStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Colors.Border)).Faint(true)
	border := " " + dotStyle.Render(strings.Repeat("·", gapLen)) + " "

	return left + border + right + "\n"
}

func (t baseTheme) renderShortHeader(appName, version, ctxKey, ctxVal string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Primary)).
		Bold(true)
	parts := appName
	if version != "" {
		parts += " " + version
	}
	if ctxKey != "" {
		parts += fmt.Sprintf(" %s:%s", ctxKey, ctxVal)
	}
	return headerStyle.Render(parts)
}

func (t baseTheme) RenderHelpPopup(keys []HelpKey, width, height int) string {
	if len(keys) == 0 {
		return ""
	}

	// Find the longest key and description for sizing.
	maxKeyLen := 0
	maxDescLen := 0
	for _, k := range keys {
		maxKeyLen = max(maxKeyLen, len(k.Key))
		maxDescLen = max(maxDescLen, len(k.Desc))
	}

	// Content width = key column + gap + desc column.
	contentW := maxKeyLen + 2 + maxDescLen
	// Add chrome: padding (3+3 horizontal).
	boxW := contentW + 6
	// Clamp to reasonable bounds.
	boxW = max(boxW, 20)
	boxW = min(boxW, width*8/10)
	innerW := boxW - 6

	bgColor := lipgloss.Color(t.Colors.Black)
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Secondary)).
		Background(bgColor).
		Bold(true).
		Width(maxKeyLen).
		Align(lipgloss.Right)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Gray)).
		Background(bgColor)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Primary)).
		Background(bgColor).
		Bold(true).
		Width(innerW).
		Align(lipgloss.Center)

	lines := make([]string, 0, len(keys)+2)
	lines = append(lines, titleStyle.Render("Help"))
	lines = append(lines, "")
	for _, k := range keys {
		line := lipgloss.JoinHorizontal(lipgloss.Top,
			keyStyle.Render(k.Key),
			descStyle.Render("  "+k.Desc),
		)
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	boxStyle := lipgloss.NewStyle().
		Background(bgColor).
		Padding(1, 3).
		Width(boxW).
		MaxHeight(height * 7 / 10)

	return boxStyle.Render(content)
}

func (t baseTheme) RenderToast(text string, lvl OutputLevel, width int) string {
	maxW := min(40, width-2)

	var accentColor string
	switch lvl {
	case OutputLevelSuccess:
		accentColor = t.Colors.Success
	case OutputLevelWarning:
		accentColor = t.Colors.Warning
	case OutputLevelError:
		accentColor = t.Colors.Error
	case OutputLevelNotice:
		accentColor = t.Colors.Emphasis
	case OutputLevelInfo:
		accentColor = t.Colors.Info
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(accentColor)).
		Padding(0, 1).
		MaxWidth(maxW)

	styledText := t.RenderLevel(text, lvl)
	return boxStyle.Render(styledText)
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
	s := list.DefaultStyles(t.isDark)
	s.StatusBar = s.StatusBar.
		Padding(0, 0, 1, 0).
		Italic(true).
		Foreground(lipgloss.Color(t.Colors.Tertiary))
	return s
}

func (t baseTheme) ListItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles(t.isDark)
	s.NormalTitle = s.NormalTitle.Foreground(lipgloss.Color(t.Colors.Secondary))
	s.NormalDesc = s.NormalDesc.Foreground(lipgloss.Color(t.Colors.Body))

	s.SelectedTitle = s.SelectedTitle.
		Border(lipgloss.DoubleBorder(), false, false, false, true).
		Foreground(lipgloss.Color(t.Colors.Primary)).
		BorderForeground(lipgloss.Color(t.Colors.Primary)).
		Bold(true)
	s.SelectedDesc = s.SelectedDesc.
		Border(lipgloss.HiddenBorder(), false, false, false, true).
		Foreground(lipgloss.Color(t.Colors.Gray))
	return s
}

func (t baseTheme) LoggerStyles() *log.Styles {
	baseStyles := log.DefaultStyles()
	baseStyles.Timestamp = baseStyles.Timestamp.Foreground(lipgloss.Color(t.Colors.Gray))
	baseStyles.Levels = map[log.Level]lipgloss.Style{
		log.InfoLevel:  baseStyles.Levels[log.InfoLevel].Foreground(lipgloss.Color(t.Colors.Info)).SetString("INF"),
		LogNoticeLevel: baseStyles.Levels[LogNoticeLevel].Foreground(lipgloss.Color(t.Colors.Warning)).SetString("NTC"),
		log.WarnLevel:  baseStyles.Levels[log.WarnLevel].Foreground(lipgloss.Color(t.Colors.Warning)).SetString("WRN"),
		log.ErrorLevel: baseStyles.Levels[log.ErrorLevel].Foreground(lipgloss.Color(t.Colors.Error)).SetString("ERR"),
		log.DebugLevel: baseStyles.Levels[log.DebugLevel].Foreground(lipgloss.Color(t.Colors.Emphasis)).SetString("DBG"),
		log.FatalLevel: baseStyles.Levels[log.FatalLevel].Foreground(lipgloss.Color(t.Colors.Error)).SetString("ERR"),
	}
	baseStyles.Message = baseStyles.Message.Foreground(lipgloss.Color(t.Colors.Body))
	baseStyles.Key = baseStyles.Key.Foreground(lipgloss.Color(t.Colors.Secondary))
	baseStyles.Value = baseStyles.Value.Foreground(lipgloss.Color(t.Colors.Gray))

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

func (t baseTheme) HuhTheme() huh.Theme {
	return huh.ThemeFunc(t.huhStyles)
}

func (t baseTheme) huhStyles(isDark bool) *huh.Styles {
	baseTheme := huh.ThemeBase(isDark)
	baseTheme.FieldSeparator = lipgloss.NewStyle().SetString("\n\n")

	baseTheme.Focused.Base = baseTheme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	baseTheme.Focused.Title = baseTheme.Focused.Title.Foreground(lipgloss.Color(t.Colors.Primary)).Bold(true)
	baseTheme.Focused.Description = baseTheme.Focused.Description.Foreground(lipgloss.Color(t.Colors.Body))
	baseTheme.Focused.ErrorMessage = baseTheme.Focused.ErrorMessage.Foreground(lipgloss.Color(t.Colors.Error))
	baseTheme.Focused.FocusedButton = baseTheme.Focused.FocusedButton.Foreground(lipgloss.Color(t.Colors.Primary)).
		Background(lipgloss.Color(t.Colors.Tertiary))
	baseTheme.Focused.BlurredButton = baseTheme.Focused.BlurredButton.Foreground(lipgloss.Color(t.Colors.Secondary)).
		Background(lipgloss.Color(t.Colors.Gray))

	baseTheme.Focused.TextInput.Placeholder =
		baseTheme.Focused.TextInput.Placeholder.Foreground(lipgloss.Color(t.Colors.Body))
	baseTheme.Focused.TextInput.Cursor =
		baseTheme.Focused.TextInput.Cursor.Foreground(lipgloss.Color(t.Colors.Secondary))
	baseTheme.Focused.TextInput.CursorText =
		baseTheme.Focused.TextInput.Cursor.Foreground(lipgloss.Color(t.Colors.Secondary))
	baseTheme.Focused.TextInput.Placeholder =
		baseTheme.Focused.TextInput.Placeholder.Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Focused.TextInput.Text =
		baseTheme.Focused.TextInput.Text.Foreground(lipgloss.Color(t.Colors.Body))
	baseTheme.Focused.TextInput.Prompt =
		baseTheme.Focused.TextInput.Prompt.Foreground(lipgloss.Color(t.Colors.Tertiary))

	baseTheme.Blurred.Title = baseTheme.Blurred.Title.Foreground(lipgloss.Color(t.Colors.White))
	baseTheme.Blurred.Description = baseTheme.Blurred.Description.Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.ErrorMessage = baseTheme.Blurred.ErrorMessage.Foreground(lipgloss.Color(t.Colors.Emphasis))
	baseTheme.Blurred.FocusedButton = baseTheme.Blurred.FocusedButton.Foreground(lipgloss.Color(t.Colors.Secondary)).
		Background(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.BlurredButton = baseTheme.Blurred.BlurredButton.Foreground(lipgloss.Color(t.Colors.Gray)).
		Background(lipgloss.Color(t.Colors.Gray))

	baseTheme.Blurred.TextInput.Placeholder = baseTheme.Blurred.TextInput.Placeholder.
		Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Cursor = baseTheme.Blurred.TextInput.Cursor.
		Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.CursorText = baseTheme.Blurred.TextInput.CursorText.
		Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Placeholder = baseTheme.Blurred.TextInput.Placeholder.
		Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Text = baseTheme.Blurred.TextInput.Text.
		Foreground(lipgloss.Color(t.Colors.Gray))
	baseTheme.Blurred.TextInput.Prompt = baseTheme.Blurred.TextInput.Prompt.
		Foreground(lipgloss.Color(t.Colors.Gray))

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
