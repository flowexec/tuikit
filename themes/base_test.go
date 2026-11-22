package themes_test

import (
	"testing"

	"charm.land/lipgloss/v2"
	"charm.land/log/v2"

	"github.com/flowexec/tuikit/themes"
)

func TestLoggerLevelsShareWidth(t *testing.T) {
	levels := map[string]log.Level{
		"debug":  log.DebugLevel,
		"info":   log.InfoLevel,
		"notice": themes.LogNoticeLevel,
		"warn":   log.WarnLevel,
		"error":  log.ErrorLevel,
		"fatal":  log.FatalLevel,
	}
	styles := themes.EverforestTheme().LoggerStyles()
	want := lipgloss.Width(styles.Levels[log.InfoLevel].String())
	for name, lvl := range levels {
		if got := lipgloss.Width(styles.Levels[lvl].String()); got != want {
			t.Errorf("%s level is %d cells wide, want %d like info", name, got, want)
		}
	}
}
