package internal

import (
	"go.uber.org/zap"
)

func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	zap.AddStacktrace(logger.Level())
	return logger, err
}
