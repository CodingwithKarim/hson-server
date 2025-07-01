package logger

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/muesli/termenv"
)

var (
	LogLevel  string
	Verbose   bool
	startTime = time.Now()
	logger    *log.Logger
)

func Setup() {
	logger = log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    Verbose,
		ReportTimestamp: Verbose,
		Level:           parseLogLevel(LogLevel),
		TimeFormat:      time.Kitchen,
		Prefix:          lipgloss.NewStyle().Render("🚀 HSON Server"),
		CallerFormatter: getCallerFormatter(),
		CallerOffset:    1,
	})

	logger.SetColorProfile(termenv.TrueColor)
	logger.SetStyles(customLogStyles())
}

func Debug(msg string, kv ...any) { logMessage(log.DebugLevel, msg, kv...) }
func Info(msg string, kv ...any)  { logMessage(log.InfoLevel, msg, kv...) }
func Warn(msg string, kv ...any)  { logMessage(log.WarnLevel, msg, kv...) }
func Error(msg string, kv ...any) { logMessage(log.ErrorLevel, msg, kv...) }
func Fatal(msg string, keyvals ...any) {
	logMessage(log.FatalLevel, msg, keyvals...)

	os.Exit(1)
}

func logMessage(level log.Level, msg string, keyvals ...any) {
	fields := append([]any(nil), keyvals...)

	if Verbose {
		fields = append(fields,
			"uptime", time.Since(startTime).String(),
			"pid", os.Getpid(),
			"goroutines", runtime.NumGoroutine(),
		)
	}

	logger.Log(level, msg, fields...)

	if level >= logger.GetLevel() {
		fmt.Println()
	}
}
