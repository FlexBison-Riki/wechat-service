package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger wraps logging functionality
type Logger struct {
	*log.Logger
	level  string
	output io.Writer
}

// New creates a new Logger instance
func New() *Logger {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: failed to create log directory: %v", err)
	}

	currentDate := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("wechat-service_%s.log", currentDate))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: failed to open log file: %v", err)
		file = os.Stdout
	}

	return &Logger{
		Logger: log.New(file, "", log.LstdFlags),
		level:  "info",
		output: file,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.Printf("[DEBUG] "+fmt.Sprint(v...)+"\n")
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Printf("[DEBUG] "+fmt.Sprintf(format, v...)+"\n")
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.Printf("[INFO] "+fmt.Sprint(v...)+"\n")
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Printf("[INFO] "+fmt.Sprintf(format, v...)+"\n")
}

// Warn logs a warning message
func (l *Logger) Warn(v ...interface{}) {
	l.Printf("[WARN] "+fmt.Sprint(v...)+"\n")
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Printf("[WARN] "+fmt.Sprintf(format, v...)+"\n")
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.Printf("[ERROR] "+fmt.Sprint(v...)+"\n")
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Printf("[ERROR] "+fmt.Sprintf(format, v...)+"\n")
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.Printf("[FATAL] "+fmt.Sprint(v...)+"\n")
	os.Exit(1)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Printf("[FATAL] "+fmt.Sprintf(format, v...)+"\n")
	os.Exit(1)
}
