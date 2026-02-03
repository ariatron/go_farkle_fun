package observability

import (
	"io"
	"log/slog"
	"os"
)

// Logger is the global structured logger instance
var Logger *slog.Logger

// logFile holds the log file handle
var logFile *os.File

// InitLogger initializes the structured logger with JSON output to both stdout and file
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: false,
	}

	// Open log file for Loki to read
	var err error
	logFile, err = os.OpenFile("/tmp/farkle-app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to stdout only if file can't be opened
		handler := slog.NewJSONHandler(os.Stdout, opts)
		Logger = slog.New(handler)
		slog.SetDefault(Logger)
		Logger.Warn("Failed to open log file, using stdout only", "error", err)
		return
	}

	// Write to both stdout and file using MultiWriter
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	handler := slog.NewJSONHandler(multiWriter, opts)
	Logger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(Logger)

	Logger.Info("Logger initialized", "format", "json", "outputs", "stdout+file")
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// SetLogLevel sets the logging level dynamically
func SetLogLevel(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		AddSource: false,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	Logger.Info("Log level updated", "level", level)
}
