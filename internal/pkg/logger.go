package pkg

import (
	"io"

	"github.com/sirupsen/logrus"
)

func DiscardLogger() *logrus.Logger {
	logger := logrus.New()
	logger.Out = io.Discard
	return logger
}
