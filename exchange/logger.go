package exchange

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DEBUG level for detailed debugging information
	DEBUG LogLevel = iota
	// INFO level for general operational information
	INFO
	// WARN level for warning messages
	WARN
	// ERROR level for error messages
	ERROR
	// FATAL level for critical errors that cause the program to exit
	FATAL
)

// Logger provides a simple logging interface
type Logger struct {
	component string
	logger    *log.Logger
	level     LogLevel
}

// NewLogger creates a new logger for a specific component
func NewLogger(component string) *Logger {
	return &Logger{
		component: component,
		logger:    log.New(os.Stdout, "", 0),
		level:     INFO, // Default log level
	}
}

// SetLevel sets the minimum log level to display
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// formatMessage formats a log message with timestamp, level, and component
func (l *Logger) formatMessage(level LogLevel, message string) string {
	levelStr := "UNKNOWN"
	switch level {
	case DEBUG:
		levelStr = "DEBUG"
	case INFO:
		levelStr = "INFO"
	case WARN:
		levelStr = "WARN"
	case ERROR:
		levelStr = "ERROR"
	case FATAL:
		levelStr = "FATAL"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, levelStr, l.component, message)
}

// log logs a message at the specified level if it's above the minimum level
func (l *Logger) log(level LogLevel, message string) {
	if level >= l.level {
		l.logger.Println(l.formatMessage(level, message))
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string) {
	l.log(DEBUG, message)
}

// Info logs an informational message
func (l *Logger) Info(message string) {
	l.log(INFO, message)
}

// Warn logs a warning message
func (l *Logger) Warn(message string) {
	l.log(WARN, message)
}

// Error logs an error message
func (l *Logger) Error(message string) {
	l.log(ERROR, message)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string) {
	l.log(FATAL, message)
	os.Exit(1)
}
