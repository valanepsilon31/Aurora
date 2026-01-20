package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	enabled     bool
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// Init initializes the logger with lumberjack for log rotation
// Call this only in desktop mode
func Init(logPath string) {
	lj := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5, // MB
		MaxBackups: 0,
		MaxAge:     0,
		Compress:   false,
	}

	flags := log.Ldate | log.Ltime | log.Lshortfile

	infoLogger = log.New(lj, "INFO  ", flags)
	warnLogger = log.New(lj, "WARN  ", flags)
	errorLogger = log.New(lj, "ERROR ", flags)

	enabled = true

	// Visual separator between sessions
	infoLogger.Output(2, "========================================")
	Info("Logger initialized")
}

// Info logs an info message
func Info(format string, v ...any) {
	if !enabled {
		return
	}
	infoLogger.Output(2, fmt.Sprintf(format, v...))
}

// Warn logs a warning message
func Warn(format string, v ...any) {
	if !enabled {
		return
	}
	warnLogger.Output(2, fmt.Sprintf(format, v...))
}

// Error logs an error message
func Error(format string, v ...any) {
	if !enabled {
		return
	}
	errorLogger.Output(2, fmt.Sprintf(format, v...))
}

// LogIfErr logs an error if err is not nil and returns true if there was an error
func LogIfErr(err error, context string) bool {
	if err == nil {
		return false
	}
	if enabled {
		errorLogger.Output(2, fmt.Sprintf("%s: %v", context, err))
	}
	return true
}

// Close can be called to ensure logs are flushed (lumberjack handles this automatically)
func Close() {
	if enabled {
		Info("Logger closing")
	}
}

// GetLogPath returns the default log path next to the executable
func GetLogPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "aurora.log"
	}
	return filepath.Join(filepath.Dir(exe), "aurora.log")
}
