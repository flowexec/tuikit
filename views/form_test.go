package views_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/views"
)

// runPipedForm runs a form with input piped on stdin and returns the form and the summary it printed.
func runPipedForm(t *testing.T, input string, fields ...*views.FormField) (*views.Form, string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	defer r.Close()

	out, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	form, err := views.NewForm(themes.EverforestTheme(), r, out, fields...)
	if err != nil {
		return nil, "", err
	}
	if err := form.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	printed, err := os.ReadFile(out.Name())
	if err != nil {
		t.Fatal(err)
	}
	return form, string(printed), nil
}

func TestPipedMaskedValueNotPrinted(t *testing.T) {
	secret := "s3cr3t-value"
	form, printed, err := runPipedForm(t, secret+"\nalice\n",
		&views.FormField{Key: "secret", Title: "Secret", Type: views.PromptTypeMasked},
		&views.FormField{Key: "user", Title: "User", Type: views.PromptTypeText},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.FindByKey("secret").Value(); got != secret {
		t.Errorf("secret = %q, want %q", got, secret)
	}
	if strings.Contains(printed, secret) {
		t.Errorf("summary leaked masked value: %q", printed)
	}
	if !strings.Contains(printed, "alice") {
		t.Errorf("summary missing unmasked value: %q", printed)
	}
}

func TestPipedSingleFieldKeepsMultilineValue(t *testing.T) {
	input := "-----BEGIN KEY-----\nabc\n-----END KEY-----\n"
	form, printed, err := runPipedForm(t, input,
		&views.FormField{Key: "secret", Title: "Secret", Type: views.PromptTypeMasked, Required: true},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.FindByKey("secret").Value(); got != input {
		t.Errorf("value = %q, want %q", got, input)
	}
	if strings.Contains(printed, "abc") {
		t.Errorf("summary leaked masked value: %q", printed)
	}
}

func TestPipedSingleLineIsTrimmed(t *testing.T) {
	form, _, err := runPipedForm(t, "  value\n",
		&views.FormField{Key: "v", Title: "V", Type: views.PromptTypeText},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.FindByKey("v").Value(); got != "value" {
		t.Errorf("value = %q, want %q", got, "value")
	}
}

func TestPipedMultilineFieldTakesRemainingInput(t *testing.T) {
	form, _, err := runPipedForm(t, "title\nline1\nline2\n",
		&views.FormField{Key: "title", Title: "Title", Type: views.PromptTypeText},
		&views.FormField{Key: "body", Title: "Body", Type: views.PromptTypeMultiline},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.FindByKey("title").Value(); got != "title" {
		t.Errorf("title = %q", got)
	}
	if got := form.FindByKey("body").Value(); got != "line1\nline2\n" {
		t.Errorf("body = %q", got)
	}
}

func TestPipedEmptyInputFailsRequiredField(t *testing.T) {
	_, _, err := runPipedForm(t, "",
		&views.FormField{Key: "secret", Title: "Secret", Type: views.PromptTypeMasked, Required: true},
	)
	if err == nil {
		t.Fatal("expected error for empty input on required field")
	}
}

func TestPipedShortInputFailsLaterRequiredField(t *testing.T) {
	_, _, err := runPipedForm(t, "first\n",
		&views.FormField{Key: "a", Title: "A", Type: views.PromptTypeText},
		&views.FormField{Key: "b", Title: "B", Type: views.PromptTypeText, Required: true},
	)
	if err == nil {
		t.Fatal("expected error when input ends before a required field")
	}
}

func TestPipedShortInputUsesDefault(t *testing.T) {
	form, _, err := runPipedForm(t, "first\n",
		&views.FormField{Key: "a", Title: "A", Type: views.PromptTypeText},
		&views.FormField{Key: "b", Title: "B", Type: views.PromptTypeText, Default: "fallback"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := form.FindByKey("b").Value(); got != "fallback" {
		t.Errorf("b = %q, want fallback", got)
	}
}

// A single-line field must not wait for EOF, so a stdin that stays open (`yes |`, a parent process holding the
// pipe) does not hang the prompt.
func TestPipedSingleLineFieldDoesNotWaitForEOF(t *testing.T) {
	for _, typ := range []views.FormFieldType{views.PromptTypeConfirm, views.PromptTypeText} {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.WriteString("yes\n"); err != nil {
			t.Fatal(err)
		}

		done := make(chan *views.Form, 1)
		go func() {
			form, _ := views.NewForm(themes.EverforestTheme(), r, nil,
				&views.FormField{Key: "v", Title: "V", Type: typ})
			done <- form
		}()
		select {
		case form := <-done:
			if form == nil {
				t.Fatalf("type %d: NewForm failed", typ)
			}
			want := "yes"
			if typ == views.PromptTypeConfirm {
				want = "true"
			}
			if got := form.FindByKey("v").Value(); got != want {
				t.Errorf("type %d: value = %q, want %q", typ, got, want)
			}
		case <-time.After(2 * time.Second):
			t.Errorf("type %d: NewForm blocked waiting for EOF on an open pipe", typ)
		}
		_ = w.Close()
		_ = r.Close()
	}
}
