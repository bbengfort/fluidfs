// Mechanism for handling application level logging

package fluid

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

//===========================================================================
// Log Level Type
//===========================================================================

// LogLevel characterizes the severity of the log message.
type LogLevel int

// Severity levels of log messages.
const (
	LevelDebug LogLevel = 1 + iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String representations of the various log levels.
var levelNames = []string{
	"DEBUG", "INFO", "WARN", "ERROR", "FATAL",
}

// String representation of the log level.
func (level LogLevel) String() string {
	return levelNames[level-1]
}

// LevelFromString parses a string and returns the LogLevel
func LevelFromString(level string) LogLevel {
	// Perform string cleanup for matching
	level = strings.ToUpper(level)
	level = strings.Trim(level, " ")

	switch level {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

//===========================================================================
// Logger wrapper for log.Logger and logging initialization methods
//===========================================================================

// Logger wraps the log.Logger to write to a file on demand and to specify a
// miminum severity that is allowed for writing.
type Logger struct {
	Level  LogLevel       // The minimum severity to log to
	logger *log.Logger    // The wrapped logger for concurrent logging
	output io.WriteCloser // Handle to the open log file or writer object
}

// InitLogger creates a Logger object by passing a configuration that contains
// the minimum log level and an optional path to write the log out to.
func InitLogger(config *LoggingConfig) (*Logger, error) {
	logger := new(Logger)
	logger.Level = LevelFromString(config.Level)

	// If a path is specified create a handle to the writer.
	if config.Path != "" {

		var err error
		logger.output, err = os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}

	} else {
		logger.output = os.Stdout
	}

	logger.logger = log.New(logger.output, "", 0)

	return logger, nil
}

// Close the logger and any open file handles.
func (logger *Logger) Close() error {
	if err := logger.output.Close(); err != nil {
		return err
	}
	return nil
}

// GetHandler returns the io.Writer object that is on the logger.
func (logger *Logger) GetHandler() io.Writer {
	return logger.output
}

// SetHandler sets a new io.WriteCloser object onto the logger
func (logger *Logger) SetHandler(writer io.WriteCloser) {
	logger.output = writer
	logger.logger.SetOutput(writer)
}

//===========================================================================
// Logging handlers
//===========================================================================

// Log a message at the appropriate severity. The Log method behaves as a
// format function, and a layout string can be passed with arguments.
// The current logging format is "%(level)s [%(jsontime)s]: %(message)s"
func (logger *Logger) Log(layout string, level LogLevel, args ...interface{}) {

	// Only log if the log level matches the log request
	if level >= logger.Level {
		msg := fmt.Sprintf(layout, args...)
		msg = fmt.Sprintf("%-7s [%s]: %s", level, time.Now().Format(JSONDateTime), msg)

		// If level is fatal then log fatal.
		if level == LevelFatal {
			logger.logger.Fatalln(msg)
		} else {
			logger.logger.Println(msg)
		}

	}

}

// Debug message helper function
func (logger *Logger) Debug(msg string, args ...interface{}) {
	logger.Log(msg, LevelDebug, args...)
}

// Info message helper function
func (logger *Logger) Info(msg string, args ...interface{}) {
	logger.Log(msg, LevelInfo, args...)
}

// Warn message helper function
func (logger *Logger) Warn(msg string, args ...interface{}) {
	logger.Log(msg, LevelWarn, args...)
}

// Error message helper function
func (logger *Logger) Error(msg string, args ...interface{}) {
	logger.Log(msg, LevelError, args...)
}

// Fatal message helper function
func (logger *Logger) Fatal(msg string, args ...interface{}) {
	logger.Log(msg, LevelFatal, args...)
}