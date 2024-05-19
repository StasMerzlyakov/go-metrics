package domain

import "go.uber.org/zap"

var mainLogger *zap.SugaredLogger

func GetMainLogger() *zap.SugaredLogger {
	return mainLogger
}

func SetMainLogger(logger *zap.SugaredLogger) {
	mainLogger = logger
}
