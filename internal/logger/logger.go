package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	DEBUG LogLevel = "DEBUG"
)

// logEntry represents a single log entry with level, message, and timestamp
type logEntry struct {
	level LogLevel
	msg   string
	ts    time.Time
}

// AsyncLogger is an interface that defines logging operations
type AsyncLogger interface {
	LogEntry(level LogLevel, msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debug(msg string)
	Close()
}

// asyncLogger is the concrete implementation of AsyncLogger
// It maintains its own channel and runs a goroutine for processing log entries
type asyncLogger struct {
	entries chan logEntry
	zapLog  *zap.Logger
	done    chan struct{}
}

// NewLogger creates and returns a new AsyncLogger instance
// The logger runs in a background goroutine and will live until Close() is called
// bufferSize controls how many log entries can be queued before blocking
func NewLogger(bufferSize int) (AsyncLogger, error) {
	if bufferSize <= 0 {
		bufferSize = 100 // default buffer size
	}

	// Create the zap logger
	zapLog, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	al := &asyncLogger{
		entries: make(chan logEntry, bufferSize),
		zapLog:  zapLog,
		done:    make(chan struct{}),
	}

	// Start the background goroutine that will live forever
	// until the entries channel is closed
	go al.processEntries()

	return al, nil
}

// processEntries runs in a background goroutine and processes log entries forever
// It only stops when the entries channel is closed (via Close())
func (al *asyncLogger) processEntries() {
	defer close(al.done)
	defer al.zapLog.Sync()

	for entry := range al.entries {
		switch entry.level {
		case INFO:
			al.zapLog.Info(entry.msg, zap.Time("ts", entry.ts))
		case WARN:
			al.zapLog.Warn(entry.msg, zap.Time("ts", entry.ts))
		case ERROR:
			al.zapLog.Error(entry.msg, zap.Time("ts", entry.ts))
		case DEBUG:
			al.zapLog.Debug(entry.msg, zap.Time("ts", entry.ts))
		default:
			al.zapLog.Info(entry.msg, zap.Time("ts", entry.ts))
		}
	}
}

// LogEntry sends a log entry with the specified level and message
// This is the primary method for logging
// The timestamp is automatically set to the current time
func (al *asyncLogger) LogEntry(level LogLevel, msg string) {
	if al == nil || al.entries == nil {
		fmt.Println("Error: Logger not initialized or already closed")
		return
	}

	// Non-blocking send with timeout protection
	select {
	case al.entries <- logEntry{
		level: level,
		msg:   msg,
		ts:    time.Now(),
	}:
		// Entry sent successfully
	case <-time.After(time.Second):
		fmt.Println("Warning: Logger channel full, entry dropped")
	}
}

// Info logs an INFO level message
func (al *asyncLogger) Info(msg string) {
	al.LogEntry(INFO, msg)
}

// Warn logs a WARN level message
func (al *asyncLogger) Warn(msg string) {
	al.LogEntry(WARN, msg)
}

// Error logs an ERROR level message
func (al *asyncLogger) Error(msg string) {
	al.LogEntry(ERROR, msg)
}

// Debug logs a DEBUG level message
func (al *asyncLogger) Debug(msg string) {
	al.LogEntry(DEBUG, msg)
}

// Close gracefully shuts down the logger
// This closes the entries channel, which signals the background goroutine to stop
// Any attempt to log after Close() will panic or fail silently
func (al *asyncLogger) Close() {
	if al.entries != nil {
		close(al.entries)
		<-al.done // Wait for the goroutine to finish
	}
}
