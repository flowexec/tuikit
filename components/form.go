package components

import (
	"fmt"
	"io"
	"regexp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/jahvon/tuikit/styles"
)

type FormFieldType uint

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
	Prompt         string
	Description    string
	Placeholder    string

	value interface{}
}

func (f *FormField) Set(val string) {
	f.value = val
}

func (f *FormField) SetAndValidate(val string) error {
	f.Set(val)
	return f.ValidateValue(val)
}

func (f *FormField) ValidateConfig() error {
	if f.Key == "" {
		return fmt.Errorf("field is missing a key")
	}
	if f.Title == "" && f.Prompt == "" && f.Description == "" {
		return fmt.Errorf("field %s must specify at least one of title, prompt, or description", f.Key)
	}
	return nil
}

func (f *FormField) Value() string {
	if f.value == nil {
		return f.Default
	}
	return fmt.Sprintf("%v", f.value)
}

func (f *FormField) ValidateValue(val string) error {
	if val == "" && f.Required {
		return fmt.Errorf("required field with key %s not set", f.Key)
	}

	if f.ValidationExpr != "" {
		r, err := regexp.Compile(f.ValidationExpr)
		if err != nil {
			return fmt.Errorf("unable to compile validation regex for field with key %s: %w", f.Key, err)
		}
		if !r.MatchString(fmt.Sprintf("%v", f.Value())) {
			return fmt.Errorf("validation (%s) failed for field with key %s", f.ValidationExpr, f.Key)
		}
	}
	return nil
}

type Form struct {
	fields []*FormField
	form   *huh.Form
	styles styles.Theme
	err    TeaModel
}

func NewForm(
	styles styles.Theme,
	accessible bool,
	in io.Reader,
	out io.Writer,
	fields ...*FormField,
) (*Form, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("no fields provided")
	}
	form := &Form{
		fields: fields,
		styles: styles,
	}

	groups := make(map[uint][]*FormField)
	for _, f := range fields {
		if groups[f.Group] == nil {
			groups[f.Group] = []*FormField{}
		}
		groups[f.Group] = append(groups[f.Group], f)
	}
	var hg []*huh.Group
	var addColumn bool
	for _, g := range groups {
		var hf []huh.Field
		for _, field := range g {
			switch field.Type {
			case PromptTypeText, PromptTypeMasked:
				mode := huh.EchoModeNormal
				if field.Type == PromptTypeMasked {
					mode = huh.EchoModePassword
				}
				var v string
				field.value = &v
				txt := huh.NewInput().
					Title(field.Title).
					Prompt(field.Prompt).
					Description(field.Description).
					Placeholder(field.Placeholder).
					EchoMode(mode).
					Key(field.Key).
					Validate(field.ValidateValue).
					Value(&v)
				hf = append(hf, txt)
			case PromptTypeMultiline:
				var v string
				field.value = &v
				txt := huh.NewText().
					Title(field.Title).
					Placeholder(field.Placeholder).
					Description(field.Description).
					Key(field.Key).
					Validate(field.ValidateValue).
					Value(&v)
				hf = append(hf, txt)
			case PromptTypeConfirm:
				var v bool
				field.value = &v
				txt := huh.NewConfirm().
					Title(field.Title).
					Description(field.Description).
					Key(field.Key).
					Value(&v)
				hf = append(hf, txt)
			default:
				return nil, fmt.Errorf("unknown field type: %v", field.Type)
			}
		}
		if len(hf) > 0 {
			hg = append(hg, huh.NewGroup(hf...))
			addColumn = addColumn || len(hf) > 3
		}
	}
	hf := huh.NewForm(hg...).
		WithProgramOptions(tea.WithInput(in), tea.WithOutput(out)).
		WithTheme(styles.FormStyles()).
		WithAccessible(accessible)
	hf.SubmitCmd = tea.Quit
	hf.CancelCmd = tea.Quit
	if addColumn {
		hf = hf.WithLayout(huh.LayoutColumns(2))
	}
	form.form = hf
	return form, nil
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

func (f *Form) RunProgram() error {
	return f.form.Run()
}

func (f *Form) Init() tea.Cmd {
	return f.form.Init()
}

func (f *Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if f.err != nil {
		return f.err.Update(msg)
	}

	//nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		//nolint:exhaustive
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return f.form, tea.Quit
		}
	}

	model, cmd := f.form.Update(msg)
	var ok bool
	f.form, ok = model.(*huh.Form)
	if !ok {
		f.err = NewErrorView(fmt.Errorf("unable to cast form model to huh.Form"), f.styles)
		return f, cmd
	}
	return f.form, cmd
}

func (f *Form) View() string {
	if f.err != nil {
		return f.err.View()
	}

	return f.form.View()
}

func (f *Form) HelpMsg() string {
	return ""
}

func (f *Form) Interactive() bool {
	return true
}

func (f *Form) Type() string {
	return "form"
}
