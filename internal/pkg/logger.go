package pkg

import (
	"io"
	"log"
	"log/slog"

	"gopkg.in/natefinch/lumberjack.v2"
	"resty.dev/v3"
)

var (
	logPath = ".go-v2ex.log"
	logSize = 10 // MB
	logger  = slog.New(slog.NewTextHandler(io.Discard, nil))
)

func SetupLogger(debug bool) {

	var (
		w = io.Discard
	)
	if debug {
		w = &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    logSize,
			MaxBackups: 1,
		}
		logger = slog.New(slog.NewTextHandler(w, nil))
	}

	log.SetOutput(w)
	slog.SetDefault(logger)
}

func RestyLogger() resty.Logger {
	return &discardLogger{l: logger}
}

// RestyLogger 实现 resty.Logger 接口，忽略所有日志输出
type discardLogger struct {
	l *slog.Logger
}

func (l *discardLogger) Errorf(format string, v ...interface{}) {
	l.l.Error(format, v)
}
func (l *discardLogger) Warnf(format string, v ...interface{}) {
	l.l.Warn(format, v)
}
func (l *discardLogger) Infof(format string, v ...interface{}) {
	l.l.Info(format, v)

}
func (l *discardLogger) Debugf(format string, v ...interface{}) {
	l.l.Debug(format, v)
}
