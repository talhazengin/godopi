package logger

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = false

	var err error

	if logger, err = cfg.Build(); err != nil {
		panic(err)
	}
}

func Logger() *zap.Logger {
	return logger
}
