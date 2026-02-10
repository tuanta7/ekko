package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(level zapcore.Level) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "json"
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.OutputPaths = []string{"stdout"}

	zl, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: zl,
	}, nil
}
