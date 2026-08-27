package io

import (
	"os"
	"strings"
)

// envTrue is the value CI providers use for their boolean environment flags.
const envTrue = "true"

// isGitHubActions reports whether the runner is GitHub Actions specifically.
//
// Workflow commands like ::group:: are GitHub syntax, not a CI convention. Any
// other provider renders them as literal text, so they must be gated on this
// rather than on a general "is this CI" check.
func isGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == envTrue
}

// workflowCommandEscaper escapes data in a GitHub workflow command. An
// unescaped newline would end the command early and emit the remainder as
// ordinary log lines; percent is escaped first so it cannot double-encode the
// replacements that follow.
var workflowCommandEscaper = strings.NewReplacer(
	"%", "%25",
	"\r", "%0D",
	"\n", "%0A",
)

// escapeWorkflowData makes a value safe to embed in a workflow command.
func escapeWorkflowData(s string) string {
	return workflowCommandEscaper.Replace(s)
}
