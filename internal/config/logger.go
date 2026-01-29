package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() error {
	logConfig := zap.NewDevelopmentEncoderConfig()
	logConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logConfig.EncodeTime = zapcore.TimeEncoderOfLayout("3:04:05 PM")
	logConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(logConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	zap.ReplaceGlobals(logger)

	return nil
}

func Sync() {
	_ = zap.L().Sync()
}
