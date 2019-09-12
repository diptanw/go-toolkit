package logger

import (
	"fmt"
	"io"
	"os"

	kit "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Logger wraps go-kit Logger and is responsible for leveled logging.
type Logger struct {
	kitLogger  kit.Logger
	errHandler ErrHandler
}

// Level is an enum for the logging level flag.
type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

// ErrHandler is a fallback handler when logging failed.
type ErrHandler func(error)

// New returns a new Logger initialized with the giver writer.
// If writer is nil, then the os.Stdout is used.
func New(w io.Writer) Logger {
	if w == nil {
		w = os.Stdout
	}

	const callDepth = 5

	kitLogger := kit.NewJSONLogger(kit.NewSyncWriter(w))
	kitLogger = kit.With(kitLogger, "timestamp", kit.DefaultTimestampUTC)
	kitLogger = kit.With(kitLogger, "caller", kit.Caller(callDepth))

	return Logger{
		kitLogger: kitLogger,
		errHandler: func(err error) {
			fmt.Fprintf(os.Stderr, "error writing to the log: %s", err)
		},
	}
}

// NewNopLogger returns a logger that doesn't do anything.
func NewNopLogger() Logger {
	return Logger{
		kitLogger: kit.NewNopLogger(),
	}
}

// WithLevel creates a new logger with a custom minimal logging level.
// It can raise the log level filter, but cannot lower the log level filter.
func (l Logger) WithLevel(lvl Level) Logger {
	return Logger{
		kitLogger:  level.NewFilter(l.kitLogger, toAllowedLevel(lvl)),
		errHandler: l.errHandler,
	}
}

// WithPrefix creates a new Logger with the keyvals prepended.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) WithPrefix(keyvals ...interface{}) Logger {
	return Logger{
		kitLogger:  kit.WithPrefix(l.kitLogger, keyvals...),
		errHandler: l.errHandler,
	}
}

// WithErrHandler creates a new Logger with the custom error handler.
func (l Logger) WithErrHandler(handler ErrHandler) Logger {
	return Logger{
		kitLogger:  l.kitLogger,
		errHandler: handler,
	}
}

// toAllowedLevel determines the level of log events allowed to pass.
func toAllowedLevel(l Level) level.Option {
	switch l {
	case LevelDebug:
		return level.AllowDebug()
	case LevelWarn:
		return level.AllowWarn()
	case LevelInfo:
		return level.AllowInfo()
	default:
		return level.AllowError()
	}
}

// Log emits a raw not leveled log event.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) Log(keyvals ...interface{}) error {
	return l.kitLogger.Log(keyvals...)
}

// Info emits an INFO level log event.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) Info(msg string, keyvals ...interface{}) {
	handleLogError(l.errHandler, level.Info(l.kitLogger), msg, keyvals)
}

// Debug emits a DEBUG level log event.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) Debug(msg string, keyvals ...interface{}) {
	handleLogError(l.errHandler, level.Debug(l.kitLogger), msg, keyvals)
}

// Warn emits a WARN level log event.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) Warn(msg string, keyvals ...interface{}) {
	handleLogError(l.errHandler, level.Warn(l.kitLogger), msg, keyvals)
}

// Error emits an ERROR level log event.
// keyvals is a variadic sequence of alternating keys and values.
func (l Logger) Error(msg string, keyvals ...interface{}) {
	handleLogError(l.errHandler, level.Error(l.kitLogger), msg, keyvals)
}

func handleLogError(handler ErrHandler, kitLogger kit.Logger, msg string, keyvals []interface{}) {
	if err := kitLogger.Log(append(keyvals, "message", msg)...); err != nil {
		handler(err)
	}
}
