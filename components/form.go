package components

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/jahvon/tuikit/styles"
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
	err    *ErrorView
	done   chan struct{}
}

//nolint:funlen,gocognit
func NewForm(
	styles styles.Theme,
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
		done:   make(chan struct{}),
	}

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
	var addColumn bool
	for _, g := range groups {
		var hf []huh.Field
		for _, field := range g {
			height := strings.Count(field.Description, "\n") + strings.Count(field.Title, "\n")
			switch field.Type {
			case PromptTypeText, PromptTypeMasked:
				mode := huh.EchoModeNormal
				if field.Type == PromptTypeMasked {
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
				if field.ValidationExpr != "" {
					txt = txt.Validate(field.ValidateValue)
				}
				hf = append(hf, txt.WithHeight(height))
			case PromptTypeMultiline:
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
				hf = append(hf, txt.WithHeight(height))
			case PromptTypeConfirm:
				txt := huh.NewConfirm().Key(field.Key).Value(&field.confirmed)
				if field.Title != "" {
					txt = txt.Title(field.Title)
				}
				if field.Description != "" {
					txt = txt.Description(field.Description)
				}
				hf = append(hf, txt.WithHeight(height))
			default:
				return nil, fmt.Errorf("unknown field type: %v", field.Type)
			}
		}
		if len(hf) > 0 {
			hg = append(hg, huh.NewGroup(hf...))
			addColumn = addColumn || len(hf) > 3
		}
	}
	accessibleMode := os.Getenv("TUI_ACCESSIBLE") != ""
	hf := huh.NewForm(hg...).
		WithProgramOptions(tea.WithInput(in), tea.WithOutput(out)).
		WithTheme(styles.FormStyles()).
		WithAccessible(accessibleMode)
	hf.SubmitCmd = SubmitMsg
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

func (f *Form) ShowFooter() bool {
	return false
}

func (f *Form) Type() string {
	return FormViewType
}
