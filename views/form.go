package views

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
)

type FormFieldType uint

const FormViewType = "form"
const (
	PromptTypeText FormFieldType = iota
	PromptTypeMasked
	PromptTypeMultiline
	PromptTypeConfirm
	// TODO: implement select/multi-select prompt types
)

type FormField struct {
	Group uint
	Type  FormFieldType
	Key   string

	Default        string
	Required       bool
	ValidationExpr string
	Title          string
	Description    string
	Placeholder    string

	value     string
	confirmed bool
}

func (f *FormField) Set(val string) {
	//nolint:exhaustive
	switch f.Type {
	case PromptTypeConfirm:
		if v := strings.ToLower(val); v == "y" || v == "yes" {
			f.confirmed = true
			return
		}
		f.confirmed, _ = strconv.ParseBool(val)
	default:
		f.value = val
	}
}

func (f *FormField) SetAndValidate(val string) error {
	f.Set(val)
	return f.ValidateValue(val)
}

func (f *FormField) ValidateConfig() error {
	if f.Key == "" {
		return fmt.Errorf("field is missing a key")
	}
	if f.Title == "" && f.Description == "" {
		return fmt.Errorf("field %s must specify at least a title or description", f.Key)
	}
	return nil
}

func (f *FormField) Value() string {
	//nolint:exhaustive
	switch f.Type {
	case PromptTypeConfirm:
		if f.Default != "" {
			d, _ := strconv.ParseBool(f.Default)
			return fmt.Sprintf("%v", f.confirmed || d)
		}
		return fmt.Sprintf("%v", f.confirmed)
	default:
		if f.value == "" {
			return f.Default
		}
		return f.value
	}
}

func (f *FormField) ValidateValue(val string) error {
	if val == "" && f.Required {
		return fmt.Errorf("required field with key '%s' not set", f.Key)
	}

	if f.ValidationExpr != "" {
		r, err := regexp.Compile(f.ValidationExpr)
		if err != nil {
			return fmt.Errorf("unable to compile validation regex for field with key '%s': %w", f.Key, err)
		}
		if !r.MatchString(fmt.Sprintf("%v", f.Value())) {
			return fmt.Errorf("validation (%s) failed for field with key '%s'", f.ValidationExpr, f.Key)
		}
	}
	return nil
}

type Form struct {
	Callback func(map[string]any) error

	fields    []*FormField
	form      *huh.Form
	theme     themes.Theme
	err       *ErrorView
	completed bool

	in, out *os.File
	mu      sync.RWMutex
}

// NewForm creates a new form model that can be run in a Bubble Tea program. It includes some extra handling
// for reading piped input.
// This form should NOT be used in a tuikit container. Instead, use NewFormView.
func NewForm(
	theme themes.Theme,
	in *os.File,
	out *os.File,
	fields ...*FormField,
) (*Form, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields provided")
	}
	form := &Form{
		fields: fields,
		theme:  theme,
		in:     in,
		out:    out,
	}

	programOpts := make([]tea.ProgramOption, 0)
	if in != nil {
		inInfo, err := in.Stat()
		if err != nil {
			return nil, fmt.Errorf("unable to get input file info: %w", err)
		}

		if inInfo.Mode()&os.ModeNamedPipe != 0 || !term.IsTerminal(int(in.Fd())) {
			if err := readPipedInput(in, fields); err != nil {
				return nil, fmt.Errorf("error reading piped input: %w", err)
			}
			form.completed = true
			return form, nil
		}
		programOpts = append(programOpts, tea.WithInput(in))
	}
	if out != nil {
		programOpts = append(programOpts, tea.WithOutput(out))
	}

	hg, err := fieldsToHuhGroups(fields)
	if err != nil {
		return nil, fmt.Errorf("unable to convert form groups: %w", err)
	}

	hf := huh.NewForm(hg...).
		WithProgramOptions(programOpts...).
		WithTheme(theme.HuhTheme()).
		WithAccessible(accessibleMode())
	if len(fields) > 5 {
		hf = hf.WithLayout(huh.LayoutColumns(2)) // TODO: make this configurable or auto-dynamic
	}
	form.form = hf
	return form, nil
}

// NewFormView creates a new form view that can be used in the tuikit container.
func NewFormView(
	state *types.RenderState,
	fields ...*FormField,
) (*Form, error) {
	hg, err := fieldsToHuhGroups(fields)
	if err != nil {
		return nil, fmt.Errorf("unable to convert form groups: %w", err)
	}
	hf := huh.NewForm(hg...).
		WithTheme(state.Theme.HuhTheme()).
		WithAccessible(accessibleMode()).
		WithWidth(state.ContentWidth).
		WithHeight(state.ContentHeight).
		WithShowHelp(true)
	hf.SubmitCmd = types.Submit
	hf.CancelCmd = tea.Quit
	if len(fields) > 5 {
		hf = hf.WithLayout(huh.LayoutColumns(2)) // TODO: make this configurable or auto-dynamic
	}
	return &Form{
		fields: fields,
		theme:  state.Theme,
		form:   hf,
	}, nil
}

func (f *Form) FindByKey(key string) *FormField {
	if f == nil {
		return nil
	}
	for _, field := range f.fields {
		if field.Key == key {
			return field
		}
	}
	return nil
}

func (f *Form) ValueMap() map[string]any {
	m := make(map[string]any)
	for _, field := range f.fields {
		m[field.Key] = field.Value()
	}
	return m
}

