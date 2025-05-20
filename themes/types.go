package themes

import (
	"github.com/charmbracelet/log"
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
