package log

import (
	"github.com/sirupsen/logrus"
	"os"
)

func InitLogger(logfile, severity string) *logrus.Logger {
	var log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)

	file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Out = os.Stdout
	}

	switch severity {
	case "info":
		log.Level = logrus.InfoLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "debug":
		log.Level = logrus.DebugLevel
	}
	return log
}