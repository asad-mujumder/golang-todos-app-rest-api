package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(env string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		return zerolog.New(
			zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
		).With().Timestamp().Caller().Logger()
	}

	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}