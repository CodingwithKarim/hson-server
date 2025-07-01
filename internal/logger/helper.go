package logger

import (
	"flag"
	"fmt"
	"hson-server/internal/utils"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

func RegisterFlags() {
	flag.StringVar(&LogLevel, "log-level", "info", "log level: debug, info, warn, error")
	flag.BoolVar(&Verbose, "verbose", false, "enable verbose logging (adds file and line number and extra fields)")
}

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
		Foreground(lipgloss.Color(utils.ColorSlate500))
	styles.Caller = styles.Caller.
		Foreground(lipgloss.Color(utils.ColorSlate500)).
		Italic(true)

	styles.Message = styles.Message.
		Foreground(lipgloss.Color(utils.ColorAlmostWhite)).
		Bold(true)

	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Foreground(lipgloss.Color(utils.ColorCyan300)).
		Bold(true)

	styles.Levels[log.InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Foreground(lipgloss.Color(utils.ColorBlue500)).
		Bold(true)

	styles.Levels[log.WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Foreground(lipgloss.Color(utils.ColorYellow500)).
		Bold(true)

	styles.Levels[log.ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Foreground(lipgloss.Color(utils.ColorRed500)).
		Bold(true)

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString("FATAL").
		Foreground(lipgloss.Color(utils.ColorRed700)).
		Bold(true)

	styles.Key = styles.Key.Foreground(lipgloss.Color(utils.ColorGray400)).Faint(true).PaddingLeft(1).PaddingRight(1)
	styles.Value = styles.Value.Foreground(lipgloss.Color(utils.ColorViolet400)).PaddingLeft(1)

	return styles
}

func getCallerFormatter() log.CallerFormatter {
	return func(file string, line int, funcName string) string {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
}
