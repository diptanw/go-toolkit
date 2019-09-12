/*
Package logger provides a basic structured logger.

The motivation is to build flexible and replaceable structured logger
abstraction to maintain the codding style. It is based on the
github.com/go-kit/kit/log fundamental interface and a type that
creates log events from key/value data. This package provides good
building blocks for custom leveled and structured logging and is
evolutionary of logging interface.

Basic Usage

Create a new logger with the given writer.
Log stream will be written to stdout and JSON formatted.

	l := logger.New(os.Stdout)

Log message without the default message field. It will only include
default prefixes.

	l.Log("example", "test")

Log message with debug level. It will also add the logging level and
message property to the output.

	l.Debug("debug example")

Log message with info level and custom fields. It will write additional
properties for the current entry.

	l.Info("info example", "field_1", "value_1")

Add default prefix to the logger. In addition to default fields, it will
add a custom prefix for every following log entry.

	l = logger.WithPrefix("filed_1", "value_1")

Set the error level filter. It will only output entries with a minimum
allowed logging level.

	l = logger.WithLevel(LevelError)

Possible logging levels are:

	LevelDebug
	LevelInfo
	LevelWarn
	LevelError

*/
package logger
