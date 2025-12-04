package middleware

import (
	"fmt"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/styltsou/url-shortener/server/pkg/logger"
	"go.uber.org/zap"
)

// ANSI color codes for HTTP methods
const (
	colorReset   = "\033[0m"
	colorBlue    = "\033[35m"
	colorGreen   = "\033[32m"
	colorOrange  = "\033[38;5;208m"
	colorRed     = "\033[31m"
	colorCyan    = "\033[36m"
	colorMagenta = "\033[35m"
)

// getMethodColor returns ANSI color code for HTTP method
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return colorGreen
	case "POST":
		return colorBlue
	case "PUT", "PATCH":
		return colorOrange
	case "DELETE":
		return colorRed
	case "OPTIONS", "HEAD":
		return colorCyan
	default:
		return colorMagenta
	}
}

// CustomLogFormatter implements chi's LogFormatter interface to use our custom logger
type CustomLogFormatter struct {
	logger logger.Logger
}

// NewLogEntry creates a new log entry for a request
func (f *CustomLogFormatter) NewLogEntry(r *http.Request) chimw.LogEntry {
	return &CustomLogEntry{
		logger:    f.logger,
		request:   r,
		startTime: time.Now(),
	}
}

// CustomLogEntry implements chi's LogEntry interface
type CustomLogEntry struct {
	logger    logger.Logger
	request   *http.Request
	startTime time.Time
}

// Write logs the completion of the request
func (e *CustomLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	// Use logger with additional skip to hide caller location
	log := e.logger.WithCallerSkip(1)

	// In dev mode, use compact format with colored HTTP method
	if e.logger.IsDev() {
		methodColor := getMethodColor(e.request.Method)
		msg := fmt.Sprintf("%s%s%s %s %d %v",
			methodColor,
			e.request.Method,
			colorReset,
			e.request.URL.Path,
			status,
			elapsed.Round(time.Millisecond))

		// Always use Info level in dev mode
		log.Info(msg)
		return
	}

	// Production: use structured logging
	// Log at appropriate level based on status code
	switch {
	case status >= 500:
		log.Error("Request completed",
			zap.String("method", e.request.Method),
			zap.String("path", e.request.URL.Path),
			zap.String("request_id", chimw.GetReqID(e.request.Context())),
			zap.Int("status", status),
			zap.Duration("latency", elapsed),
			zap.String("user_agent", e.request.UserAgent()),
		)
	case status >= 400:
		log.Warn("Request completed",
			zap.String("method", e.request.Method),
			zap.String("path", e.request.URL.Path),
			zap.String("request_id", chimw.GetReqID(e.request.Context())),
			zap.Int("status", status),
			zap.Duration("latency", elapsed),
			zap.String("user_agent", e.request.UserAgent()),
		)
	default:
		log.Info("Request completed",
			zap.String("method", e.request.Method),
			zap.String("path", e.request.URL.Path),
			zap.String("request_id", chimw.GetReqID(e.request.Context())),
			zap.Int("status", status),
			zap.Duration("latency", elapsed),
			zap.String("user_agent", e.request.UserAgent()),
		)
	}
}

// Panic logs panic information
// NOTE: This should rarely be called. Panics indicate bugs or unexpected conditions.
// We should always return errors instead of panicking. This method is required by
// Chi's LogEntry interface and is called by Recoverer middleware when it recovers from panics.
func (e *CustomLogEntry) Panic(v interface{}, stack []byte) {
	log := e.logger.WithCallerSkip(1)
	log.Error("Unexpected panic recovered - this indicates a bug",
		zap.String("method", e.request.Method),
		zap.String("path", e.request.URL.Path),
		zap.String("request_id", chimw.GetReqID(e.request.Context())),
		zap.Any("panic_value", v),
		zap.String("stack_trace", string(stack)),
	)
}

// RequestLogger returns request logging middleware that uses our custom logger.
// In development, it uses a compact colored format for readability.
// In production, it uses structured logging with zap fields for log aggregation.
func RequestLogger(log logger.Logger) func(http.Handler) http.Handler {
	// Always use our custom formatter, which handles dev vs prod formatting internally
	return chimw.RequestLogger(&CustomLogFormatter{logger: log})
}
