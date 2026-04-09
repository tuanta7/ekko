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

	// Bubble Tea renders the TUI to stdout, so application logs must go to a file
	cfg.OutputPaths = []string{"ekko.log"}
	cfg.ErrorOutputPaths = []string{"ekko.log"}

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
