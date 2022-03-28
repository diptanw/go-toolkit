package logger

import (
	"fmt"
	"io"
)

// Level is the logging level type.
type Level int

// Available logging levels.
const (
	Error Level = iota
	Warn
	Info
	Debug
)

// Logger is a minimalistic leveled logger.
type Logger struct {
	writer io.Writer
	level  Level
}

// New return a new Logger instance configured with a given logging Level and
// writer.
func New(writer io.Writer, level Level) Logger {
	return Logger{
		writer: writer,
		level:  level,
	}
}

// Infof writes to the output with a Info logging Level.
func (l Logger) Infof(format string, vals ...interface{}) {
	l.write(Info, format, vals)
}

// Errorf writes to the output with a Error logging Level.
func (l Logger) Errorf(format string, vals ...interface{}) {
	l.write(Error, format, vals)
}

// Warnf writes to the output with a Warn logging Level.
func (l Logger) Warnf(format string, vals ...interface{}) {
	l.write(Warn, format, vals)
}

// Debugf writes to the output with a Debug logging Level.
func (l Logger) Debugf(format string, vals ...interface{}) {
	l.write(Debug, format, vals)
}

func (l Logger) prefix(level Level) (string, bool) {
	var prefix string

	switch level {
	case Error:
		prefix = "ERR: "
	case Warn:
		prefix = "WRN: "
	case Info:
		prefix = "INF: "
	case Debug:
		prefix = "DBG: "
	}

	return prefix, level <= l.level
}

func (l Logger) write(level Level, format string, a []interface{}) {
	if prefix, ok := l.prefix(level); ok {
		_, err := fmt.Fprintln(l.writer, fmt.Sprintf(prefix+format, a...))
		if err != nil {
			return
		}
	}
}
