package utils

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// Global logger instance
var Logger = logrus.New()

func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
}