func (f *Form) Completed() bool {
	return f.completed
}

func (f *Form) Init() tea.Cmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.form.Init()
}

func (f *Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.err != nil {
		return f.err.Update(msg)
	}

	//nolint:gocritic
	switch msg.(type) {
	case types.SubmitMsg:
		f.completed = true
		if f.Callback != nil {
			if err := f.Callback(f.ValueMap()); err != nil {
				f.err = NewErrorView(err, f.theme)
				return f, nil
			}
		}
		return f, types.ReplaceView
	}

	model, cmd := f.form.Update(msg)
	var ok bool
	f.form, ok = model.(*huh.Form)
	if !ok {
		f.err = NewErrorView(fmt.Errorf("unable to cast form model to huh.Form"), f.theme)
		return f, cmd
	}
	return f, cmd
}

func (f *Form) View() string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.err != nil {
		return f.err.View()
	}

	return f.form.View()
}

func (f *Form) HelpMsg() string {
	return ""
}

func (f *Form) ShowFooter() bool {
	return false
}

func (f *Form) Type() string {
	return FormViewType
}

func (f *Form) Run(ctx context.Context) error {
	if f.completed {
		printFieldsSummary(f.out, f.fields, f.theme)
		return nil
	}
	return f.form.RunWithContext(ctx)
}

func fieldsToHuhGroups(fields []*FormField) ([]*huh.Group, error) {
	groups := make(map[uint][]*FormField)
	for _, f := range fields {
		if err := f.ValidateConfig(); err != nil {
			return nil, fmt.Errorf("invalid field config: %w", err)
		}
		if groups[f.Group] == nil {
			groups[f.Group] = []*FormField{}
		}
		groups[f.Group] = append(groups[f.Group], f)
	}
	var hg []*huh.Group
	for _, g := range groups {
		var hf []huh.Field
		for _, field := range g {
			height := strings.Count(field.Description, "\n") + strings.Count(field.Title, "\n")
			switch field.Type {
			case PromptTypeText, PromptTypeMasked:
				masked := field.Type == PromptTypeMasked
				hf = append(hf, textHuhField(field, masked).WithHeight(height))
			case PromptTypeMultiline:
				hf = append(hf, multilineHuhField(field).WithHeight(height))
			case PromptTypeConfirm:
				hf = append(hf, confirmHuhField(field).WithHeight(height))
			default:
				return nil, fmt.Errorf("unknown field type: %v", field.Type)
			}
		}
		if len(hf) > 0 {
			hg = append(hg, huh.NewGroup(hf...))
		}
	}
	return hg, nil
}

func readPipedInput(in *os.File, fields []*FormField) error {
	reader := bufio.NewReader(in)
	for _, field := range fields {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("error reading input line: %w", err)
		} else if line == "" && errors.Is(err, io.EOF) {
			return fmt.Errorf("not enough input lines")
		}
		if !field.Required && line == "" && field.Default != "" {
			line = field.Default
		}
		if err := field.SetAndValidate(strings.TrimSpace(line)); err != nil {
			return fmt.Errorf("error setting field value: %w", err)
		}
	}
	return nil
}

func printFieldsSummary(out *os.File, fields []*FormField, styles themes.Theme) {
	if out == nil {
		return
	}
	groupedFields := make(map[uint][]*FormField)
	for _, field := range fields {
		if groupedFields[field.Group] == nil {
			groupedFields[field.Group] = []*FormField{}
		}
		groupedFields[field.Group] = append(groupedFields[field.Group], field)
	}
	for _, group := range groupedFields {
		for _, field := range group {
			_, _ = fmt.Fprintln(out, styles.RenderKeyAndValueWithBreak(field.Title, field.Value()))
		}
		_, _ = fmt.Fprintln(out)
	}
}

func textHuhField(field *FormField, masked bool) huh.Field {
	mode := huh.EchoModeNormal
	if masked {
		mode = huh.EchoModePassword
	}
	txt := huh.NewInput().EchoMode(mode).Prompt("> ").Key(field.Key).Value(&field.value)
	if field.Title != "" {
		txt = txt.Title(field.Title)
	}
	if field.Description != "" {
		txt = txt.Description(field.Description)
	}
	if field.Placeholder != "" {
		txt = txt.Placeholder(field.Placeholder)
	} else if field.Default != "" {
		txt = txt.Placeholder(field.Default)
	}
	if field.ValidationExpr != "" || field.Required {
		txt = txt.Validate(field.ValidateValue)
	}
	return txt
}

func multilineHuhField(field *FormField) huh.Field {
	txt := huh.NewText().Key(field.Key).Value(&field.value)
	if field.Title != "" {
		txt = txt.Title(field.Title)
	}
	if field.Placeholder != "" {
		txt = txt.Placeholder(field.Placeholder)
	} else if field.Default != "" {
		txt = txt.Placeholder(field.Default)
	}
	if field.Description != "" {
		txt = txt.Description(field.Description)
	}
	if field.ValidationExpr != "" {
		txt = txt.Validate(field.ValidateValue)
	}
	return txt
}

func confirmHuhField(field *FormField) huh.Field {
	txt := huh.NewConfirm().Key(field.Key).Value(&field.confirmed)
	if field.Title != "" {
		txt = txt.Title(field.Title)
	}
	if field.Description != "" {
		txt = txt.Description(field.Description)
	}
	return txt
}

func accessibleMode() bool {
	return os.Getenv("TUI_ACCESSIBLE") != ""
}
