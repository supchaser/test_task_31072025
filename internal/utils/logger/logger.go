package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log   *zap.Logger
	Sugar *zap.SugaredLogger
)

func Init(mode string) error {
	var cfg zap.Config

	switch mode {
	case "dev":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	case "prod":
		cfg = zap.Config{
			Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
			Development:      false,
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}
	default:
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = logger
	Sugar = logger.Sugar()
	return nil
}

func Sync() error {
	return Log.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

func Debugf(format string, args ...interface{}) {
	Sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	Sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Sugar.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Sugar.Fatalf(format, args...)
}

func InitTestLogger() {
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	Log = logger
	Sugar = logger.Sugar()
}
