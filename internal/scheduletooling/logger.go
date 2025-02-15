package scheduletooling

import (
	"fmt"

	quartzlogger "github.com/reugn/go-quartz/logger"

	"github.com/yanakipre/bot/internal/logger"
)

type quartzLog struct {
	l logger.Logger
}

func (l quartzLog) Trace(msg any) {
}

func (l quartzLog) Tracef(format string, args ...any) {
}

func (l quartzLog) Debug(msg any) {
	l.l.Debug(fmt.Sprintf("%s", msg))
}

func (l quartzLog) Debugf(format string, args ...any) {
	l.l.Debug(fmt.Sprintf(format, args...))
}

func (l quartzLog) Info(msg any) {
	l.l.Info(fmt.Sprintf("%s", msg))
}

func (l quartzLog) Infof(format string, args ...any) {
	l.l.Info(fmt.Sprintf(format, args...))
}

func (l quartzLog) Warn(msg any) {
	l.l.Warn(fmt.Sprintf("%s", msg))
}

func (l quartzLog) Warnf(format string, args ...any) {
	l.l.Warn(fmt.Sprintf(format, args...))
}

func (l quartzLog) Error(msg any) {
	l.l.Error(fmt.Sprintf("%s", msg))
}

func (l quartzLog) Errorf(format string, args ...any) {
	l.l.Warn(fmt.Sprintf(format, args...))
}

func (l quartzLog) Enabled(level quartzlogger.Level) bool {
	return level >= quartzlogger.LevelError
}

var _ quartzlogger.Logger = &quartzLog{}
