package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	kit "github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultJSONFormatter(t *testing.T) {
	wr := bytes.Buffer{}
	logger := New(&wr)

	logger.Log("msg", "test")

	assert.True(t, json.Valid(wr.Bytes()))
	assert.Contains(t, wr.String(), `"msg":"test"`)
}

func TestNew_DefaultPrefix(t *testing.T) {
	wr := bytes.Buffer{}
	logger := New(&wr)

	logger.Log("")

	assert.Contains(t, wr.String(), `"timestamp":`)
	assert.Contains(t, wr.String(), `"caller":`)
}

func TestNewNopLogger(t *testing.T) {
	log := NewNopLogger()

	assert.NotNil(t, log)
	assert.IsType(t, kit.NewNopLogger(), log.kitLogger)
}

func TestLogger_WithErrHandler(t *testing.T) {
	var handled bool

	handler := func(err error) {
		assert.Error(t, err)
		assert.EqualError(t, err, assert.AnError.Error())

		handled = true
	}

	logger := Logger{}.WithErrHandler(handler)
	logger.errHandler(assert.AnError)

	assert.True(t, handled)
}

func TestLogger_WithPrefix(t *testing.T) {
	logger := Logger{
		kitLogger:  &kitLoggerFake{},
		errHandler: func(error) {},
	}.WithPrefix("prefix", "1")

	assert.NotNil(t, logger.kitLogger)
	assert.NotNil(t, logger.errHandler)
}

func TestLogger_WithLevel(t *testing.T) {
	logger := Logger{
		kitLogger:  &kitLoggerFake{},
		errHandler: func(error) {},
	}.WithLevel(LevelInfo)

	assert.NotNil(t, logger.kitLogger)
	assert.NotNil(t, logger.errHandler)
}

func TestLogger_Error(t *testing.T) {
	tests := map[string]struct {
		giveFilter Level
		wantLogged bool
	}{
		"error with error filter": {
			LevelError,
			true,
		},
		"error with warn filter": {
			LevelWarn,
			true,
		},
		"error with info filter": {
			LevelInfo,
			true,
		},
		"error with debug filter": {
			LevelDebug,
			true,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			kitLogger := kitLoggerFake{}
			logger := Logger{kitLogger: &kitLogger}.WithLevel(tc.giveFilter)

			logger.Error("test error")

			if tc.wantLogged {
				assert.NotEmpty(t, kitLogger.recEntry)
				assert.Contains(t, kitLogger.recEntry, "level")
			} else {
				assert.Empty(t, kitLogger.recEntry)
			}
		})
	}
}

func TestLogger_Warn(t *testing.T) {
	tests := map[string]struct {
		giveFilter Level
		wantLogged bool
	}{
		"warn with error filter": {
			LevelError,
			false,
		},
		"warn with warn filter": {
			LevelWarn,
			true,
		},
		"warn with info filter": {
			LevelInfo,
			true,
		},
		"warn with debug filter": {
			LevelDebug,
			true,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			kitLogger := kitLoggerFake{}
			logger := Logger{kitLogger: &kitLogger}.WithLevel(tc.giveFilter)

			logger.Warn("test warn")

			if tc.wantLogged {
				assert.NotEmpty(t, kitLogger.recEntry)
				assert.Contains(t, kitLogger.recEntry, "level")
			} else {
				assert.Empty(t, kitLogger.recEntry)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	tests := map[string]struct {
		giveFilter Level
		wantLogged bool
	}{
		"info with error filter": {
			LevelError,
			false,
		},
		"info with warn filter": {
			LevelWarn,
			false,
		},
		"info with info filter": {
			LevelInfo,
			true,
		},
		"info with debug filter": {
			LevelDebug,
			true,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			kitLogger := kitLoggerFake{}
			logger := Logger{kitLogger: &kitLogger}.WithLevel(tc.giveFilter)

			logger.Info("test info")

			if tc.wantLogged {
				assert.NotEmpty(t, kitLogger.recEntry)
				assert.Contains(t, kitLogger.recEntry, "level")
			} else {
				assert.Empty(t, kitLogger.recEntry)
			}
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	tests := map[string]struct {
		giveFilter Level
		wantLogged bool
	}{
		"debug with error filter": {
			LevelError,
			false,
		},
		"debug with warn filter": {
			LevelWarn,
			false,
		},
		"debug with info filter": {
			LevelInfo,
			false,
		},
		"debug with debug filter": {
			LevelDebug,
			true,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			kitLogger := kitLoggerFake{}
			logger := Logger{kitLogger: &kitLogger}.WithLevel(tc.giveFilter)

			logger.Debug("test debug")

			if tc.wantLogged {
				assert.NotEmpty(t, kitLogger.recEntry)
				assert.Contains(t, kitLogger.recEntry, "level")
			} else {
				assert.Empty(t, kitLogger.recEntry)
			}
		})
	}
}

func TestLogger_Log(t *testing.T) {
	kitLogger := kitLoggerFake{}
	logger := Logger{kitLogger: &kitLogger}

	logger.Log("msg", "test", "lvl", "critical")

	assert.Contains(t, kitLogger.recEntry, "msg")
	assert.Equal(t, kitLogger.recEntry["msg"], "test")

	assert.Contains(t, kitLogger.recEntry, "lvl")
	assert.Equal(t, kitLogger.recEntry["lvl"], "critical")
}

func TestHandleLogError(t *testing.T) {
	tests := map[string]struct {
		giveKitLogger     kitLoggerFake
		wantHandlerCalled bool
	}{
		"no error": {
			kitLoggerFake{},
			false,
		},
		"return error": {
			kitLoggerFake{retErr: assert.AnError},
			true,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			var handled bool

			handleLogError(func(err error) {
				handled = true
			}, &tc.giveKitLogger, "msg", nil)

			assert.Equal(t, tc.wantHandlerCalled, handled)
			assert.Contains(t, tc.giveKitLogger.recEntry, "message")
		})
	}
}

type kitLoggerFake struct {
	recEntry map[string]interface{}
	retErr   error
}

func (l *kitLoggerFake) Log(keyvals ...interface{}) error {
	l.recEntry = make(map[string]interface{})

	for i := 0; i < len(keyvals)-1; i += 2 {
		l.recEntry[keyvals[i].(string)] = keyvals[i+1]
	}

	return l.retErr
}
