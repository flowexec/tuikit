package themes

import (
	"charm.land/log/v2"
)

type OutputLevel string

const (
	OutputLevelSuccess OutputLevel = "success"
	OutputLevelNotice  OutputLevel = "notice"
	OutputLevelInfo    OutputLevel = "info"
	OutputLevelWarning OutputLevel = "warning"
	OutputLevelError   OutputLevel = "error"

	// LogNoticeLevel is a custom log level for notices.
	LogNoticeLevel = log.InfoLevel + 1
)

// HelpKey represents a single keybinding with its description for display in help overlays.
type HelpKey struct {
	Key  string
	Desc string
}
