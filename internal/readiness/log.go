package readiness

import (
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
)

type Log interface {
	Ready(dependencyName string)
	NotReady(dependencyName string, isOptional bool, err error)
}

type LogImp struct {
	l logger.Logger
}

func (l *LogImp) Ready(dependencyName string) {
	l.l.Info("Service is ready", zap.String("service", dependencyName))
}

func (l *LogImp) NotReady(dependencyName string, isOptional bool, err error) {
	msg := "Service is not ready"
	if isOptional {
		msg = "Service is not ready, but optional. Skipped."
	}

	l.l.Warn(msg, zap.String("service", dependencyName), zap.Error(err))
}

func Logger(l logger.Logger) *LogImp {
	return &LogImp{l: l}
}

var _ Log = &LogImp{}
