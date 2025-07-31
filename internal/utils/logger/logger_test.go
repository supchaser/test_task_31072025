package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInit(t *testing.T) {
	t.Run("development mode", func(t *testing.T) {
		err := Init("dev")
		require.NoError(t, err)
		assert.NotNil(t, Log)
		assert.NotNil(t, Sugar)
		assert.True(t, Log.Core().Enabled(zapcore.DebugLevel))
	})

	t.Run("production mode", func(t *testing.T) {
		err := Init("prod")
		require.NoError(t, err)
		assert.NotNil(t, Log)
		assert.NotNil(t, Sugar)
		assert.False(t, Log.Core().Enabled(zapcore.DebugLevel))
		assert.True(t, Log.Core().Enabled(zapcore.InfoLevel))
	})

	t.Run("invalid mode", func(t *testing.T) {
		err := Init("invalid")
		require.NoError(t, err)
		assert.NotNil(t, Log)
		assert.NotNil(t, Sugar)
	})
}

func TestLoggingFunctions(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	Log = zap.New(core)
	Sugar = Log.Sugar()

	t.Run("structured logging", func(t *testing.T) {
		recorded.TakeAll()

		msg := "test message"
		field := zap.String("key", "value")

		Debug(msg, field)
		Info(msg, field)
		Warn(msg, field)
		Error(msg, field)

		logs := recorded.All()
		require.Len(t, logs, 4)

		assert.Equal(t, zapcore.DebugLevel, logs[0].Level)
		assert.Equal(t, msg, logs[0].Message)
		assert.Equal(t, "value", logs[0].ContextMap()["key"])

		assert.Equal(t, zapcore.InfoLevel, logs[1].Level)
		assert.Equal(t, zapcore.WarnLevel, logs[2].Level)
		assert.Equal(t, zapcore.ErrorLevel, logs[3].Level)
	})

	t.Run("formatted logging", func(t *testing.T) {
		recorded.TakeAll()

		format := "formatted %s"
		arg := "message"

		Debugf(format, arg)
		Infof(format, arg)
		Warnf(format, arg)
		Errorf(format, arg)

		logs := recorded.All()
		require.Len(t, logs, 4)

		assert.Equal(t, "formatted message", logs[0].Message)
		assert.Equal(t, zapcore.DebugLevel, logs[0].Level)
		assert.Equal(t, "formatted message", logs[1].Message)
		assert.Equal(t, zapcore.InfoLevel, logs[1].Level)
		assert.Equal(t, "formatted message", logs[2].Message)
		assert.Equal(t, zapcore.WarnLevel, logs[2].Level)
		assert.Equal(t, "formatted message", logs[3].Message)
		assert.Equal(t, zapcore.ErrorLevel, logs[3].Level)
	})
}

func TestSync(t *testing.T) {
	var buf bytes.Buffer
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	Log = zap.New(core)
	Sugar = Log.Sugar()

	Info("test sync message")
	err := Sync()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test sync message")
}

func TestInitTestLogger(t *testing.T) {
	InitTestLogger()
	assert.NotNil(t, Log)
	assert.NotNil(t, Sugar)
	assert.True(t, Log.Core().Enabled(zapcore.DebugLevel))
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	)
	Log = zap.New(core)
	Sugar = Log.Sugar()

	Info("test output message")
	output := buf.String()
	assert.Contains(t, output, "test output message")
	assert.Contains(t, output, "INFO")
}
