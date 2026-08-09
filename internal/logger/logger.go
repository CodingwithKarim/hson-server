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

type loggerInstance struct {
	logger    *log.Logger
	verbose   bool
	startTime time.Time
}

var instance *loggerInstance

func SetupSingleton(logLevel string, isVerbose bool) {
	charmLogger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    isVerbose,
		ReportTimestamp: isVerbose,
		Level:           parseLogLevel(logLevel),
		TimeFormat:      time.Kitchen,
		Prefix:          lipgloss.NewStyle().Render("🚀 HJSON Server"),
		CallerFormatter: getCallerFormatter(),
		CallerOffset:    1,
	})

	charmLogger.SetColorProfile(termenv.TrueColor)
	charmLogger.SetStyles(customLogStyles())

	instance = &loggerInstance{
		logger:    charmLogger,
		verbose:   isVerbose,
		startTime: time.Now(),
	}
}

func Debug(msg string, kv ...any)      { instance.logMessage(log.DebugLevel, msg, kv...) }
func Info(msg string, kv ...any)       { instance.logMessage(log.InfoLevel, msg, kv...) }
func Warn(msg string, kv ...any)       { instance.logMessage(log.WarnLevel, msg, kv...) }
func Error(msg string, kv ...any)      { instance.logMessage(log.ErrorLevel, msg, kv...) }
func Fatal(msg string, keyvals ...any) { instance.logMessage(log.FatalLevel, msg, keyvals...) }

func (l *loggerInstance) logMessage(level log.Level, msg string, keyvals ...any) {
	fields := append([]any(nil), keyvals...)

	if l.verbose {
		fields = append(fields,
			"uptime", time.Since(l.startTime).String(),
			"pid", os.Getpid(),
			"goroutines", runtime.NumGoroutine(),
		)
	}

	l.logger.Log(level, msg, fields...)

	if level >= l.logger.GetLevel() {
		fmt.Println()
	}
}
