package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()

	logger.Out = os.Stderr

	logger.Formatter = &logrus.TextFormatter{}

	log.SetLevel(log.InfoLevel)
}
