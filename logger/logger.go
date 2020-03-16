package logger

import (
	"log"

	"go.uber.org/zap"
)

// Logger is the logger used by Pandor
var Logger *zap.Logger

//InitLogger builds Logger
func InitLogger() *zap.Logger {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	return l
}
