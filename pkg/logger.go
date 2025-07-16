package pkg

import (
	"io"
	"log"
	"log/slog"

	"github.com/seth-shi/go-v2ex/v2/model"
	"gopkg.in/natefinch/lumberjack.v2"
	"resty.dev/v3"
)

var (
	logPath = ".go-v2ex.log"
	logSize = 10 // MB
	logger  = slog.New(slog.NewTextHandler(io.Discard, nil))
)

func SetupLogger(conf *model.FileConfig) {

	var (
		w = io.Discard
	)
	if !conf.IsProductionEnv() {
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
	return &customLogger{l: logger}
}

// RestyLogger 实现 resty.Logger 接口，忽略所有日志输出
type customLogger struct {
	l *slog.Logger
}

func (l *customLogger) Errorf(format string, v ...interface{}) {
	l.l.Error(format, v)
}
func (l *customLogger) Warnf(format string, v ...interface{}) {
	l.l.Warn(format, v)
}
func (l *customLogger) Infof(format string, v ...interface{}) {
	l.l.Info(format, v)

}
func (l *customLogger) Debugf(format string, v ...interface{}) {
	l.l.Debug(format, v)
}
