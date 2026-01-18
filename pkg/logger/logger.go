// Package logger provides a single entry for logs aggregation and collection over the codebase.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

const (
	component = "component"
	// envLogLevel specifies name of the environmental variable that sets the log level
	envLogLevel = "LOG_LEVEL"

	// set of log levels for illustrational purposes:
	logLevelDebug    = "DEBUG"
	logLevelInfo     = "INFO"
	logLevelWarn     = "WARN"
	logLevelError    = "ERROR"
	logLevelFatal    = "FATAL"
	logLevelPanic    = "PANIC"
	logLevelNoLevel  = "NO_LEVEL"
	logLevelDisabled = "DISABLED"
	logLevelTrace    = "TRACE"
)

var level = zerolog.DebugLevel // default log level

func init() {
	switch os.Getenv(envLogLevel) {
	case logLevelDebug:
		level = zerolog.DebugLevel
	case logLevelInfo:
		level = zerolog.InfoLevel
	case logLevelWarn:
		level = zerolog.WarnLevel
	case logLevelError:
		level = zerolog.ErrorLevel
	case logLevelFatal:
		level = zerolog.FatalLevel
	case logLevelPanic:
		level = zerolog.PanicLevel
	case logLevelNoLevel:
		level = zerolog.NoLevel
	case logLevelDisabled:
		level = zerolog.Disabled
	case logLevelTrace:
		level = zerolog.TraceLevel
	default:
		// default log level is DEBUG
		level = zerolog.DebugLevel
	}
}

// NewLogger returns a wrapper for new logger instance
func NewLogger(name string) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			return filepath.Dir(fmt.Sprintf("%s/", i))
		},
	}).Level(level).With().Caller().Timestamp().Str(component, name).Logger()
}
