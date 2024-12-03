package utils

import (
	"time"

	"github.com/rs/zerolog"
)

func Timer(name string, log zerolog.Logger) func() {
	start := time.Now()
	return func() {
		log.Info().Msgf("%s took %v seconds", name, time.Since(start))
	}
}
