package internal

import (
	"go.uber.org/zap"
)

// Function initializes and returns logger entity
func NewLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	zap.AddStacktrace(logger.Level())
	return logger, err
}
