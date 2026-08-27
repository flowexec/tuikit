package io_test

import (
	"os"
	"strings"
	"testing"

	"github.com/flowexec/tuikit/io"
)

// groupOutput runs fn against a logger writing to a temp file and returns what
// was written, with the given CI-related environment applied.
func groupOutput(t *testing.T, env map[string]string, fn func(l *io.StandardLogger)) string {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}

	f, err := os.CreateTemp(t.TempDir(), "tuikit-group")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer f.Close()

	logger := io.NewLogger(io.WithMode(io.Text), io.WithOutput(f))
	fn(logger)

	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("seek: %v", err)
	}
	buf := make([]byte, 8192)
	n, _ := f.Read(buf)
	return string(buf[:n])
}

func TestBeginGroupEmitsWorkflowCommandOnGitHubActions(t *testing.T) {
	out := groupOutput(t, map[string]string{"GITHUB_ACTIONS": "true", "CI": "true"},
		func(l *io.StandardLogger) {
			l.BeginGroup("build app")
			l.EndGroup()
		})

	if !strings.Contains(out, "::group::build app") {
		t.Errorf("output = %q, want a ::group:: command", out)
	}
	if !strings.Contains(out, "::endgroup::") {
		t.Errorf("output = %q, want an ::endgroup:: command", out)
	}
}

// ::group:: is GitHub syntax, not a CI convention. Emitting it on another
// provider prints literal noise into the log rather than collapsing anything.
func TestBeginGroupIsPlainOnNonGitHubCI(t *testing.T) {
	out := groupOutput(t, map[string]string{"GITHUB_ACTIONS": "", "CI": "true"},
		func(l *io.StandardLogger) {
			l.BeginGroup("build app")
			l.EndGroup()
		})

	if strings.Contains(out, "::group::") || strings.Contains(out, "::endgroup::") {
		t.Errorf("output = %q, want no GitHub workflow commands outside GitHub Actions", out)
	}
	if !strings.Contains(out, "build app") {
		t.Errorf("output = %q, want the group name in the plain header", out)
	}
}

// GitHub does not support nested groups: a second ::group:: before the first
// closes swallows the remainder of the log.
func TestNestedGroupsEmitOnlyTheOutermost(t *testing.T) {
	out := groupOutput(t, map[string]string{"GITHUB_ACTIONS": "true"},
		func(l *io.StandardLogger) {
			l.BeginGroup("outer")
			l.BeginGroup("inner")
			l.EndGroup()
			l.EndGroup()
		})

	if got := strings.Count(out, "::group::"); got != 1 {
		t.Errorf("::group:: count = %d, want 1; output = %q", got, out)
	}
	if got := strings.Count(out, "::endgroup::"); got != 1 {
		t.Errorf("::endgroup:: count = %d, want 1; output = %q", got, out)
	}
	if !strings.Contains(out, "::group::outer") {
		t.Errorf("output = %q, want the outermost group to be the one emitted", out)
	}
}

// A stray ::endgroup:: would close whatever GitHub had open around it, so an
// unmatched EndGroup must emit nothing. This is what lets a caller skip
// BeginGroup conditionally without having to mirror that condition on the way out.
func TestUnmatchedEndGroupEmitsNothing(t *testing.T) {
	out := groupOutput(t, map[string]string{"GITHUB_ACTIONS": "true"},
		func(l *io.StandardLogger) {
			l.EndGroup()
		})

	if strings.Contains(out, "::endgroup::") {
		t.Errorf("output = %q, want no ::endgroup:: without a matching BeginGroup", out)
	}
}

// An unescaped newline ends the workflow command early and leaks the remainder
// as ordinary log lines.
func TestGroupNameIsEscaped(t *testing.T) {
	out := groupOutput(t, map[string]string{"GITHUB_ACTIONS": "true"},
		func(l *io.StandardLogger) {
			l.BeginGroup("build\nrm -rf / 100%")
			l.EndGroup()
		})

	if !strings.Contains(out, "::group::build%0Arm -rf / 100%25") {
		t.Errorf("output = %q, want newline and percent escaped", out)
	}
	if strings.Count(out, "\n::group::") > 0 && strings.Count(out, "::group::") != 1 {
		t.Errorf("output = %q, want a single well-formed group command", out)
	}
}
