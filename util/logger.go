package util

import log "github.com/sirupsen/logrus"

func NewLogger() *log.Logger {
	logger := log.New()

	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	return logger
}
