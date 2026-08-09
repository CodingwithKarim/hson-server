package logger

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

const (
	colorSlate500    = "#6B7280"
	colorAlmostWhite = "#F9FAFB"
	colorCyan300     = "#53EAFD"
	colorBlue500     = "#3B82F6"
	colorAmber400    = "#FBBF24"
	colorRed500      = "#EF4444"
	colorRed700      = "#C10007"
	colorGray400     = "#9CA3AF"
	colorViolet400   = "#A78BFA"
	colorYellow500   = "#F0B100"
)

func parseLogLevel(logLevel string) log.Level {
	switch strings.ToLower(logLevel) {
	case "debug", "debugging":
		return log.DebugLevel
	case "info", "information":
		return log.InfoLevel
	case "warn", "warning":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

func customLogStyles() *log.Styles {
	styles := log.DefaultStyles()

	styles.Timestamp = styles.Timestamp.
		Foreground(lipgloss.Color(colorSlate500))
	styles.Caller = styles.Caller.
		Foreground(lipgloss.Color(colorSlate500)).
		Italic(true)

	styles.Message = styles.Message.
		Foreground(lipgloss.Color(colorAlmostWhite)).
		Bold(true)

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Foreground(lipgloss.Color(colorCyan300)).
		Bold(true)

	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Foreground(lipgloss.Color(colorBlue500)).
		Bold(true)

	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Foreground(lipgloss.Color(colorYellow500)).
		Bold(true)

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Foreground(lipgloss.Color(colorRed500)).
		Bold(true)

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("FATAL").
		Foreground(lipgloss.Color(colorRed700)).
		Bold(true)

	styles.Key = styles.Key.Foreground(lipgloss.Color(colorGray400)).Faint(true).PaddingLeft(1).PaddingRight(1)
	styles.Value = styles.Value.Foreground(lipgloss.Color(colorViolet400)).PaddingLeft(1)

	return styles
}

func getCallerFormatter() log.CallerFormatter {
	return func(file string, line int, funcName string) string {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
}
