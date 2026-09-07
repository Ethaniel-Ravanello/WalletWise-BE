package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() error {
	config := zap.NewProductionConfig()

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
}
