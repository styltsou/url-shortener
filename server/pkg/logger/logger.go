package logger

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Field wraps zap.Field to provide abstraction
type Field = zap.Field

// Logger defines the logging interface for dependency injection.
// This allows using mock loggers in tests and swapping implementations.
// Use zap.String(), zap.Int(), zap.Error(), etc. to create fields.
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithCallerSkip(skip int) Logger
	IsDev() bool
	Sync() error
}

// ZapLogger is the concrete implementation of the Logger interface using zap.
// It wraps zap.Logger to provide a consistent logging interface.
type ZapLogger struct {
	logger *zap.Logger
	isDev  bool
}

// New creates a new logger instance based on the environment.
// Development: Pretty console output with colors
// Production: JSON output for log aggregation
// Returns a concrete ZapLogger instance that implements the Logger interface.
func New(env string) (*ZapLogger, error) {
	var zapLogger *zap.Logger
	var err error

	isDev := env == "dev" || env == "development"

	if isDev {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = coloredTimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		config.EncoderConfig.ConsoleSeparator = " "

		// Wrap the encoder to pretty-print JSON strings and structs
		encoder := zapcore.NewConsoleEncoder(config.EncoderConfig)
		prettyEncoder := &prettyEncoder{Encoder: encoder}

		// Build logger with custom encoder for pretty printing
		zapLogger = zap.New(
			zapcore.NewCore(prettyEncoder, zapcore.AddSync(os.Stdout), config.Level),
			zap.AddCallerSkip(1),
			zap.Development(),
		)
	} else {
		config := zap.NewProductionConfig()
		zapLogger, err = config.Build(zap.AddCallerSkip(1))
	}

	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger: zapLogger, isDev: isDev}, nil
}

// Info logs an info-level message with optional zap fields
// Usage: logger.Info("message", zap.String("key", "value"), zap.Int("count", 42))
func (l *ZapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

// Error logs an error-level message with optional zap fields
// Usage: logger.Error("message", zap.String("key", "value"), zap.Error(err))
func (l *ZapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

// Warn logs a warning-level message with optional zap fields
// Usage: logger.Warn("message", zap.String("key", "value"))
func (l *ZapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

// Debug logs a debug-level message with optional zap fields
// Usage: logger.Debug("message", zap.String("key", "value"))
func (l *ZapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

// Fatal logs a fatal-level message and then calls os.Exit(1)
// Usage: logger.Fatal("message", zap.String("key", "value"))
func (l *ZapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger with the given zap fields
// Usage: logger.With(zap.String("key", "value"), zap.Int("count", 42))
func (l *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{logger: l.logger.With(fields...), isDev: l.isDev}
}

// Sync flushes any buffered log entries
// Should be called before application exits
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// ZapLogger returns the underlying zap.Logger for integrations that need it
func (l *ZapLogger) ZapLogger() *zap.Logger {
	return l.logger
}

// WithCallerSkip creates a new logger with additional caller skip
func (l *ZapLogger) WithCallerSkip(skip int) Logger {
	return &ZapLogger{
		logger: l.logger.WithOptions(zap.AddCallerSkip(skip)),
		isDev:  l.isDev,
	}
}

// IsDev returns true if the logger is configured for development mode
func (l *ZapLogger) IsDev() bool {
	return l.isDev
}

// prettyEncoder wraps a zapcore.Encoder to pretty-print JSON strings and structs in development mode
type prettyEncoder struct {
	zapcore.Encoder
}

// Clone creates a copy of the encoder
func (e *prettyEncoder) Clone() zapcore.Encoder {
	return &prettyEncoder{Encoder: e.Encoder.Clone()}
}

// EncodeEntry encodes a log entry, pretty-printing JSON strings and complex types
func (e *prettyEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Process fields to pretty-print JSON strings and complex types
	for i := range fields {
		switch fields[i].Type {
		case zapcore.StringType:
			// Pretty-print JSON strings
			if isJSON(fields[i].String) {
				var prettyJSON interface{}
				if err := json.Unmarshal([]byte(fields[i].String), &prettyJSON); err == nil {
					if prettyBytes, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
						fields[i].String = string(prettyBytes)
					}
				}
			}
		case zapcore.ReflectType, zapcore.ObjectMarshalerType:
			// Pretty-print structs and objects by converting to JSON
			if fields[i].Interface != nil {
				if prettyBytes, err := json.MarshalIndent(fields[i].Interface, "", "  "); err == nil {
					// Replace the field with a pretty-printed string version
					fields[i] = zapcore.Field{
						Key:       fields[i].Key,
						Type:      zapcore.StringType,
						String:    string(prettyBytes),
						Interface: nil,
					}
				}
			}
		}
	}

	// Encode with the base encoder
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return buf, err
	}

	// Add newline after message and format fields on separate lines
	return formatBufferWithNewlines(buf), nil
}

// formatBufferWithNewlines reformats the buffer to add newline after message and format fields
func formatBufferWithNewlines(buf *buffer.Buffer) *buffer.Buffer {
	content := buf.String()

	// Remove trailing newline if present (we'll add it back)
	hasNewline := strings.HasSuffix(content, "\n")
	if hasNewline {
		content = strings.TrimSuffix(content, "\n")
	}

	// Simple approach: find patterns like " key=" (space followed by word and =)
	// This indicates the start of a field, replace the space with newline+tab
	// Use regex to find field boundaries: space followed by word characters and =
	fieldPattern := regexp.MustCompile(`(\s)([a-zA-Z_][a-zA-Z0-9_]*=)`)

	// Replace space before field with newline+tab
	formatted := fieldPattern.ReplaceAllString(content, "\n\t$2")

	// Build new buffer
	newBuf := buffer.NewPool().Get()
	newBuf.AppendString(formatted)

	// Add final newline
	newBuf.AppendString("\n")

	return newBuf
}

// isJSON checks if a string is valid JSON
func isJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil && len(s) > 0 && (s[0] == '{' || s[0] == '[')
}

// coloredTimeEncoder formats timestamps with a bold color for better visual separation
func coloredTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// ANSI color codes
	// Use bold cyan for timestamps (1 = bold, 36 = cyan)
	const timeColor = "\033[1;36m" // Bold Cyan
	const resetColor = "\033[0m"

	// Format: colored timestamp
	enc.AppendString(timeColor + t.Format("2006-01-02T15:04:05.000Z0700") + resetColor)
}
