package cases

import "go.uber.org/zap"

//go:generate mockgen -source=logger.go -destination=mocks/logger_mock.go -package=mocks
type Logger interface {
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}
