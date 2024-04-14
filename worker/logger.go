package worker

import (
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type customLoger struct{}

func NewLoger() asynq.Logger {
	return &customLoger{}
}

func (logger customLoger) Print(logLevel zerolog.Level, args ...interface{}) {
	log.WithLevel(logLevel).Msg(fmt.Sprint(args...))
}

func (logger customLoger) Debug(args ...interface{}) {
	logger.Print(zerolog.DebugLevel, args...)
}

func (logger customLoger) Info(args ...interface{}) {
	logger.Print(zerolog.InfoLevel, args...)
}

func (logger customLoger) Warn(args ...interface{}) {
	logger.Print(zerolog.WarnLevel, args...)
}

func (logger customLoger) Error(args ...interface{}) {
	logger.Print(zerolog.ErrorLevel, args...)
}

func (logger customLoger) Fatal(args ...interface{}) {
	logger.Print(zerolog.FatalLevel, args...)
}
