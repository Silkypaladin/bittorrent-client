package logger

import (
	"github.com/phuslu/log"
)

type Logger struct {
	log.Logger
}

func New() Logger {
	return Logger{
		log.Logger{
			Level:      log.InfoLevel,
			Caller:     1,
			TimeFormat: "2006-01-02 15:04:05",
			Writer: &log.ConsoleWriter{
				EndWithMessage: true,
				QuoteString:    true,
			},
		},
	}
}
